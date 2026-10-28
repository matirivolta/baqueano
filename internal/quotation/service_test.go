package quotation_test

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/quotation"
)

type mockQuotationRepo struct {
	cotizaciones []domain.Cotizacion
	err          error
}

func (m *mockQuotationRepo) GetUltimasCotizaciones(ctx context.Context) ([]domain.Cotizacion, error) {
	return m.cotizaciones, m.err
}

func (m *mockQuotationRepo) GetSerieHistorica(ctx context.Context, categoriaID int32, desde time.Time) ([]domain.Cotizacion, error) {
	return m.cotizaciones, m.err
}

func (m *mockQuotationRepo) GetCotizacionPrevia(ctx context.Context, categoriaID int32, fuente string, fecha time.Time) (*domain.Cotizacion, error) {
	if len(m.cotizaciones) > 0 {
		return &m.cotizaciones[0], m.err
	}
	return nil, m.err
}

func (m *mockQuotationRepo) UpsertCotizacion(ctx context.Context, cot domain.Cotizacion) error {
	m.cotizaciones = append(m.cotizaciones, cot)
	return m.err
}

type mockCategoriaRepo struct {
	categorias []domain.CategoriaCanon
	err        error
}

func (m *mockCategoriaRepo) ListCategorias(ctx context.Context) ([]domain.CategoriaCanon, error) {
	return m.categorias, m.err
}

func (m *mockCategoriaRepo) GetCategoriaByCodigo(ctx context.Context, codigo string) (domain.CategoriaCanon, error) {
	for _, c := range m.categorias {
		if c.Codigo == codigo {
			return c, nil
		}
	}
	return domain.CategoriaCanon{}, m.err
}

func (m *mockCategoriaRepo) GetCategoriaByID(ctx context.Context, id int32) (domain.CategoriaCanon, error) {
	for _, c := range m.categorias {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.CategoriaCanon{}, m.err
}

func TestService_ObtenerResumenMercado(t *testing.T) {
	t.Parallel()

	precio := decimal.NewFromFloat(2900.50)
	cots := []domain.Cotizacion{
		{
			ID:               1,
			Fecha:            time.Now(),
			Fuente:           domain.FuenteMAG,
			CategoriaCanonID: 1,
			CategoriaCodigo:  "NOVILLITO",
			PrecioPromedio:   precio,
			VolumenCabezas:   1200,
		},
	}

	repo := &mockQuotationRepo{cotizaciones: cots}
	catRepo := &mockCategoriaRepo{}
	svc := quotation.NewService(repo, catRepo, nil)

	res, err := svc.ObtenerResumenMercado(context.Background())
	if err != nil {
		t.Fatalf("error inesperado en ObtenerResumenMercado: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("esperado 1 elemento, obtenido %d", len(res))
	}

	if res[0].CategoriaCodigo != "NOVILLITO" {
		t.Errorf("esperado NOVILLITO, obtenido %s", res[0].CategoriaCodigo)
	}
}

func TestService_SimularLiquidacion(t *testing.T) {
	t.Parallel()

	svc := quotation.NewService(&mockQuotationRepo{}, &mockCategoriaRepo{}, nil)

	params := domain.ParametrosLiquidacion{
		Cabezas:               20,
		PesoBrutoKg:           decimal.NewFromInt(300),
		EsPesoIndividual:      true,
		PrecioPorKg:           decimal.NewFromInt(2800),
		PorcentajeDesbaste:    decimal.NewFromInt(3),
		PorcentajeComision:    decimal.NewFromInt(2),
		PorcentajeIVAComision: decimal.NewFromInt(21),
	}

	liq, err := svc.SimularLiquidacion(params)
	if err != nil {
		t.Fatalf("error simulando liquidación: %v", err)
	}

	if liq.KilosBrutos.LessThanOrEqual(decimal.Zero) {
		t.Errorf("esperado kilos brutos > 0, obtenido %s", liq.KilosBrutos)
	}
	if liq.CobroNetoEfectivo.LessThanOrEqual(decimal.Zero) {
		t.Errorf("esperado cobro neto > 0, obtenido %s", liq.CobroNetoEfectivo)
	}
}

