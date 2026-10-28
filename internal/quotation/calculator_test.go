package quotation_test

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/quotation"
)

func TestCalcularLiquidacion_VentaDirectaCampo(t *testing.T) {
	t.Parallel()

	calc := quotation.NewCalculator()

	// Escenario: 40 novillitos de 350 kg brutos, vendidos en campo con 3% de desbaste,
	// $2.900/kg pactado, 2% de comisión (+ 21% IVA), sin flete (paga comprador), contado.
	params := domain.ParametrosLiquidacion{
		Cabezas:               40,
		PesoBrutoKg:           decimal.NewFromInt(350),
		EsPesoIndividual:      true,
		PrecioPorKg:           decimal.NewFromInt(2900),
		PorcentajeDesbaste:    decimal.NewFromInt(3), // 3%
		PorcentajeComision:    decimal.NewFromInt(2), // 2%
		PorcentajeIVAComision: decimal.NewFromInt(21),
		CostoFleteTotal:       decimal.Zero,
		OtrosCostosFijos:      decimal.Zero,
		PlazoDias:             0,
		TasaInteresAnual:      decimal.Zero,
	}

	liq, err := calc.CalcularLiquidacion(params)
	if err != nil {
		t.Fatalf("error inesperado calculando liquidación: %v", err)
	}

	// Kilos brutos = 40 * 350 = 14.000 kg
	expectedKilosBrutos := decimal.NewFromInt(14000)
	if !liq.KilosBrutos.Equal(expectedKilosBrutos) {
		t.Errorf("KilosBrutos: esperado %s, obtenido %s", expectedKilosBrutos, liq.KilosBrutos)
	}

	// Kilos desbaste = 14.000 * 0.03 = 420 kg
	expectedDesbaste := decimal.NewFromInt(420)
	if !liq.KilosDesbaste.Equal(expectedDesbaste) {
		t.Errorf("KilosDesbaste: esperado %s, obtenido %s", expectedDesbaste, liq.KilosDesbaste)
	}

	// Kilos netos = 14.000 - 420 = 13.580 kg
	expectedKilosNetos := decimal.NewFromInt(13580)
	if !liq.KilosNetos.Equal(expectedKilosNetos) {
		t.Errorf("KilosNetos: esperado %s, obtenido %s", expectedKilosNetos, liq.KilosNetos)
	}

	// Importe Bruto = 13.580 * 2.900 = $39.382.000,00
	expectedBruto := decimal.NewFromInt(39382000)
	if !liq.ImporteBruto.Equal(expectedBruto) {
		t.Errorf("ImporteBruto: esperado %s, obtenido %s", expectedBruto, liq.ImporteBruto)
	}

	// Comisión = 39.382.000 * 0.02 = $787.640,00
	expectedComision := decimal.NewFromInt(787640)
	if !liq.MontoComision.Equal(expectedComision) {
		t.Errorf("MontoComision: esperado %s, obtenido %s", expectedComision, liq.MontoComision)
	}

	// IVA Comisión (21%) = 787.640 * 0.21 = $165.404,40
	expectedIVA := decimal.NewFromFloat(165404.40)
	if !liq.MontoIVAComision.Equal(expectedIVA) {
		t.Errorf("MontoIVAComision: esperado %s, obtenido %s", expectedIVA, liq.MontoIVAComision)
	}

	// Cobro neto = 39.382.000 - 787.640 - 165.404,40 = $38.428.955,60
	expectedNeto := decimal.NewFromFloat(38428955.60)
	if !liq.CobroNetoEfectivo.Equal(expectedNeto) {
		t.Errorf("CobroNetoEfectivo: esperado %s, obtenido %s", expectedNeto, liq.CobroNetoEfectivo)
	}

	// Precio neto por kilo neto = 38.428.955,60 / 13.580 = $2.829,82
	expectedPrecioKiloNeto := decimal.NewFromFloat(2829.82)
	if !liq.PrecioNetoPorKiloNeto.Equal(expectedPrecioKiloNeto) {
		t.Errorf("PrecioNetoPorKiloNeto: esperado %s, obtenido %s", expectedPrecioKiloNeto, liq.PrecioNetoPorKiloNeto)
	}
}

