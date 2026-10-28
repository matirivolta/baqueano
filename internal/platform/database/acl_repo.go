package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/matirivolta/baqueano/internal/db"
)

// MapeoFuenteResult representa el mapeo homologado o pendiente entre fuente externa y categoría canónica.
type MapeoFuenteResult struct {
	ID                  int32  `json:"id"`
	Fuente              string `json:"fuente"`
	TextoOrigen         string `json:"texto_origen"`
	CategoriaCanonID    *int32 `json:"categoria_canon_id,omitempty"`
	VerificadoPorHumano bool   `json:"verificado_por_humano"`
}

// PostgresACLRepository gestiona la traducción y persistencia del Anti-Corruption Layer.
type PostgresACLRepository struct {
	q db.Querier
}

// NewACLRepository construye una nueva instancia del repositorio de ACL.
func NewACLRepository(q db.Querier) *PostgresACLRepository {
	return &PostgresACLRepository{q: q}
}

// GetMapeo busca el mapeo canónico configurado para el texto crudo extraído de una fuente.
func (r *PostgresACLRepository) GetMapeo(ctx context.Context, fuente, textoOrigen string) (*MapeoFuenteResult, error) {
	row, err := r.q.GetMapeoFuente(ctx, db.GetMapeoFuenteParams{
		Fuente:      fuente,
		TextoOrigen: textoOrigen,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error consultando mapeo fuente (%s: %s): %w", fuente, textoOrigen, err)
	}

	var catID *int32
	if row.CategoriaCanonID.Valid {
		id := row.CategoriaCanonID.Int32
		catID = &id
	}

	return &MapeoFuenteResult{
		ID:                  row.ID,
		Fuente:              row.Fuente,
		TextoOrigen:         row.TextoOrigen,
		CategoriaCanonID:    catID,
		VerificadoPorHumano: row.VerificadoPorHumano,
	}, nil
}

// UpsertMapeo registra o actualiza un término de mercado, admitiendo categoría nula si está pendiente de revisión.
func (r *PostgresACLRepository) UpsertMapeo(ctx context.Context, fuente, textoOrigen string, categoriaID *int32, verificado bool) error {
	var catID pgtype.Int4
	if categoriaID != nil {
		catID = pgtype.Int4{Int32: *categoriaID, Valid: true}
	}

	_, err := r.q.UpsertMapeoFuente(ctx, db.UpsertMapeoFuenteParams{
		Fuente:              fuente,
		TextoOrigen:         textoOrigen,
		CategoriaCanonID:    catID,
		VerificadoPorHumano: verificado,
	})
	if err != nil {
		return fmt.Errorf("error guardando mapeo en ACL: %w", err)
	}
	return nil
}

