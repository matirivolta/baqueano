package quotation

import (
	"context"
	"time"

	"github.com/matirivolta/baqueano/internal/domain"
)

// QuotationRepository define el contrato para acceder a cotizaciones históricas y vigentes.
type QuotationRepository interface {
	GetUltimasCotizaciones(ctx context.Context) ([]domain.Cotizacion, error)
	GetSerieHistorica(ctx context.Context, categoriaID int32, desde time.Time) ([]domain.Cotizacion, error)
	GetCotizacionPrevia(ctx context.Context, categoriaID int32, fuente string, fecha time.Time) (*domain.Cotizacion, error)
	UpsertCotizacion(ctx context.Context, cot domain.Cotizacion) error
}

// CategoriaRepository define el contrato para consultar las categorías canónicas de hacienda.
type CategoriaRepository interface {
	ListCategorias(ctx context.Context) ([]domain.CategoriaCanon, error)
	GetCategoriaByCodigo(ctx context.Context, codigo string) (domain.CategoriaCanon, error)
	GetCategoriaByID(ctx context.Context, id int32) (domain.CategoriaCanon, error)
}

// Calculator define las operaciones matemáticas y financieras de liquidación neta.
type Calculator interface {
	CalcularLiquidacion(params domain.ParametrosLiquidacion) (domain.LiquidacionNeta, error)
	CompararOfertas(ofertas []domain.ParametrosLiquidacion) ([]domain.LiquidacionNeta, error)
}

// Service expone los casos de uso del Bounded Context de Cotizaciones y Valuación.
type Service interface {
	ObtenerResumenMercado(ctx context.Context) ([]domain.Cotizacion, error)
	ObtenerSerieHistorica(ctx context.Context, categoriaID int32, desde time.Time) ([]domain.Cotizacion, error)
	SimularLiquidacion(params domain.ParametrosLiquidacion) (domain.LiquidacionNeta, error)
	CompararOfertas(ofertas []domain.ParametrosLiquidacion) ([]domain.LiquidacionNeta, error)
}

