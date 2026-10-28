package rest

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// NewRouter ensambla el multiplexor HTTP con las rutas de la API v1, estáticos de la web y la cadena de middlewares.
func NewRouter(h *Handler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	// Probes y diagnóstico
	mux.HandleFunc("GET /api/v1/healthz", h.Healthz)

	// Catálogo y cotizaciones
	mux.HandleFunc("GET /api/v1/categorias", h.ListCategorias)
	mux.HandleFunc("GET /api/v1/cotizaciones/resumen", h.GetCotizacionesResumen)
	mux.HandleFunc("GET /api/v1/cotizaciones/historico/{categoria_id}", h.GetSerieHistorica)

	// Simulador y calculadora financiera de liquidaciones
	mux.HandleFunc("POST /api/v1/calculadora/neto", h.SimularLiquidacion)
	mux.HandleFunc("POST /api/v1/calculadora/comparar", h.CompararOfertas)

	// Servir archivos estáticos del Frontend Web SPA (si el directorio existe)
	webDir := "web"
	if _, err := os.Stat(webDir); err != nil {
		if _, err2 := os.Stat("../../web"); err2 == nil {
			webDir = "../../web"
		}
	}

	if _, err := os.Stat(webDir); err == nil {
		fileServer := http.FileServer(http.Dir(webDir))
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}

			// Si es la raíz o el archivo no existe, servir index.html
			targetPath := filepath.Join(webDir, filepath.Clean(r.URL.Path))
			info, err := os.Stat(targetPath)
			if err != nil || info.IsDir() {
				http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
				return
			}

			fileServer.ServeHTTP(w, r)
		})
	}

	// Cadena de middlewares: CORS -> Logging -> Panic Recovery -> Mux
	var handler http.Handler = mux
	handler = Logging(logger)(handler)
	handler = Recoverer(logger)(handler)
	handler = CORS(handler)

	return handler
}