func TestCalcularLiquidacion_MercadoCañuelasConPlazo(t *testing.T) {
	t.Parallel()

	calc := quotation.NewCalculator()

	// Escenario: Tropa pesada en destino (14.700 kg total para 35 cabezas), 0% desbaste,
	// $3.050/kg, 3% comisión (+21% IVA), flete $450.000, gastos $65.000, 14 días plazo al 40% TNA.
	params := domain.ParametrosLiquidacion{
		Cabezas:               35,
		PesoBrutoKg:           decimal.NewFromInt(14700),
		EsPesoIndividual:      false,
		PrecioPorKg:           decimal.NewFromInt(3050),
		PorcentajeDesbaste:    decimal.Zero,
		PorcentajeComision:    decimal.NewFromInt(3),
		PorcentajeIVAComision: decimal.NewFromInt(21),
		CostoFleteTotal:       decimal.NewFromInt(450000),
		OtrosCostosFijos:      decimal.NewFromInt(65000),
		PlazoDias:             14,
		TasaInteresAnual:      decimal.NewFromInt(40), // 40% TNA
	}

	liq, err := calc.CalcularLiquidacion(params)
	if err != nil {
		t.Fatalf("error inesperado calculando liquidación en Cañuelas: %v", err)
	}

	// Importe Bruto = 14.700 * 3.050 = $44.835.000,00
	expectedBruto := decimal.NewFromInt(44835000)
	if !liq.ImporteBruto.Equal(expectedBruto) {
		t.Errorf("ImporteBruto: esperado %s, obtenido %s", expectedBruto, liq.ImporteBruto)
	}

	// Costo Financiero = 42.692.489,50 * (0.40 * 14 / 365) = $655.008,06
	expectedCostoFinanciero := decimal.NewFromFloat(655008.06)
	if !liq.CostoFinanciero.Equal(expectedCostoFinanciero) {
		t.Errorf("CostoFinanciero: esperado %s, obtenido %s", expectedCostoFinanciero, liq.CostoFinanciero)
	}

	// Cobro neto efectivo = 42.692.489,50 - 655.008,06 = $42.037.481,44
	expectedNeto := decimal.NewFromFloat(42037481.44)
	if !liq.CobroNetoEfectivo.Equal(expectedNeto) {
		t.Errorf("CobroNetoEfectivo: esperado %s, obtenido %s", expectedNeto, liq.CobroNetoEfectivo)
	}
}

func TestCalcularLiquidacion_Validaciones(t *testing.T) {
	t.Parallel()

	calc := quotation.NewCalculator()

	baseValid := domain.ParametrosLiquidacion{
		Cabezas:               10,
		PesoBrutoKg:           decimal.NewFromInt(300),
		EsPesoIndividual:      true,
		PrecioPorKg:           decimal.NewFromInt(2500),
		PorcentajeDesbaste:    decimal.NewFromInt(3),
		PorcentajeComision:    decimal.NewFromInt(2),
		PorcentajeIVAComision: decimal.NewFromInt(21),
	}

	tests := []struct {
		name        string
		mutator     func(p *domain.ParametrosLiquidacion)
		expectedErr error
	}{
		{
			name: "Cabezas cero",
			mutator: func(p *domain.ParametrosLiquidacion) {
				p.Cabezas = 0
			},
			expectedErr: domain.ErrCabezasInvalidas,
		},
		{
			name: "Peso negativo",
			mutator: func(p *domain.ParametrosLiquidacion) {
				p.PesoBrutoKg = decimal.NewFromInt(-10)
			},
			expectedErr: domain.ErrPesoInvalido,
		},
		{
			name: "Desbaste 100% o superior",
			mutator: func(p *domain.ParametrosLiquidacion) {
				p.PorcentajeDesbaste = decimal.NewFromInt(100)
			},
			expectedErr: domain.ErrDesbasteInvalido,
		},
		{
			name: "Plazo en días negativo",
			mutator: func(p *domain.ParametrosLiquidacion) {
				p.PlazoDias = -5
			},
			expectedErr: domain.ErrPlazoInvalido,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			params := baseValid
			tt.mutator(&params)

			_, err := calc.CalcularLiquidacion(params)
			if err != tt.expectedErr {
				t.Errorf("esperado error %v, obtenido %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCompararOfertas(t *testing.T) {
	t.Parallel()

	calc := quotation.NewCalculator()

	ofertaCampo := domain.ParametrosLiquidacion{
		Cabezas:               50,
		PesoBrutoKg:           decimal.NewFromInt(320),
		EsPesoIndividual:      true,
		PrecioPorKg:           decimal.NewFromInt(2850),
		PorcentajeDesbaste:    decimal.NewFromInt(3),
		PorcentajeComision:    decimal.NewFromInt(2),
		PorcentajeIVAComision: decimal.NewFromInt(21),
		PlazoDias:             0,
	}

	ofertaMercado := domain.ParametrosLiquidacion{
		Cabezas:               50,
		PesoBrutoKg:           decimal.NewFromInt(320),
		EsPesoIndividual:      true,
		PrecioPorKg:           decimal.NewFromInt(3100),
		PorcentajeDesbaste:    decimal.Zero,
		PorcentajeComision:    decimal.NewFromInt(3),
		PorcentajeIVAComision: decimal.NewFromInt(21),
		CostoFleteTotal:       decimal.NewFromInt(500000),
		OtrosCostosFijos:      decimal.NewFromInt(80000),
		PlazoDias:             14,
		TasaInteresAnual:      decimal.NewFromInt(42),
	}

	comparativa, err := calc.CompararOfertas([]domain.ParametrosLiquidacion{ofertaCampo, ofertaMercado})
	if err != nil {
		t.Fatalf("error comparando ofertas: %v", err)
	}

	if len(comparativa) != 2 {
		t.Fatalf("se esperaban 2 resultados, obtenidos %d", len(comparativa))
	}

	// Ambas ofertas deben tener resultados coherentes y estrictamente positivos
	if comparativa[0].CobroNetoEfectivo.LessThanOrEqual(decimal.Zero) {
		t.Errorf("oferta 1 cobro neto no positivo")
	}
	if comparativa[1].CobroNetoEfectivo.LessThanOrEqual(decimal.Zero) {
		t.Errorf("oferta 2 cobro neto no positivo")
	}
}
