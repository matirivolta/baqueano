package quotation

import (
	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/domain"
)

var (
	cien     = decimal.NewFromInt(100)
	diasAnio = decimal.NewFromInt(365)
)

// DefaultCalculator implementa la interfaz Calculator aplicando las reglas contables
// y financieras de la comercialización ganadera argentina.
type DefaultCalculator struct{}

// NewCalculator retorna una nueva instancia del calculador de liquidaciones ganaderas.
func NewCalculator() *DefaultCalculator {
	return &DefaultCalculator{}
}

// CalcularLiquidacion ejecuta el desglose matemático: pesaje con desbaste, deducciones
// comerciales (comisión de consignataria, IVA 21%, flete y gastos) y descuento financiero por plazo.
func (c *DefaultCalculator) CalcularLiquidacion(p domain.ParametrosLiquidacion) (domain.LiquidacionNeta, error) {
	if err := p.Validate(); err != nil {
		return domain.LiquidacionNeta{}, err
	}

	cabezasDec := decimal.NewFromInt(int64(p.Cabezas))

	// 1. Kilos Brutos y peso promedio bruto
	var kilosBrutos, pesoPromedioBruto decimal.Decimal
	if p.EsPesoIndividual {
		pesoPromedioBruto = p.PesoBrutoKg.Round(2)
		kilosBrutos = pesoPromedioBruto.Mul(cabezasDec).Round(2)
	} else {
		kilosBrutos = p.PesoBrutoKg.Round(2)
		pesoPromedioBruto = kilosBrutos.Div(cabezasDec).Round(2)
	}

	// 2. Desbaste en balanza
	kilosDesbaste := kilosBrutos.Mul(p.PorcentajeDesbaste).Div(cien).Round(2)
	kilosNetos := kilosBrutos.Sub(kilosDesbaste).Round(2)
	pesoPromedioNeto := kilosNetos.Div(cabezasDec).Round(2)

	// 3. Importe Bruto Hacienda
	importeBruto := kilosNetos.Mul(p.PrecioPorKg).Round(2)

	// 4. Deducciones Comerciales e Impositivas
	montoComision := importeBruto.Mul(p.PorcentajeComision).Div(cien).Round(2)
	montoIVAComision := montoComision.Mul(p.PorcentajeIVAComision).Div(cien).Round(2)
	montoFlete := p.CostoFleteTotal.Round(2)
	montoOtrosGastos := p.OtrosCostosFijos.Round(2)

	totalDeducciones := montoComision.Add(montoIVAComision).Add(montoFlete).Add(montoOtrosGastos).Round(2)
	subtotalSinDescuento := importeBruto.Sub(totalDeducciones).Round(2)

	// 5. Costo Financiero (descuento comercial simple: TNA * plazo / 365)
	costoFinanciero := decimal.Zero
	if p.PlazoDias > 0 && p.TasaInteresAnual.GreaterThan(decimal.Zero) {
		diasDec := decimal.NewFromInt(int64(p.PlazoDias))
		factor := p.TasaInteresAnual.Div(cien).Mul(diasDec).Div(diasAnio)
		costoFinanciero = subtotalSinDescuento.Mul(factor).Round(2)
	}

	cobroNetoEfectivo := subtotalSinDescuento.Sub(costoFinanciero).Round(2)

	// 6. Precios Netos de Bolsillo Unitarios
	precioNetoPorKiloNeto := decimal.Zero
	if kilosNetos.GreaterThan(decimal.Zero) {
		precioNetoPorKiloNeto = cobroNetoEfectivo.Div(kilosNetos).Round(2)
	}

	precioNetoPorKiloBruto := decimal.Zero
	if kilosBrutos.GreaterThan(decimal.Zero) {
		precioNetoPorKiloBruto = cobroNetoEfectivo.Div(kilosBrutos).Round(2)
	}

	return domain.LiquidacionNeta{
		Cabezas:                        p.Cabezas,
		KilosBrutos:                    kilosBrutos,
		KilosDesbaste:                  kilosDesbaste,
		KilosNetos:                     kilosNetos,
		PesoPromedioBrutoKg:            pesoPromedioBruto,
		PesoPromedioNetoKg:             pesoPromedioNeto,
		PrecioPorKg:                    p.PrecioPorKg.Round(2),
		ImporteBruto:                   importeBruto,
		PorcentajeComision:             p.PorcentajeComision.Round(2),
		MontoComision:                  montoComision,
		MontoIVAComision:               montoIVAComision,
		MontoFlete:                     montoFlete,
		MontoOtrosGastos:               montoOtrosGastos,
		TotalDeduccionesComerciales:    totalDeducciones,
		SubtotalSinDescuentoFinanciero: subtotalSinDescuento,
		PlazoDias:                      p.PlazoDias,
		TasaInteresAnual:               p.TasaInteresAnual.Round(2),
		CostoFinanciero:                costoFinanciero,
		CobroNetoEfectivo:              cobroNetoEfectivo,
		PrecioNetoPorKiloNeto:          precioNetoPorKiloNeto,
		PrecioNetoPorKiloBruto:         precioNetoPorKiloBruto,
	}, nil
}

// CompararOfertas evalúa múltiples propuestas de comercialización y devuelve sus liquidaciones
// para determinar cuál maximiza el cobro neto efectivo.
func (c *DefaultCalculator) CompararOfertas(ofertas []domain.ParametrosLiquidacion) ([]domain.LiquidacionNeta, error) {
	resultados := make([]domain.LiquidacionNeta, 0, len(ofertas))
	for _, oferta := range ofertas {
		res, err := c.CalcularLiquidacion(oferta)
		if err != nil {
			return nil, err
		}
		resultados = append(resultados, res)
	}
	return resultados, nil
}

