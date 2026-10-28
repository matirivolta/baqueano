package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	FuenteMAG        = "MAG"
	FuenteRosgan     = "ROSGAN"
	FuenteTattersall = "TATTERSALL"
	FuenteDeCampo    = "DECAMPO"
)

var (
	ErrFuenteVacia          = errors.New("la fuente de mercado no puede estar vacía")
	ErrFechaInvalida        = errors.New("la fecha de cotización no puede ser nula")
	ErrPrecioPromedioInvalido = errors.New("el precio promedio debe ser estrictamente positivo")
	ErrRangoPreciosInvalido = errors.New("el precio mínimo no puede ser mayor al precio máximo")
	ErrVolumenNegativo      = errors.New("el volumen de cabezas no puede ser negativo")
)

// Cotizacion representa el precio de mercado registrado para una categoría en una fecha y plaza dada.
type Cotizacion struct {
	ID               int64                  `json:"id"`
	Fecha            time.Time              `json:"fecha"`
	Fuente           string                 `json:"fuente"`
	CategoriaCanonID int32                  `json:"categoria_canon_id"`
	CategoriaCodigo  string                 `json:"categoria_codigo,omitempty"`
	CategoriaNombre  string                 `json:"categoria_nombre,omitempty"`
	TipoDestino      TipoDestino            `json:"tipo_destino,omitempty"`
	PrecioMin        *decimal.Decimal       `json:"precio_min,omitempty"`
	PrecioMax        *decimal.Decimal       `json:"precio_max,omitempty"`
	PrecioPromedio   decimal.Decimal        `json:"precio_promedio"`
	VolumenCabezas   int                    `json:"volumen_cabezas"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

// Validate asegura la consistencia de los datos de la cotización.
func (c Cotizacion) Validate() error {
	if strings.TrimSpace(c.Fuente) == "" {
		return ErrFuenteVacia
	}
	if c.Fecha.IsZero() {
		return ErrFechaInvalida
	}
	if c.PrecioPromedio.LessThanOrEqual(decimal.Zero) {
		return ErrPrecioPromedioInvalido
	}
	if c.VolumenCabezas < 0 {
		return ErrVolumenNegativo
	}
	if c.PrecioMin != nil && c.PrecioMax != nil {
		if c.PrecioMin.GreaterThan(*c.PrecioMax) {
			return ErrRangoPreciosInvalido
		}
	}
	return nil
}

// Spread calcula la dispersión de precios (Max - Min) si ambos extremos están definidos.
func (c Cotizacion) Spread() (decimal.Decimal, bool) {
	if c.PrecioMin == nil || c.PrecioMax == nil {
		return decimal.Zero, false
	}
	return c.PrecioMax.Sub(*c.PrecioMin), true
}

