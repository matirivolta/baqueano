package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type TipoAlerta string

const (
	AlertaVariacionDiaria        TipoAlerta = "VARIACION_DIARIA"
	AlertaOportunidadEstacional TipoAlerta = "OPORTUNIDAD_ESTACIONAL"
)

type CanalNotificacion string

const (
	CanalTelegram CanalNotificacion = "TELEGRAM"
	CanalWhatsApp CanalNotificacion = "WHATSAPP"
	CanalEmail    CanalNotificacion = "EMAIL"
)

var (
	ErrTipoAlertaInvalido = errors.New("tipo de alerta inválido")
	ErrCanalInvalido      = errors.New("canal de notificación inválido")
	ErrMensajeVacio       = errors.New("el mensaje de la alerta no puede estar vacío")
)

// Alerta representa una notificación de mercado generada para su despacho.
type Alerta struct {
	ID               string                 `json:"id"`
	Fecha            time.Time              `json:"fecha"`
	CategoriaCanonID int32                  `json:"categoria_canon_id"`
	CategoriaCodigo  string                 `json:"categoria_codigo,omitempty"`
	TipoAlerta       TipoAlerta             `json:"tipo_alerta"`
	Canal            CanalNotificacion      `json:"canal"`
	Mensaje          string                 `json:"mensaje"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	DisparadoAt      time.Time              `json:"disparado_at"`
}

// Validate asegura que la alerta tenga todos sus campos requeridos válidos.
func (a Alerta) Validate() error {
	switch a.TipoAlerta {
	case AlertaVariacionDiaria, AlertaOportunidadEstacional:
	default:
		return ErrTipoAlertaInvalido
	}

	switch a.Canal {
	case CanalTelegram, CanalWhatsApp, CanalEmail:
	default:
		return ErrCanalInvalido
	}

	if strings.TrimSpace(a.Mensaje) == "" {
		return ErrMensajeVacio
	}

	if a.Fecha.IsZero() {
		return ErrFechaInvalida
	}

	return nil
}

// CriterioVariacionDiaria encapsula la regla de cambio significativo intradiario/interdiario.
type CriterioVariacionDiaria struct {
	UmbralPorcentaje decimal.Decimal // ej: 4.0 para 4%
}

// Evaluar verifica si la variación absoluta |(nuevo - previo) / previo| supera el umbral.
func (c CriterioVariacionDiaria) Evaluar(precioPrevio, precioNuevo decimal.Decimal) (cumple bool, variacionPct decimal.Decimal) {
	if precioPrevio.IsZero() {
		return false, decimal.Zero
	}
	// variacion = ((precioNuevo - precioPrevio) / precioPrevio) * 100
	diff := precioNuevo.Sub(precioPrevio)
	variacionPct = diff.Div(precioPrevio).Mul(decimal.NewFromInt(100)).Round(2)
	absVariacion := variacionPct.Abs()
	return absVariacion.GreaterThanOrEqual(c.UmbralPorcentaje), variacionPct
}

