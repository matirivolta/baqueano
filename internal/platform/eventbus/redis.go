package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/matirivolta/baqueano/internal/domain"
)

const (
	DefaultStreamName = domain.CotizacionesStreamName
	DefaultGroupName  = "alert-dispatcher-group"
	PayloadField      = "payload"
)

var (
	ErrEmptyPayload = errors.New("el mensaje en Redis Streams no contiene campo payload")
)

// Producer define el contrato de publicación de eventos de mercado hacia el Event Bus.
type Producer interface {
	PublishCotizacion(ctx context.Context, event domain.CotizacionIngeridaEvent) (string, error)
}

// HandlerFunc es la función invocada al recibir un evento desde el stream.
type HandlerFunc func(ctx context.Context, event domain.CotizacionIngeridaEvent) error

// Consumer define el contrato de suscripción resiliente mediante Consumer Groups.
type Consumer interface {
	EnsureConsumerGroup(ctx context.Context, stream, group string) error
	Consume(ctx context.Context, stream, group, consumerName string, handler HandlerFunc) error
}

// RedisEventBus encapsula la interacción con Redis Streams para publicación y consumo confiable.
type RedisEventBus struct {
	client *redis.Client
}

// NewRedisEventBus construye una nueva instancia del bus de eventos sobre Redis.
func NewRedisEventBus(client *redis.Client) *RedisEventBus {
	return &RedisEventBus{client: client}
}

// PublishCotizacion serializa y publica el evento en cotizaciones:stream mediante XADD.
func (b *RedisEventBus) PublishCotizacion(ctx context.Context, event domain.CotizacionIngeridaEvent) (string, error) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("error serializando evento de cotización: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: DefaultStreamName,
		Values: map[string]interface{}{
			PayloadField:      string(payloadBytes),
			"event_id":         event.EventID,
			"categoria_codigo": event.CategoriaCodigo,
			"fuente":           event.Fuente,
		},
	}

	msgID, err := b.client.XAdd(ctx, args).Result()
	if err != nil {
		return "", fmt.Errorf("error ejecutando XADD en Redis Streams: %w", err)
	}

	return msgID, nil
}

// EnsureConsumerGroup crea el grupo de consumidores si no existe (idempotente frente a BUSYGROUP).
func (b *RedisEventBus) EnsureConsumerGroup(ctx context.Context, stream, group string) error {
	err := b.client.XGroupCreateMkStream(ctx, stream, group, "$").Err()
	if err != nil {
		// En Redis, si el grupo ya existe, devuelve "BUSYGROUP Consumer Group name already exists"
		if strings.Contains(err.Error(), "BUSYGROUP") {
			return nil
		}
		return fmt.Errorf("error creando consumer group %s en stream %s: %w", group, stream, err)
	}
	return nil
}

// Consume ejecuta el loop de escucha con XREADGROUP, despachando eventos al handler y confirmando con XACK.
func (b *RedisEventBus) Consume(ctx context.Context, stream, group, consumerName string, handler HandlerFunc) error {
	if err := b.EnsureConsumerGroup(ctx, stream, group); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		entries, err := b.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumerName,
			Streams:  []string{stream, ">"},
			Count:    10,
			Block:    2 * time.Second,
		}).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			return fmt.Errorf("error leyendo de Redis Streams: %w", err)
		}

		for _, streamStream := range entries {
			for _, message := range streamStream.Messages {
				payloadRaw, exists := message.Values[PayloadField]
				if !exists {
					continue
				}

				payloadStr, ok := payloadRaw.(string)
				if !ok {
					continue
				}

				var event domain.CotizacionIngeridaEvent
				if err := json.Unmarshal([]byte(payloadStr), &event); err != nil {
					// Mensaje malformado: loggear o descartar para no bloquear el stream
					continue
				}

				if err := handler(ctx, event); err == nil {
					// XACK solo tras procesamiento exitoso
					_ = b.client.XAck(ctx, stream, group, message.ID).Err()
				}
			}
		}
	}
}

