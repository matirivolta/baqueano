package alert

import (
	"context"
	"fmt"
	"log"

	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/platform/eventbus"
)

// Dispatcher orquesta el consumo de eventos desde Redis Streams y su procesamiento reactivo en el motor de reglas.
type Dispatcher struct {
	consumer     eventbus.Consumer
	engine       *DefaultEngine
	streamName   string
	groupName    string
	consumerName string
}

// NewDispatcher construye una nueva instancia del orquestador de despacho reactivo.
func NewDispatcher(
	consumer eventbus.Consumer,
	engine *DefaultEngine,
	streamName, groupName, consumerName string,
) *Dispatcher {
	if streamName == "" {
		streamName = eventbus.DefaultStreamName
	}
	if groupName == "" {
		groupName = eventbus.DefaultGroupName
	}
	if consumerName == "" {
		consumerName = "dispatcher-worker-1"
	}

	return &Dispatcher{
		consumer:     consumer,
		engine:       engine,
		streamName:   streamName,
		groupName:    groupName,
		consumerName: consumerName,
	}
}

// Start inicia el loop de consumo y evaluación reactiva bloqueante hasta que el contexto sea cancelado.
func (d *Dispatcher) Start(ctx context.Context) error {
	log.Printf("[alert-dispatcher] Iniciando consumo en stream '%s', grupo '%s', consumidor '%s'...",
		d.streamName, d.groupName, d.consumerName)

	handler := func(ctx context.Context, event domain.CotizacionIngeridaEvent) error {
		log.Printf("[alert-dispatcher] Evento recibido: Fuente=%s Categoria=%s Precio=%s",
			event.Fuente, event.CategoriaCodigo, event.PrecioPromedio.StringFixed(2))

		if err := d.engine.ProcessEvent(ctx, event); err != nil {
			log.Printf("[alert-dispatcher] Error procesando evento %s: %v", event.EventID, err)
			return err
		}
		return nil
	}

	if err := d.consumer.Consume(ctx, d.streamName, d.groupName, d.consumerName, handler); err != nil {
		if ctx.Err() != nil {
			log.Printf("[alert-dispatcher] Consumo detenido por cancelación de contexto.")
			return nil
		}
		return fmt.Errorf("fallo fatal en loop de consumo del dispatcher: %w", err)
	}

	return nil
}

