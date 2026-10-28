package domain

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrCabezasInvalidas    = errors.New("la cantidad de cabezas debe ser mayor a cero")
	ErrPesoInvalido        = errors.New("el peso bruto debe ser estrictamente positivo")
	ErrPrecioInvalido      = errors.New("el precio por kilogramo debe ser estrictamente positivo")
	ErrDesbasteInvalido    = errors.New("el porcentaje de desbaste debe estar entre 0% y 99.99%")
	ErrComisionInvalida    = errors.New("el porcentaje de comisión no puede ser negativo ni superar el 100%")
	ErrIVAInvalido         = errors.New("el porcentaje de IVA de comisión no puede ser negativo")
	ErrCostoNegativo       = errors.New("los costos de flete u otros gastos no pueden ser negativos")
	ErrPlazoInvalido       = errors.New("el plazo en días no puede ser negativo")
	ErrTasaInteresInvalida = errors.New("la tasa de interés no puede ser negativa")
)

// ParametrosLiquidacion contiene las variables de entrada requeridas para simular la liquidación ganadera.
type ParametrosLiquidacion struct {
	Cabezas               int             `json:"cabezas"`
	PesoBrutoKg           decimal.Decimal `json:"peso_bruto_kg"`
	EsPesoIndividual      bool            `json:"es_peso_individual"`
	PrecioPorKg           decimal.Decimal `json:"precio_por_kg"`
	PorcentajeDesbaste    decimal.Decimal `json:"porcentaje_desbaste"`
	PorcentajeComision    decimal.Decimal `json:"porcentaje_comision"`
	PorcentajeIVAComision decimal.Decimal `json:"porcentaje_iva_comision"`
	CostoFleteTotal       decimal.Decimal `json:"costo_flete_total"`
	OtrosCostosFijos      decimal.Decimal `json:"otros_costos_fijos"`
	PlazoDias             int             `json:"plazo_dias"`
	TasaInteresAnual      decimal.Decimal `json:"tasa_interes_anual"`
}

// Validate asegura que todos los parámetros de entrada respeten las reglas del dominio ganadero.
func (p ParametrosLiquidacion) Validate() error {
	if p.Cabezas <= 0 {
		return ErrCabezasInvalidas
	}
	if p.PesoBrutoKg.LessThanOrEqual(decimal.Zero) {
		return ErrPesoInvalido
	}
	if p.PrecioPorKg.LessThanOrEqual(decimal.Zero) {
		return ErrPrecioInvalido
	}
	if p.PorcentajeDesbaste.LessThan(decimal.Zero) || p.PorcentajeDesbaste.GreaterThanOrEqual(decimal.NewFromInt(100)) {
		return ErrDesbasteInvalido
	}
	if p.PorcentajeComision.LessThan(decimal.Zero) || p.PorcentajeComision.GreaterThanOrEqual(decimal.NewFromInt(100)) {
		return ErrComisionInvalida
	}
	if p.PorcentajeIVAComision.LessThan(decimal.Zero) {
		return ErrIVAInvalido
	}
	if p.CostoFleteTotal.LessThan(decimal.Zero) || p.OtrosCostosFijos.LessThan(decimal.Zero) {
		return ErrCostoNegativo
	}
	if p.PlazoDias < 0 {
		return ErrPlazoInvalido
	}
	if p.TasaInteresAnual.LessThan(decimal.Zero) {
		return ErrTasaInteresInvalida
	}
	return nil
}

// LiquidacionNeta representa el desglose financiero integral y resultado neto de la venta de hacienda.
type LiquidacionNeta struct {
	Cabezas                         int             `json:"cabezas"`
	KilosBrutos                     decimal.Decimal `json:"kilos_brutos"`
	KilosDesbaste                   decimal.Decimal `json:"kilos_desbaste"`
	KilosNetos                      decimal.Decimal `json:"kilos_netos"`
	PesoPromedioBrutoKg             decimal.Decimal `json:"peso_promedio_bruto_kg"`
	PesoPromedioNetoKg              decimal.Decimal `json:"peso_promedio_neto_kg"`
	PrecioPorKg                     decimal.Decimal `json:"precio_por_kg"`
	ImporteBruto                    decimal.Decimal `json:"importe_bruto"`
	PorcentajeComision              decimal.Decimal `json:"porcentaje_comision"`
	MontoComision                   decimal.Decimal `json:"monto_comision"`
	MontoIVAComision                decimal.Decimal `json:"monto_iva_comision"`
	MontoFlete                      decimal.Decimal `json:"monto_flete"`
	MontoOtrosGastos                decimal.Decimal `json:"monto_otros_gastos"`
	TotalDeduccionesComerciales     decimal.Decimal `json:"total_deducciones_comerciales"`
	SubtotalSinDescuentoFinanciero  decimal.Decimal `json:"subtotal_sin_descuento_financiero"`
	PlazoDias                       int             `json:"plazo_dias"`
	TasaInteresAnual                decimal.Decimal `json:"tasa_interes_anual"`
	CostoFinanciero                 decimal.Decimal `json:"costo_financiero"`
	CobroNetoEfectivo               decimal.Decimal `json:"cobro_neto_efectivo"`
	PrecioNetoPorKiloNeto           decimal.Decimal `json:"precio_neto_por_kilo_neto"`
	PrecioNetoPorKiloBruto          decimal.Decimal `json:"precio_neto_por_kilo_bruto"`
}

