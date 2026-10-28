package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/db"
	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/platform/database"
)

// mockQuerier implementa db.Querier para pruebas unitarias de los adaptadores de persistencia.
type mockQuerier struct {
	db.Querier
	listCategoriasFn func(ctx context.Context) ([]db.CategoriasCanon, error)
	getByCodigoFn    func(ctx context.Context, codigo string) (db.CategoriasCanon, error)
	getUltimasFn     func(ctx context.Context) ([]db.GetUltimasCotizacionesPorCategoriaRow, error)
	getPreviaFn      func(ctx context.Context, arg db.GetCotizacionPreviaParams) (db.Cotizacione, error)
	upsertCotFn      func(ctx context.Context, arg db.UpsertCotizacionParams) (db.Cotizacione, error)
	checkAlertaFn    func(ctx context.Context, arg db.CheckAlertaEmitidaParams) (bool, error)
	registrarAlertaFn func(ctx context.Context, arg db.RegistrarAlertaParams) (db.AlertasLog, error)
	getMapeoFn       func(ctx context.Context, arg db.GetMapeoFuenteParams) (db.MapeosFuente, error)
	upsertMapeoFn    func(ctx context.Context, arg db.UpsertMapeoFuenteParams) (db.MapeosFuente, error)
}

func (m *mockQuerier) ListCategoriasCanon(ctx context.Context) ([]db.CategoriasCanon, error) {
	if m.listCategoriasFn != nil {
		return m.listCategoriasFn(ctx)
	}
	return nil, nil
}

func (m *mockQuerier) GetCategoriaCanonByCodigo(ctx context.Context, codigo string) (db.CategoriasCanon, error) {
	if m.getByCodigoFn != nil {
		return m.getByCodigoFn(ctx, codigo)
	}
	return db.CategoriasCanon{}, nil
}

func (m *mockQuerier) GetUltimasCotizacionesPorCategoria(ctx context.Context) ([]db.GetUltimasCotizacionesPorCategoriaRow, error) {
	if m.getUltimasFn != nil {
		return m.getUltimasFn(ctx)
	}
	return nil, nil
}

func (m *mockQuerier) GetCotizacionPrevia(ctx context.Context, arg db.GetCotizacionPreviaParams) (db.Cotizacione, error) {
	if m.getPreviaFn != nil {
		return m.getPreviaFn(ctx, arg)
	}
	return db.Cotizacione{}, nil
}

func (m *mockQuerier) UpsertCotizacion(ctx context.Context, arg db.UpsertCotizacionParams) (db.Cotizacione, error) {
	if m.upsertCotFn != nil {
		return m.upsertCotFn(ctx, arg)
	}
	return db.Cotizacione{}, nil
}

func (m *mockQuerier) CheckAlertaEmitida(ctx context.Context, arg db.CheckAlertaEmitidaParams) (bool, error) {
	if m.checkAlertaFn != nil {
		return m.checkAlertaFn(ctx, arg)
	}
	return false, nil
}

func (m *mockQuerier) RegistrarAlerta(ctx context.Context, arg db.RegistrarAlertaParams) (db.AlertasLog, error) {
	if m.registrarAlertaFn != nil {
		return m.registrarAlertaFn(ctx, arg)
	}
	return db.AlertasLog{}, nil
}

func (m *mockQuerier) GetMapeoFuente(ctx context.Context, arg db.GetMapeoFuenteParams) (db.MapeosFuente, error) {
	if m.getMapeoFn != nil {
		return m.getMapeoFn(ctx, arg)
	}
	return db.MapeosFuente{}, nil
}

func (m *mockQuerier) UpsertMapeoFuente(ctx context.Context, arg db.UpsertMapeoFuenteParams) (db.MapeosFuente, error) {
	if m.upsertMapeoFn != nil {
		return m.upsertMapeoFn(ctx, arg)
	}
	return db.MapeosFuente{}, nil
}

func TestCategoriaRepository_ListAndGet(t *testing.T) {
	t.Parallel()

	q := &mockQuerier{
		listCategoriasFn: func(ctx context.Context) ([]db.CategoriasCanon, error) {
			return []db.CategoriasCanon{
				{
					ID:            1,
					Codigo:        "NOVILLITO",
					NombreVisible: "Novillitos Consumo",
					TipoDestino:   "FAENA",
					PesoMinimoKg:  pgtype.Int4{Int32: 300, Valid: true},
					PesoMaximoKg:  pgtype.Int4{Int32: 390, Valid: true},
					Activo:        true,
				},
			}, nil
		},
		getByCodigoFn: func(ctx context.Context, codigo string) (db.CategoriasCanon, error) {
			if codigo == "NOVILLITO" {
				return db.CategoriasCanon{
					ID:            1,
					Codigo:        "NOVILLITO",
					NombreVisible: "Novillitos Consumo",
					TipoDestino:   "FAENA",
					Activo:        true,
				}, nil
			}
			return db.CategoriasCanon{}, pgx.ErrNoRows
		},
	}

	repo := database.NewCategoriaRepository(q)

	list, err := repo.ListCategorias(context.Background())
	if err != nil {
		t.Fatalf("error inesperado listando: %v", err)
	}
	if len(list) != 1 || list[0].Codigo != "NOVILLITO" {
		t.Fatalf("listado incorrecto: %+v", list)
	}
	if list[0].PesoMinimoKg == nil || *list[0].PesoMinimoKg != 300 {
		t.Errorf("peso mínimo incorrecto: %v", list[0].PesoMinimoKg)
	}

	cat, err := repo.GetCategoriaByCodigo(context.Background(), "NOVILLITO")
	if err != nil {
		t.Fatalf("error buscando categoría existente: %v", err)
	}
	if cat.Codigo != "NOVILLITO" {
		t.Errorf("esperado NOVILLITO, obtenido %s", cat.Codigo)
	}

	_, errNotFound := repo.GetCategoriaByCodigo(context.Background(), "INEXISTENTE")
	if errNotFound != database.ErrCategoriaNotFound {
		t.Errorf("esperado ErrCategoriaNotFound, obtenido %v", errNotFound)
	}
}

