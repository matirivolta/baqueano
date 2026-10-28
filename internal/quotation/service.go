package quotation

import (
	"context"
	"errors"
	"time"

	"github.com/matirivolta/baqueano/internal/domain"
)

var ErrRepoNoDisponible = errors.New("repositorio de datos no disponible")

// DefaultService implementa los casos de uso de negocio del Bounded Context de Cotizaciones.
type DefaultService struct {
	repo       QuotationRepository
	catRepo    CategoriaRepository
	calculator Calculator
}

// NewService construye un nuevo servicio inyectando sus dependencias desacopladas mediante interfaces.
func NewService(repo QuotationRepository, catRepo CategoriaRepository, calc Calculator) *DefaultService {
	if calc == nil {
		calc = NewCalculator()
	}
	return &DefaultService{
		repo:       repo,
		catRepo:    catRepo,
		calculator: calc,
	}
}

// ObtenerResumenMercado obtiene el tablero consolidado de últimas cotizaciones por categoría y mercado.
func (s *DefaultService) ObtenerResumenMercado(ctx context.Context) ([]domain.Cotizacion, error) {
	if s.repo == nil {
		return nil, ErrRepoNoDisponible
	}
	return s.repo.GetUltimasCotizaciones(ctx)
}

// ObtenerSerieHistorica retorna la serie temporal de precios para análisis gráfico o de dispersión.
func (s *DefaultService) ObtenerSerieHistorica(ctx context.Context, categoriaID int32, desde time.Time) ([]domain.Cotizacion, error) {
	if s.repo == nil {
		return nil, ErrRepoNoDisponible
	}
	return s.repo.GetSerieHistorica(ctx, categoriaID, desde)
}

// SimularLiquidacion ejecuta el cálculo financiero de cobro neto según los parámetros suministrados.
func (s *DefaultService) SimularLiquidacion(params domain.ParametrosLiquidacion) (domain.LiquidacionNeta, error) {
	return s.calculator.CalcularLiquidacion(params)
}

// CompararOfertas evalúa un conjunto de ofertas y computa sus liquidaciones para decisión comercial.
func (s *DefaultService) CompararOfertas(ofertas []domain.ParametrosLiquidacion) ([]domain.LiquidacionNeta, error) {
	return s.calculator.CompararOfertas(ofertas)
}

