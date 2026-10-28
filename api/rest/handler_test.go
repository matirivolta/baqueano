package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/quotation"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	return m.err
}

type mockService struct {
	resumen        []domain.Cotizacion
	resumenErr     error
	serie          []domain.Cotizacion
	serieErr       error
	calc           quotation.Calculator
}

func newMockService() *mockService {
	return &mockService{
		calc: quotation.NewCalculator(),
	}
}

func (m *mockService) ObtenerResumenMercado(ctx context.Context) ([]domain.Cotizacion, error) {
	return m.resumen, m.resumenErr
}

func (m *mockService) ObtenerSerieHistorica(ctx context.Context, categoriaID int32, desde time.Time) ([]domain.Cotizacion, error) {
	return m.serie, m.serieErr
}

func (m *mockService) SimularLiquidacion(params domain.ParametrosLiquidacion) (domain.LiquidacionNeta, error) {
	return m.calc.CalcularLiquidacion(params)
}

func (m *mockService) CompararOfertas(ofertas []domain.ParametrosLiquidacion) ([]domain.LiquidacionNeta, error) {
	return m.calc.CompararOfertas(ofertas)
}

type mockCatRepo struct {
	cats []domain.CategoriaCanon
	err  error
}

func (m *mockCatRepo) ListCategorias(ctx context.Context) ([]domain.CategoriaCanon, error) {
	return m.cats, m.err
}

func (m *mockCatRepo) GetCategoriaByCodigo(ctx context.Context, codigo string) (domain.CategoriaCanon, error) {
	return domain.CategoriaCanon{}, errors.New("not implemented")
}

func (m *mockCatRepo) GetCategoriaByID(ctx context.Context, id int32) (domain.CategoriaCanon, error) {
	return domain.CategoriaCanon{}, errors.New("not implemented")
}

func setupTestServer(svc quotation.Service, catRepo quotation.CategoriaRepository, pinger Pinger) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewHandler(svc, catRepo, pinger)
	return NewRouter(h, logger)
}

func TestHealthz(t *testing.T) {
	t.Run("Healthz OK", func(t *testing.T) {
		router := setupTestServer(newMockService(), &mockCatRepo{}, &mockPinger{err: nil})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("se esperaba status 200, obtenido %d", rec.Code)
		}

		var res HealthResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("error decodificando json: %v", err)
		}
		if res.Status != "ok" {
			t.Errorf("status esperado 'ok', obtenido '%s'", res.Status)
		}
		if res.Services["postgres"] != "connected" {
			t.Errorf("postgres esperado 'connected', obtenido '%s'", res.Services["postgres"])
		}
	})

	t.Run("Healthz Degraded", func(t *testing.T) {
		router := setupTestServer(newMockService(), &mockCatRepo{}, &mockPinger{err: errors.New("conn refused")})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("se esperaba status 503, obtenido %d", rec.Code)
		}
	})
}

func TestListCategorias(t *testing.T) {
	catRepo := &mockCatRepo{
		cats: []domain.CategoriaCanon{
			{ID: 1, Codigo: "NOVILLO_ESP", NombreVisible: "Novillo Especial", TipoDestino: domain.DestinoFaena, Activo: true},
			{ID: 2, Codigo: "TERNERO_INVERNADA", NombreVisible: "Ternero Invernada", TipoDestino: domain.DestinoInvernada, Activo: true},
		},
	}
	router := setupTestServer(newMockService(), catRepo, &mockPinger{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categorias", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, obtenido %d", rec.Code)
	}

	var cats []domain.CategoriaCanon
	if err := json.NewDecoder(rec.Body).Decode(&cats); err != nil {
		t.Fatalf("error decodificando json: %v", err)
	}
	if len(cats) != 2 {
		t.Fatalf("se esperaban 2 categorías, obtenidas %d", len(cats))
	}
}