func TestQuotationRepository_GetUltimasAndUpsert(t *testing.T) {
	t.Parallel()

	precioProm := decimal.NewFromFloat(2890.50)
	precioMin := decimal.NewFromFloat(2800.00)

	var capturedParams db.UpsertCotizacionParams

	q := &mockQuerier{
		getUltimasFn: func(ctx context.Context) ([]db.GetUltimasCotizacionesPorCategoriaRow, error) {
			return []db.GetUltimasCotizacionesPorCategoriaRow{
				{
					ID:               10,
					Fecha:            pgtype.Date{Time: time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC), Valid: true},
					Fuente:           "MAG",
					CategoriaCanonID: 1,
					CategoriaCodigo:  "NOVILLITO",
					CategoriaNombre:  "Novillitos",
					TipoDestino:      "FAENA",
					PrecioPromedio:   precioProm,
					PrecioMin:        decimal.NullDecimal{Decimal: precioMin, Valid: true},
					VolumenCabezas:   pgtype.Int4{Int32: 2500, Valid: true},
				},
			}, nil
		},
		upsertCotFn: func(ctx context.Context, arg db.UpsertCotizacionParams) (db.Cotizacione, error) {
			capturedParams = arg
			return db.Cotizacione{
				ID:             1,
				PrecioPromedio: arg.PrecioPromedio,
			}, nil
		},
	}

	repo := database.NewQuotationRepository(q)

	ultimas, err := repo.GetUltimasCotizaciones(context.Background())
	if err != nil {
		t.Fatalf("error obteniendo ultimas cotizaciones: %v", err)
	}
	if len(ultimas) != 1 {
		t.Fatalf("esperada 1 cotización, obtenidas %d", len(ultimas))
	}
	if !ultimas[0].PrecioPromedio.Equal(precioProm) {
		t.Errorf("esperado %s, obtenido %s", precioProm, ultimas[0].PrecioPromedio)
	}
	if ultimas[0].PrecioMin == nil || !ultimas[0].PrecioMin.Equal(precioMin) {
		t.Errorf("esperado precioMin %s, obtenido %v", precioMin, ultimas[0].PrecioMin)
	}

	nuevaCot := domain.Cotizacion{
		Fecha:            time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC),
		Fuente:           "MAG",
		CategoriaCanonID: 1,
		PrecioPromedio:   precioProm,
		VolumenCabezas:   2500,
	}

	if err := repo.UpsertCotizacion(context.Background(), nuevaCot); err != nil {
		t.Fatalf("error haciendo upsert: %v", err)
	}

	if capturedParams.Fuente != "MAG" {
		t.Errorf("fuente capturada incorrecta: %s", capturedParams.Fuente)
	}
	if !capturedParams.PrecioPromedio.Equal(precioProm) {
		t.Errorf("precio promedio capturado incorrecto: %s", capturedParams.PrecioPromedio)
	}
}

func TestAlertRepository_CheckAndRegister(t *testing.T) {
	t.Parallel()

	q := &mockQuerier{
		checkAlertaFn: func(ctx context.Context, arg db.CheckAlertaEmitidaParams) (bool, error) {
			return arg.CategoriaCanonID == 1, nil
		},
		registrarAlertaFn: func(ctx context.Context, arg db.RegistrarAlertaParams) (db.AlertasLog, error) {
			return db.AlertasLog{
				TipoAlerta: arg.TipoAlerta,
				Canal:      arg.Canal,
				Mensaje:    arg.Mensaje,
			}, nil
		},
	}

	repo := database.NewAlertRepository(q)

	emitida, err := repo.CheckAlertaEmitida(context.Background(), time.Now(), 1, domain.AlertaVariacionDiaria, domain.CanalTelegram)
	if err != nil {
		t.Fatalf("error inesperado en CheckAlertaEmitida: %v", err)
	}
	if !emitida {
		t.Errorf("se esperaba que la alerta estuviera emitida")
	}

	alertaValida := domain.Alerta{
		Fecha:            time.Now(),
		CategoriaCanonID: 1,
		TipoAlerta:       domain.AlertaVariacionDiaria,
		Canal:            domain.CanalTelegram,
		Mensaje:          "Suba del 5.2% en Novillitos MAG",
	}

	if err := repo.RegistrarAlerta(context.Background(), alertaValida); err != nil {
		t.Fatalf("error registrando alerta: %v", err)
	}
}

