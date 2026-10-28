package domain_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/domain"
)

func TestCategoriaCanon_Validate(t *testing.T) {
	t.Parallel()

	minKg := 300
	maxKg := 400

	validCat := domain.CategoriaCanon{
		Codigo:        "NOVILLITO",
		NombreVisible: "Novillitos Consumo",
		TipoDestino:   domain.DestinoFaena,
		PesoMinimoKg:  &minKg,
		PesoMaximoKg:  &maxKg,
		Activo:        true,
	}

	if err := validCat.Validate(); err != nil {
		t.Fatalf("se esperaba categoría válida, obtenido error: %v", err)
	}

	if !validCat.MatchesWeight(350) {
		t.Errorf("peso 350 debería encuadrar en rango [300, 400]")
	}
	if validCat.MatchesWeight(250) {
		t.Errorf("peso 250 NO debería encuadrar en rango [300, 400]")
	}
	if validCat.MatchesWeight(450) {
		t.Errorf("peso 450 NO debería encuadrar en rango [300, 400]")
	}

	invalidWeight := validCat
	invMin := 450
	invMax := 400
	invalidWeight.PesoMinimoKg = &invMin
	invalidWeight.PesoMaximoKg = &invMax
	if err := invalidWeight.Validate(); err != domain.ErrRangoPesoInvalido {
		t.Errorf("se esperaba ErrRangoPesoInvalido, obtenido %v", err)
	}
}

func TestCotizacion_Spread(t *testing.T) {
	t.Parallel()

	min := decimal.NewFromInt(2800)
	max := decimal.NewFromInt(3000)
	prom := decimal.NewFromInt(2900)

	cot := domain.Cotizacion{
		Fuente:           domain.FuenteMAG,
		Fecha:            time.Now(),
		CategoriaCanonID: 1,
		PrecioMin:        &min,
		PrecioMax:        &max,
		PrecioPromedio:   prom,
		VolumenCabezas:   1500,
	}

	if err := cot.Validate(); err != nil {
		t.Fatalf("cotización válida falló: %v", err)
	}

	spread, ok := cot.Spread()
	if !ok {
		t.Fatalf("spread debería ser válido")
	}

	expectedSpread := decimal.NewFromInt(200)
	if !spread.Equal(expectedSpread) {
		t.Errorf("esperado spread %s, obtenido %s", expectedSpread, spread)
	}
}

func TestCriterioVariacionDiaria(t *testing.T) {
	t.Parallel()

	criterio := domain.CriterioVariacionDiaria{
		UmbralPorcentaje: decimal.NewFromInt(4), // 4%
	}

	precioAyer := decimal.NewFromInt(2500)
	precioHoySuba := decimal.NewFromInt(2625) // +5%
	precioHoyBajaLeve := decimal.NewFromInt(2450) // -2%

	cumple, pct := criterio.Evaluar(precioAyer, precioHoySuba)
	if !cumple {
		t.Errorf("debería cumplir alerta con +5%% de suba")
	}
	if !pct.Equal(decimal.NewFromInt(5)) {
		t.Errorf("esperado +5%%, obtenido %s", pct)
	}

	cumpleBaja, pctBaja := criterio.Evaluar(precioAyer, precioHoyBajaLeve)
	if cumpleBaja {
		t.Errorf("no debería cumplir alerta con solo -2%% de variación")
	}
	if !pctBaja.Equal(decimal.NewFromInt(-2)) {
		t.Errorf("esperado -2%%, obtenido %s", pctBaja)
	}
}

