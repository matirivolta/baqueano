package collector_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/collector"
	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/platform/database"
	"github.com/matirivolta/baqueano/internal/platform/resilience"
)

type mockScraper struct {
	name   string
	quotes []collector.RawQuote
	err    error
}

func (m *mockScraper) Name() string {
	return m.name
}

func (m *mockScraper) Scrape(ctx context.Context) ([]collector.RawQuote, error) {
	return m.quotes, m.err
}

type mockQuotationRepo struct {
	saved []domain.Cotizacion
}

func (m *mockQuotationRepo) GetUltimasCotizaciones(ctx context.Context) ([]domain.Cotizacion, error) {
	return m.saved, nil
}
func (m *mockQuotationRepo) GetSerieHistorica(ctx context.Context, categoriaID int32, desde time.Time) ([]domain.Cotizacion, error) {
	return m.saved, nil
}
func (m *mockQuotationRepo) GetCotizacionPrevia(ctx context.Context, categoriaID int32, fuente string, fecha time.Time) (*domain.Cotizacion, error) {
	return nil, nil
}
func (m *mockQuotationRepo) UpsertCotizacion(ctx context.Context, cot domain.Cotizacion) error {
	m.saved = append(m.saved, cot)
	return nil
}

type mockProducer struct {
	published []domain.CotizacionIngeridaEvent
}

func (m *mockProducer) PublishCotizacion(ctx context.Context, event domain.CotizacionIngeridaEvent) (string, error) {
	m.published = append(m.published, event)
	return "msg-123", nil
}

func TestEngine_ConcurrentIngestionWithFailureIsolation(t *testing.T) {
	t.Parallel()

	// 1. Configurar ACL con categoría conocida
	aclRepo := newMockACLRepo()
	catID := int32(2)
	aclRepo.mapeos["MAG:Novillitos"] = &database.MapeoFuenteResult{
		ID:                  1,
		Fuente:              "MAG",
		TextoOrigen:         "Novillitos",
		CategoriaCanonID:    &catID,
		VerificadoPorHumano: true,
	}

	normalizer := collector.NewNormalizer(aclRepo)
	quotRepo := &mockQuotationRepo{}
	producer := &mockProducer{}

	engine := collector.NewEngine(normalizer, quotRepo, producer, resilience.Config{
		MaxFailures: 3,
		Timeout:     time.Minute,
	})

	// Scraper 1: Exitoso (MAG)
	magScraper := &mockScraper{
		name: domain.FuenteMAG,
		quotes: []collector.RawQuote{
			{
				Fuente:         domain.FuenteMAG,
				Fecha:          time.Now(),
				TextoOrigen:    "Novillitos",
				PrecioPromedio: decimal.NewFromFloat(2950.00),
				VolumenCabezas: 1500,
			},
		},
	}

	// Scraper 2: Fallido (Rosgan con timeout simulado)
	rosganScraper := &mockScraper{
		name: domain.FuenteRosgan,
		err:  errors.New("timeout de conexión simulado"),
	}

	engine.RegisterScraper(magScraper)
	engine.RegisterScraper(rosganScraper)

	// Ejecutar motor
	cotizaciones, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("el motor no debería fallar completamente si una fuente responde: %v", err)
	}

	// Debe haber procesado exitosamente la cotización de MAG a pesar de la falla de Rosgan
	if len(cotizaciones) != 1 {
		t.Fatalf("se esperaba 1 cotización procesada, obtenidas %d", len(cotizaciones))
	}
	if cotizaciones[0].Fuente != domain.FuenteMAG {
		t.Errorf("esperada fuente MAG, obtenida %s", cotizaciones[0].Fuente)
	}

	// Verificar persistencia en repositorio
	if len(quotRepo.saved) != 1 {
		t.Errorf("se esperaba 1 cotización guardada en BD, obtenidas %d", len(quotRepo.saved))
	}

	// Verificar emisión en Event Bus
	if len(producer.published) != 1 {
		t.Errorf("se esperaba 1 evento publicado a Redis Streams, obtenidos %d", len(producer.published))
	}
}

