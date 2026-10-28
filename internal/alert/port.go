package alert

import (
	"context"
	"time"

	"github.com/matirivolta/baqueano/internal/domain"
)

// Notifier define el contrato para el envío de notificaciones hacia canales externos (Telegram, etc.).
type Notifier interface {
	Canal() domain.CanalNotificacion
	Send(ctx context.Context, alerta domain.Alerta) error
}

// AlertRepository define las operaciones de auditoría y verificación de idempotencia.
type AlertRepository interface {
	CheckAlertaEmitida(ctx context.Context, fecha time.Time, categoriaID int32, tipo domain.TipoAlerta, canal domain.CanalNotificacion) (bool, error)
	RegistrarAlerta(ctx context.Context, alerta domain.Alerta) error
}

// HistoricalRepository define la consulta de cotizaciones previas requerida para evaluar variaciones.
type HistoricalRepository interface {
	GetCotizacionPrevia(ctx context.Context, categoriaID int32, fuente string, fecha time.Time) (*domain.Cotizacion, error)
}

// RuleEvaluator define la evaluación de reglas reactivas sobre eventos de mercado.
type RuleEvaluator interface {
	Evaluate(ctx context.Context, event domain.CotizacionIngeridaEvent) ([]domain.Alerta, error)
}

