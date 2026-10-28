package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/matirivolta/baqueano/internal/db"
	"github.com/matirivolta/baqueano/internal/domain"
)

// PostgresAlertRepository gestiona el registro de auditoría e idempotencia de alertas emitidas.
type PostgresAlertRepository struct {
	q db.Querier
}

// NewAlertRepository crea una nueva instancia del repositorio de auditoría de alertas.
func NewAlertRepository(q db.Querier) *PostgresAlertRepository {
	return &PostgresAlertRepository{q: q}
}

// CheckAlertaEmitida verifica si ya se emitió una alerta idéntica en la fecha especificada (garantía de idempotencia).
func (r *PostgresAlertRepository) CheckAlertaEmitida(ctx context.Context, fecha time.Time, categoriaID int32, tipo domain.TipoAlerta, canal domain.CanalNotificacion) (bool, error) {
	emitida, err := r.q.CheckAlertaEmitida(ctx, db.CheckAlertaEmitidaParams{
		Fecha:            pgtype.Date{Time: fecha, Valid: true},
		CategoriaCanonID: categoriaID,
		TipoAlerta:       string(tipo),
		Canal:            string(canal),
	})
	if err != nil {
		return false, fmt.Errorf("error verificando idempotencia de alerta: %w", err)
	}
	return emitida, nil
}

// RegistrarAlerta almacena la notificación despachada para auditoría histórica.
func (r *PostgresAlertRepository) RegistrarAlerta(ctx context.Context, alerta domain.Alerta) error {
	if err := alerta.Validate(); err != nil {
		return fmt.Errorf("alerta inválida: %w", err)
	}

	metadataBytes := []byte("{}")
	if len(alerta.Metadata) > 0 {
		var err error
		metadataBytes, err = json.Marshal(alerta.Metadata)
		if err != nil {
			return fmt.Errorf("error serializando metadata de alerta: %w", err)
		}
	}

	_, err := r.q.RegistrarAlerta(ctx, db.RegistrarAlertaParams{
		Fecha:            pgtype.Date{Time: alerta.Fecha, Valid: true},
		CategoriaCanonID: alerta.CategoriaCanonID,
		TipoAlerta:       string(alerta.TipoAlerta),
		Canal:            string(alerta.Canal),
		Mensaje:          alerta.Mensaje,
		Metadata:         metadataBytes,
	})
	if err != nil {
		return fmt.Errorf("error insertando alerta en alertas_log: %w", err)
	}

	return nil
}

