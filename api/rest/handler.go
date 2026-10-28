package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/quotation"
)

// Pinger define una abstracción para verificar conectividad con servicios dependientes.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler gestiona las peticiones HTTP y orquesta los casos de uso del dominio.
type Handler struct {
	svc     quotation.Service
	catRepo quotation.CategoriaRepository
	dbPing  Pinger
}

// NewHandler construye una nueva instancia del Handler HTTP inyectando sus dependencias.
func NewHandler(svc quotation.Service, catRepo quotation.CategoriaRepository, dbPing Pinger) *Handler {
	return &Handler{
		svc:     svc,
		catRepo: catRepo,
		dbPing:  dbPing,
	}
}

// HealthResponse estructura la respuesta del probe de salud.
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// Healthz responde con el estado de liveness y conectividad a base de datos.
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	services := make(map[string]string)
	status := "ok"

	if h.dbPing != nil {
		if err := h.dbPing.Ping(ctx); err != nil {
			services["postgres"] = "unhealthy: " + err.Error()
			status = "degraded"
		} else {
			services["postgres"] = "connected"
		}
	} else {
		services["postgres"] = "not_configured"
	}

	statusCode := http.StatusOK
	if status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	respondJSON(w, statusCode, HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Services:  services,
	})
}

// ListCategorias expone el catálogo canónico de categorías de hacienda.
func (h *Handler) ListCategorias(w http.ResponseWriter, r *http.Request) {
	if h.catRepo == nil {
		respondJSON(w, http.StatusOK, []domain.CategoriaCanon{})
		return
	}
	categorias, err := h.catRepo.ListCategorias(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error consultando categorías: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, categorias)
}

// ResumenResponse estructura la respuesta del tablero consolidado junto al ratio de reposición.
type ResumenResponse struct {
	Cotizaciones      []domain.Cotizacion  `json:"cotizaciones"`
	RelacionReposicion *RelacionReposicion `json:"relacion_reposicion,omitempty"`
}

// RelacionReposicion computa el ratio Ternero/Novillo clave para el ciclo productivo argentino.
type RelacionReposicion struct {
	PrecioTernero decimal.Decimal `json:"precio_ternero"`
	PrecioNovillo decimal.Decimal `json:"precio_novillo"`
	Ratio         decimal.Decimal `json:"ratio"`
	Interpretacion string         `json:"interpretacion"`
}

// GetCotizacionesResumen devuelve el último registro de precios por plaza y categoría, y el ratio ternero/novillo.
func (h *Handler) GetCotizacionesResumen(w http.ResponseWriter, r *http.Request) {
	cotizaciones, err := h.svc.ObtenerResumenMercado(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error obteniendo cotizaciones: "+err.Error())
		return
	}

	res := ResumenResponse{
		Cotizaciones: cotizaciones,
	}

	// Buscar cotización de Ternero de Invernada y Novillo para calcular la relación Ternero/Novillo
	var precioTernero, precioNovillo *decimal.Decimal
	for _, c := range cotizaciones {
		cod := strings.ToLower(c.CategoriaCodigo)
		nom := strings.ToLower(c.CategoriaNombre)
		if (strings.Contains(cod, "ternero") || strings.Contains(nom, "ternero")) && precioTernero == nil {
			val := c.PrecioPromedio
			precioTernero = &val
		}
		if (strings.Contains(cod, "novillo") || strings.Contains(nom, "novillo")) && !strings.Contains(cod, "novillito") && precioNovillo == nil {
			val := c.PrecioPromedio
			precioNovillo = &val
		}
	}

	if precioTernero != nil && precioNovillo != nil && !precioNovillo.IsZero() {
		ratio := precioTernero.DivRound(*precioNovillo, 3)
		var interpretacion string
		// En la ganadería argentina:
		// Ratio normal histórico ~1.15 a 1.25.
		// Si ratio > 1.25 -> El ternero está caro respecto al novillo (favorable a cría / costoso reponer feedlot).
		// Si ratio < 1.15 -> Favorable para la invernada y compra de terneros para engorde.
		if ratio.GreaterThan(decimal.NewFromFloat(1.25)) {
			interpretacion = "Relación favorable para la cría. Reposición de invernada costosa."
		} else if ratio.LessThan(decimal.NewFromFloat(1.15)) {
			interpretacion = "Relación favorable para la invernada/feedlot. Oportunidad de compra de terneros."
		} else {
			interpretacion = "Relación en equilibrio histórico promedio (1.15 - 1.25)."
		}

		res.RelacionReposicion = &RelacionReposicion{
			PrecioTernero: *precioTernero,
			PrecioNovillo: *precioNovillo,
			Ratio:         ratio,
			Interpretacion: interpretacion,
		}
	}

	respondJSON(w, http.StatusOK, res)
}

// GetSerieHistorica retorna la evolución cronológica de cotizaciones para una categoría.
func (h *Handler) GetSerieHistorica(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("categoria_id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "categoria_id debe ser un entero positivo")
		return
	}

	desde := time.Now().AddDate(0, -1, 0) // default últimos 30 días
	if desdeStr := r.URL.Query().Get("desde"); desdeStr != "" {
		if t, parseErr := time.Parse("2006-01-02", desdeStr); parseErr == nil {
			desde = t
		}
	}

	serie, err := h.svc.ObtenerSerieHistorica(r.Context(), int32(id), desde)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error obteniendo serie histórica: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, serie)
}

// SimularLiquidacion ejecuta el motor de cálculo financiero para liquidación neta al bolsillo.
func (h *Handler) SimularLiquidacion(w http.ResponseWriter, r *http.Request) {
	var params domain.ParametrosLiquidacion
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de solicitud JSON inválido: "+err.Error())
		return
	}

	if err := params.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, "parámetros inválidos: "+err.Error())
		return
	}

	res, err := h.svc.SimularLiquidacion(params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error en cálculo de liquidación: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, res)
}

// CompararOfertas evalúa múltiples propuestas de compra o venta de hacienda.
func (h *Handler) CompararOfertas(w http.ResponseWriter, r *http.Request) {
	var ofertas []domain.ParametrosLiquidacion
	if err := json.NewDecoder(r.Body).Decode(&ofertas); err != nil {
		respondError(w, http.StatusBadRequest, "cuerpo de solicitud JSON inválido: "+err.Error())
		return
	}

	if len(ofertas) == 0 {
		respondError(w, http.StatusBadRequest, "debe proveer al menos una oferta para comparar")
		return
	}

	for i, of := range ofertas {
		if err := of.Validate(); err != nil {
			respondError(w, http.StatusBadRequest, "oferta en índice "+strconv.Itoa(i)+" inválida: "+err.Error())
			return
		}
	}

	resultados, err := h.svc.CompararOfertas(ofertas)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "error comparando ofertas: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resultados)
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
