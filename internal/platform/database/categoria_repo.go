package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/matirivolta/baqueano/internal/db"
	"github.com/matirivolta/baqueano/internal/domain"
)

var ErrCategoriaNotFound = errors.New("categoría canónica no encontrada")

// PostgresCategoriaRepository implementa quotation.CategoriaRepository mediado por sqlc.
type PostgresCategoriaRepository struct {
	q db.Querier
}

// NewCategoriaRepository crea una nueva instancia del repositorio de categorías.
func NewCategoriaRepository(q db.Querier) *PostgresCategoriaRepository {
	return &PostgresCategoriaRepository{q: q}
}

// ListCategorias retorna todas las categorías activas ordenadas por destino y nombre.
func (r *PostgresCategoriaRepository) ListCategorias(ctx context.Context) ([]domain.CategoriaCanon, error) {
	rows, err := r.q.ListCategoriasCanon(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listando categorías canónicas: %w", err)
	}

	result := make([]domain.CategoriaCanon, len(rows))
	for i, row := range rows {
		result[i] = mapDBCategoria(row)
	}
	return result, nil
}

// GetCategoriaByCodigo busca una categoría canónica por su código único de negocio.
func (r *PostgresCategoriaRepository) GetCategoriaByCodigo(ctx context.Context, codigo string) (domain.CategoriaCanon, error) {
	row, err := r.q.GetCategoriaCanonByCodigo(ctx, codigo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CategoriaCanon{}, ErrCategoriaNotFound
		}
		return domain.CategoriaCanon{}, fmt.Errorf("error buscando categoría por código %s: %w", codigo, err)
	}
	return mapDBCategoria(row), nil
}

// GetCategoriaByID busca una categoría por su clave primaria serial.
func (r *PostgresCategoriaRepository) GetCategoriaByID(ctx context.Context, id int32) (domain.CategoriaCanon, error) {
	row, err := r.q.GetCategoriaCanonByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CategoriaCanon{}, ErrCategoriaNotFound
		}
		return domain.CategoriaCanon{}, fmt.Errorf("error buscando categoría por id %d: %w", id, err)
	}
	return mapDBCategoria(row), nil
}

func mapDBCategoria(row db.CategoriasCanon) domain.CategoriaCanon {
	cat := domain.CategoriaCanon{
		ID:            row.ID,
		Codigo:        row.Codigo,
		NombreVisible: row.NombreVisible,
		TipoDestino:   domain.TipoDestino(row.TipoDestino),
		Activo:        row.Activo,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
	if row.PesoMinimoKg.Valid {
		min := int(row.PesoMinimoKg.Int32)
		cat.PesoMinimoKg = &min
	}
	if row.PesoMaximoKg.Valid {
		max := int(row.PesoMaximoKg.Int32)
		cat.PesoMaximoKg = &max
	}
	return cat
}