func TestGetCotizacionesResumen(t *testing.T) {
	svc := newMockService()
	svc.resumen = []domain.Cotizacion{
		{
			CategoriaCodigo: "TERNERO_INVERNADA",
			CategoriaNombre: "Terneros 160-180kg",
			PrecioPromedio:  decimal.NewFromInt(2500),
			Fuente:          "ROSGAN",
		},
		{
			CategoriaCodigo: "NOVILLO_ESP",
			CategoriaNombre: "Novillos Consumo",
			PrecioPromedio:  decimal.NewFromInt(2000),
			Fuente:          "MAG",
		},
	}
	router := setupTestServer(svc, &mockCatRepo{}, &mockPinger{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cotizaciones/resumen", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, obtenido %d", rec.Code)
	}

	var res ResumenResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("error decodificando json: %v", err)
	}
	if len(res.Cotizaciones) != 2 {
		t.Fatalf("se esperaban 2 cotizaciones, obtenidas %d", len(res.Cotizaciones))
	}
	if res.RelacionReposicion == nil {
		t.Fatalf("se esperaba cálculo de relación de reposición")
	}

	// 2500 / 2000 = 1.25
	expectedRatio := decimal.NewFromFloat(1.25)
	if !res.RelacionReposicion.Ratio.Equal(expectedRatio) {
		t.Errorf("ratio esperado %s, obtenido %s", expectedRatio, res.RelacionReposicion.Ratio)
	}
}

func TestGetSerieHistorica(t *testing.T) {
	svc := newMockService()
	svc.serie = []domain.Cotizacion{
		{Fecha: time.Now().AddDate(0, 0, -5), PrecioPromedio: decimal.NewFromInt(1950), Fuente: "MAG"},
		{Fecha: time.Now(), PrecioPromedio: decimal.NewFromInt(2000), Fuente: "MAG"},
	}
	router := setupTestServer(svc, &mockCatRepo{}, &mockPinger{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cotizaciones/historico/1?desde=2026-10-01", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, obtenido %d", rec.Code)
	}

	var serie []domain.Cotizacion
	if err := json.NewDecoder(rec.Body).Decode(&serie); err != nil {
		t.Fatalf("error decodificando json: %v", err)
	}
	if len(serie) != 2 {
		t.Fatalf("serie esperada de 2 elementos, obtenido %d", len(serie))
	}
}

func TestSimularLiquidacion(t *testing.T) {
	svc := newMockService()
	router := setupTestServer(svc, &mockCatRepo{}, &mockPinger{})

	t.Run("Cálculo exitoso de liquidación neta", func(t *testing.T) {
		payload := domain.ParametrosLiquidacion{
			Cabezas:               50,
			PesoBrutoKg:           decimal.NewFromInt(400),
			EsPesoIndividual:      true,
			PrecioPorKg:           decimal.NewFromInt(2000),
			PorcentajeDesbaste:    decimal.NewFromInt(3),
			PorcentajeComision:    decimal.NewFromInt(2),
			PorcentajeIVAComision: decimal.NewFromInt(21),
			CostoFleteTotal:       decimal.NewFromInt(300000),
			OtrosCostosFijos:      decimal.NewFromInt(50000),
			PlazoDias:             30,
			TasaInteresAnual:      decimal.NewFromInt(45),
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/calculadora/neto", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status esperado 200, obtenido %d: %s", rec.Code, rec.Body.String())
		}

		var liq domain.LiquidacionNeta
		if err := json.NewDecoder(rec.Body).Decode(&liq); err != nil {
			t.Fatalf("error decodificando json: %v", err)
		}

		if liq.CobroNetoEfectivo.LessThanOrEqual(decimal.Zero) {
			t.Errorf("cobro neto esperado > 0, obtenido %s", liq.CobroNetoEfectivo)
		}
		if liq.PrecioNetoPorKiloNeto.LessThanOrEqual(decimal.Zero) {
			t.Errorf("precio neto por kg esperado > 0, obtenido %s", liq.PrecioNetoPorKiloNeto)
		}
	})

	t.Run("Error de validación por cabezas negativas", func(t *testing.T) {
		payload := domain.ParametrosLiquidacion{
			Cabezas:     0,
			PesoBrutoKg: decimal.NewFromInt(400),
			PrecioPorKg: decimal.NewFromInt(2000),
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/calculadora/neto", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status esperado 400, obtenido %d", rec.Code)
		}
	})
}

func TestCORSPreflight(t *testing.T) {
	router := setupTestServer(newMockService(), &mockCatRepo{}, &mockPinger{})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/cotizaciones/resumen", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status esperado 204 para OPTIONS preflight, obtenido %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("header CORS esperado '*', obtenido '%s'", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestStaticFiles(t *testing.T) {
	router := setupTestServer(newMockService(), &mockCatRepo{}, &mockPinger{})

	t.Run("Servir index.html en la raíz", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status esperado 200 para '/', obtenido %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "BAQUEANO") {
			t.Errorf("se esperaba que index.html contenga 'BAQUEANO'")
		}
	})

	t.Run("Servir archivo CSS estático", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status esperado 200 para '/static/css/style.css', obtenido %d", rec.Code)
		}
	})
}


