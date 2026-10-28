package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// CotizacionesStreamName es el tópico/stream canónico en Redis Streams para cotizaciones ingeridas.
const CotizacionesStreamName = "cotizaciones:stream"

// CotizacionIngeridaEvent representa el evento de dominio publicado tras el scraping y normalización exitosa.
type CotizacionIngeridaEvent struct {
	EventID          string           `json:"event_id"`
	CotizacionID     int64            `json:"cotizacion_id"`
	Fecha            time.Time        `json:"fecha"`
	Fuente           string           `json:"fuente"`
	CategoriaCanonID int32            `json:"categoria_canon_id"`
	CategoriaCodigo  string           `json:"categoria_codigo"`
	PrecioPromedio   decimal.Decimal  `json:"precio_promedio"`
	PrecioMin        *decimal.Decimal `json:"precio_min,omitempty"`
	PrecioMax        *decimal.Decimal `json:"precio_max,omitempty"`
	VolumenCabezas   int              `json:"volumen_cabezas"`
	Timestamp        time.Time        `json:"timestamp"`
}

// AlertaDespachadaEvent es emitido una vez confirmada la entrega al canal (ej: Telegram).
type AlertaDespachadaEvent struct {
	AlertaID         string            `json:"alerta_id"`
	TipoAlerta       TipoAlerta        `json:"tipo_alerta"`
	Canal            CanalNotificacion `json:"canal"`
	Fecha            time.Time         `json:"fecha"`
	CategoriaCanonID int32             `json:"categoria_canon_id"`
	DisparadoAt      time.Time         `json:"disparado_at"`
}

