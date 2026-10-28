package alert

import (
	"context"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/domain"
)

// Config define los umbrales paramétricos del motor de reglas de alertas.
type Config struct {
	UmbralVariacionDiariaPct decimal.Decimal // Por defecto: 4.0%
	UmbralEstacionalPct      decimal.Decimal // Por defecto: 8.0%
}

// DefaultConfig retorna la configuración recomendada según las reglas de negocio de Baqueano.
func DefaultConfig() Config {
	return Config{
		UmbralVariacionDiariaPct: decimal.NewFromInt(4),
		UmbralEstacionalPct:      decimal.NewFromInt(8),
	}
}

// DefaultEngine implementa la evaluación de reglas y el despacho con idempotencia estricta.
type DefaultEngine struct {
	cfg       Config
	histRepo  HistoricalRepository
	alertRepo AlertRepository
	notifiers []Notifier
}

// NewEngine construye una nueva instancia del motor de reglas de alertas.
func NewEngine(cfg Config, histRepo HistoricalRepository, alertRepo AlertRepository) *DefaultEngine {
	if cfg.UmbralVariacionDiariaPct.LessThanOrEqual(decimal.Zero) {
		cfg.UmbralVariacionDiariaPct = decimal.NewFromInt(4)
	}
	if cfg.UmbralEstacionalPct.LessThanOrEqual(decimal.Zero) {
		cfg.UmbralEstacionalPct = decimal.NewFromInt(8)
	}

	return &DefaultEngine{
		cfg:       cfg,
		histRepo:  histRepo,
		alertRepo: alertRepo,
		notifiers: make([]Notifier, 0),
	}
}

// RegisterNotifier añade un adaptador de notificación (ej: Telegram, WhatsApp).
func (e *DefaultEngine) RegisterNotifier(n Notifier) {
	e.notifiers = append(e.notifiers, n)
}

// ProcessEvent evalúa las reglas contra el evento de mercado y despacha alertas garantizando idempotencia.
func (e *DefaultEngine) ProcessEvent(ctx context.Context, event domain.CotizacionIngeridaEvent) error {
	alertas, err := e.Evaluate(ctx, event)
	if err != nil {
		return fmt.Errorf("error evaluando reglas sobre evento %s: %w", event.EventID, err)
	}

	for _, alerta := range alertas {
		for _, n := range e.notifiers {
			canal := n.Canal()
			alertaCanal := alerta
			alertaCanal.Canal = canal

			// Garantía de Idempotencia: Verificar en alertas_log antes de notificar
			if e.alertRepo != nil {
				emitida, err := e.alertRepo.CheckAlertaEmitida(ctx, alertaCanal.Fecha, alertaCanal.CategoriaCanonID, alertaCanal.TipoAlerta, canal)
				if err != nil {
					return fmt.Errorf("error verificando idempotencia de alerta: %w", err)
				}
				if emitida {
					// Ya fue notificado al productor en el día de hoy para este canal
					continue
				}
			}

			// Despacho al canal externo
			if err := n.Send(ctx, alertaCanal); err != nil {
				return fmt.Errorf("falló despacho hacia canal %s: %w", canal, err)
			}

			// Auditoría y persistencia del despacho exitoso
			if e.alertRepo != nil {
				if err := e.alertRepo.RegistrarAlerta(ctx, alertaCanal); err != nil {
					return fmt.Errorf("error auditando alerta en BD: %w", err)
				}
			}
		}
	}

	return nil
}

// Evaluate aplica los criterios analíticos comparando con cotizaciones previas.
func (e *DefaultEngine) Evaluate(ctx context.Context, event domain.CotizacionIngeridaEvent) ([]domain.Alerta, error) {
	if e.histRepo == nil {
		return nil, nil
	}

	cotPrevia, err := e.histRepo.GetCotizacionPrevia(ctx, event.CategoriaCanonID, event.Fuente, event.Fecha)
	if err != nil {
		return nil, fmt.Errorf("error consultando cotización previa: %w", err)
	}
	if cotPrevia == nil || cotPrevia.PrecioPromedio.LessThanOrEqual(decimal.Zero) {
		// Primera observación: no hay base para medir variación diaria
		return nil, nil
	}

	var alertas []domain.Alerta

	// Regla 1: Variación Diaria Significativa (|Δ%| >= umbral)
	criterio := domain.CriterioVariacionDiaria{
		UmbralPorcentaje: e.cfg.UmbralVariacionDiariaPct,
	}

	cumple, variacionPct := criterio.Evaluar(cotPrevia.PrecioPromedio, event.PrecioPromedio)
	if cumple {
		icono := "📈"
		tipoMovimiento := "Suba Diaria"
		if variacionPct.LessThan(decimal.Zero) {
			icono = "📉"
			tipoMovimiento = "Baja Diaria"
		}

		catCodigo := event.CategoriaCodigo
		if strings.TrimSpace(catCodigo) == "" {
			catCodigo = fmt.Sprintf("ID %d", event.CategoriaCanonID)
		}

		signo := ""
		if variacionPct.GreaterThan(decimal.Zero) {
			signo = "+"
		}

		mensaje := fmt.Sprintf(
			"🚨 *Baqueano Alerta Ganadera* | *%s*\n\n"+
				"🐂 *Categoría:* `%s`\n"+
				"🏛 *Plaza:* %s\n"+
				"%s *Variación:* *%s%s%%*\n"+
				"💵 *Precio Actual:* $%s /kg\n"+
				"🔙 *Precio Previo:* $%s /kg\n"+
				"📊 *Volumen:* %d cabezas\n"+
				"📅 *Fecha:* %s\n\n"+
				"_Notificación automática del monitor ganadero Baqueano_",
			tipoMovimiento,
			catCodigo,
			event.Fuente,
			icono,
			signo,
			variacionPct.StringFixed(2),
			event.PrecioPromedio.StringFixed(2),
			cotPrevia.PrecioPromedio.StringFixed(2),
			event.VolumenCabezas,
			event.Fecha.Format("02/01/2006"),
		)

		alertas = append(alertas, domain.Alerta{
			Fecha:            event.Fecha,
			CategoriaCanonID: event.CategoriaCanonID,
			CategoriaCodigo:  event.CategoriaCodigo,
			TipoAlerta:       domain.AlertaVariacionDiaria,
			Mensaje:          mensaje,
			Metadata: map[string]interface{}{
				"variacion_pct": variacionPct.StringFixed(2),
				"precio_actual": event.PrecioPromedio.StringFixed(2),
				"precio_previo": cotPrevia.PrecioPromedio.StringFixed(2),
				"fuente":        event.Fuente,
			},
		})
	}

	return alertas, nil
}
