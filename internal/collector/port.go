package collector

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/platform/database"
)

// RawQuote representa la información de cotización extraída sin procesar de una fuente web.
type RawQuote struct {
	Fuente         string                 `json:"fuente"`
	Fecha          time.Time              `json:"fecha"`
	TextoOrigen    string                 `json:"texto_origen"`
	PrecioMin      *decimal.Decimal       `json:"precio_min,omitempty"`
	PrecioMax      *decimal.Decimal       `json:"precio_max,omitempty"`
	PrecioPromedio decimal.Decimal        `json:"precio_promedio"`
	VolumenCabezas int                    `json:"volumen_cabezas"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// SourceScraper define la interfaz polimórfica que debe implementar cada estrategia de mercado.
type SourceScraper interface {
	Name() string
	Scrape(ctx context.Context) ([]RawQuote, error)
}

// ACLRepository define el contrato que requiere el normalizador para resolver y registrar mapeos.
type ACLRepository interface {
	GetMapeo(ctx context.Context, fuente, textoOrigen string) (*database.MapeoFuenteResult, error)
	UpsertMapeo(ctx context.Context, fuente, textoOrigen string, categoriaID *int32, verificado bool) error
}

// Normalizer define el contrato del Anti-Corruption Layer para transformar cotizaciones crudas en canónicas.
type Normalizer interface {
	Normalize(ctx context.Context, raw RawQuote) (*domain.Cotizacion, error)
}

// Engine orquesta concurrentemente todos los scrapers registrados aplicando Circuit Breakers.
type Engine interface {
	RegisterScraper(scraper SourceScraper)
	Run(ctx context.Context) ([]domain.Cotizacion, error)
}

