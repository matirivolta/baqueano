package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/db"
	"github.com/matirivolta/baqueano/internal/domain"
)

// PostgresQuotationRepository implementa quotation.QuotationRepository mediado por sqlc.
type PostgresQuotationRepository struct {
	q db.Querier
}

// NewQuotationRepository construye una nueva instancia del repositorio de cotizaciones.
func NewQuotationRepository(q db.Querier) *PostgresQuotationRepository {
	return &PostgresQuotationRepository{q: q}
}

// GetUltimasCotizaciones recupera el tablero consolidado de últimas cotizaciones vigentes por categoría y plaza.
func (r *PostgresQuotationRepository) GetUltimasCotizaciones(ctx context.Context) ([]domain.Cotizacion, error) {
	rows, err := r.q.GetUltimasCotizacionesPorCategoria(ctx)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo últimas cotizaciones: %w", err)
	}

	result := make([]domain.Cotizacion, len(rows))
	for i, row := range rows {
		var pMin, pMax *decimal.Decimal
		if row.PrecioMin.Valid {
			val := row.PrecioMin.Decimal
			pMin = &val
		}
		if row.PrecioMax.Valid {
			val := row.PrecioMax.Decimal
			pMax = &val
		}

		vol := 0
		if row.VolumenCabezas.Valid {
			vol = int(row.VolumenCabezas.Int32)
		}

		result[i] = domain.Cotizacion{
			ID:               row.ID,
			Fecha:            row.Fecha.Time,
			Fuente:           row.Fuente,
			CategoriaCanonID: row.CategoriaCanonID,
			CategoriaCodigo:  row.CategoriaCodigo,
			CategoriaNombre:  row.CategoriaNombre,
			TipoDestino:      domain.TipoDestino(row.TipoDestino),
			PrecioMin:        pMin,
			PrecioMax:        pMax,
			PrecioPromedio:   row.PrecioPromedio,
			VolumenCabezas:   vol,
		}
	}
	return result, nil
}

// GetSerieHistorica retorna las cotizaciones históricas para una categoría desde una fecha dada.
func (r *PostgresQuotationRepository) GetSerieHistorica(ctx context.Context, categoriaID int32, desde time.Time) ([]domain.Cotizacion, error) {
	rows, err := r.q.GetSerieHistorica(ctx, db.GetSerieHistoricaParams{
		CategoriaCanonID: categoriaID,
		Fecha:            pgtype.Date{Time: desde, Valid: !desde.IsZero()},
	})
	if err != nil {
		return nil, fmt.Errorf("error obteniendo serie histórica para categoría %d: %w", categoriaID, err)
	}

	result := make([]domain.Cotizacion, len(rows))
	for i, row := range rows {
		var pMin, pMax *decimal.Decimal
		if row.PrecioMin.Valid {
			val := row.PrecioMin.Decimal
			pMin = &val
		}
		if row.PrecioMax.Valid {
			val := row.PrecioMax.Decimal
			pMax = &val
		}

		vol := 0
		if row.VolumenCabezas.Valid {
			vol = int(row.VolumenCabezas.Int32)
		}

		result[i] = domain.Cotizacion{
			Fecha:            row.Fecha.Time,
			Fuente:           row.Fuente,
			CategoriaCanonID: categoriaID,
			PrecioMin:        pMin,
			PrecioMax:        pMax,
			PrecioPromedio:   row.PrecioPromedio,
			VolumenCabezas:   vol,
		}
	}
	return result, nil
}

// GetCotizacionPrevia obtiene la última cotización previa a la fecha dada para evaluar variaciones interdiarias.
func (r *PostgresQuotationRepository) GetCotizacionPrevia(ctx context.Context, categoriaID int32, fuente string, fecha time.Time) (*domain.Cotizacion, error) {
	row, err := r.q.GetCotizacionPrevia(ctx, db.GetCotizacionPreviaParams{
		CategoriaCanonID: categoriaID,
		Fuente:           fuente,
		Fecha:            pgtype.Date{Time: fecha, Valid: !fecha.IsZero()},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo cotización previa: %w", err)
	}

	var pMin, pMax *decimal.Decimal
	if row.PrecioMin.Valid {
		val := row.PrecioMin.Decimal
		pMin = &val
	}
	if row.PrecioMax.Valid {
		val := row.PrecioMax.Decimal
		pMax = &val
	}

	vol := 0
	if row.VolumenCabezas.Valid {
		vol = int(row.VolumenCabezas.Int32)
	}

	return &domain.Cotizacion{
		ID:               row.ID,
		Fecha:            row.Fecha.Time,
		Fuente:           row.Fuente,
		CategoriaCanonID: row.CategoriaCanonID,
		PrecioMin:        pMin,
		PrecioMax:        pMax,
		PrecioPromedio:   row.PrecioPromedio,
		VolumenCabezas:   vol,
		CreatedAt:        row.CreatedAt.Time,
	}, nil
}

// UpsertCotizacion inserta o actualiza una cotización con resolución de conflictos por clave natural.
func (r *PostgresQuotationRepository) UpsertCotizacion(ctx context.Context, cot domain.Cotizacion) error {
	if err := cot.Validate(); err != nil {
		return fmt.Errorf("validación de cotización falló: %w", err)
	}

	var pMin, pMax decimal.NullDecimal
	if cot.PrecioMin != nil {
		pMin = decimal.NullDecimal{Decimal: *cot.PrecioMin, Valid: true}
	}
	if cot.PrecioMax != nil {
		pMax = decimal.NullDecimal{Decimal: *cot.PrecioMax, Valid: true}
	}

	metadataBytes := []byte("{}")
	if len(cot.Metadata) > 0 {
		var err error
		metadataBytes, err = json.Marshal(cot.Metadata)
		if err != nil {
			return fmt.Errorf("error serializando metadata de cotización: %w", err)
		}
	}

	_, err := r.q.UpsertCotizacion(ctx, db.UpsertCotizacionParams{
		Fecha:            pgtype.Date{Time: cot.Fecha, Valid: true},
		Fuente:           cot.Fuente,
		CategoriaCanonID: cot.CategoriaCanonID,
		PrecioMin:        pMin,
		PrecioMax:        pMax,
		PrecioPromedio:   cot.PrecioPromedio,
		VolumenCabezas:   pgtype.Int4{Int32: int32(cot.VolumenCabezas), Valid: true},
		Metadata:         metadataBytes,
	})
	if err != nil {
		return fmt.Errorf("error ejecutando UpsertCotizacion: %w", err)
	}

	return nil
}

