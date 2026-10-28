package domain

import (
	"errors"
	"strings"
	"time"
)

// TipoDestino representa el destino zootécnico y comercial de la hacienda.
type TipoDestino string

const (
	DestinoFaena     TipoDestino = "FAENA"
	DestinoInvernada TipoDestino = "INVERNADA"
	DestinoCria      TipoDestino = "CRIA"
)

var (
	ErrCodigoVacio        = errors.New("el código de la categoría no puede estar vacío")
	ErrNombreVacio        = errors.New("el nombre de la categoría no puede estar vacío")
	ErrDestinoInvalido     = errors.New("tipo de destino inválido (debe ser FAENA, INVERNADA o CRIA)")
	ErrRangoPesoInvalido  = errors.New("el peso mínimo no puede ser mayor al peso máximo")
)

// CategoriaCanon representa una categoría canónica normalizada en el sistema.
type CategoriaCanon struct {
	ID            int32       `json:"id"`
	Codigo        string      `json:"codigo"`
	NombreVisible string      `json:"nombre_visible"`
	TipoDestino   TipoDestino `json:"tipo_destino"`
	PesoMinimoKg  *int        `json:"peso_minimo_kg,omitempty"`
	PesoMaximoKg  *int        `json:"peso_maximo_kg,omitempty"`
	Activo        bool        `json:"activo"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// Validate valida las invariantes de negocio de la categoría canónica.
func (c CategoriaCanon) Validate() error {
	if strings.TrimSpace(c.Codigo) == "" {
		return ErrCodigoVacio
	}
	if strings.TrimSpace(c.NombreVisible) == "" {
		return ErrNombreVacio
	}
	switch c.TipoDestino {
	case DestinoFaena, DestinoInvernada, DestinoCria:
	default:
		return ErrDestinoInvalido
	}
	if c.PesoMinimoKg != nil && c.PesoMaximoKg != nil {
		if *c.PesoMinimoKg > *c.PesoMaximoKg {
			return ErrRangoPesoInvalido
		}
	}
	return nil
}

// MatchesWeight verifica si un peso determinado en kg encuadra en el rango de la categoría.
func (c CategoriaCanon) MatchesWeight(kg int) bool {
	if c.PesoMinimoKg != nil && kg < *c.PesoMinimoKg {
		return false
	}
	if c.PesoMaximoKg != nil && kg > *c.PesoMaximoKg {
		return false
	}
	return true
}

