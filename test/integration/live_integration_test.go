//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/db"
	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/platform/database"
	"github.com/matirivolta/baqueano/internal/platform/eventbus"
)

func TestLivePostgreSQLIntegration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL no configurada; omitiendo test de integración con PostgreSQL real")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, database.DefaultConfig(dbURL))
	if err != nil {
		t.Fatalf("error conectando a PostgreSQL real: %v", err)
	}
	defer pool.Close()

	// Cargar DDL y semillas
	schemaSQL, err := os.ReadFile("../../sql/migrations/000001_init_schema.up.sql")
	if err != nil {
		t.Fatalf("error leyendo schema DDL: %v", err)
	}
	if _, err := pool.Exec(ctx, string(schemaSQL)); err != nil {
		t.Fatalf("error aplicando schema en PostgreSQL: %v", err)
	}

	seedsSQL, err := os.ReadFile("../../sql/seeds/baseline_data.sql")
	if err == nil {
		_, _ = pool.Exec(ctx, string(seedsSQL))
	}

	queries := db.New(pool)
	catRepo := database.NewCategoriaRepository(queries)
	quotRepo := database.NewQuotationRepository(queries)
	alertRepo := database.NewAlertRepository(queries)

	// 1. Verificar listado de categorías inicial
	cats, err := catRepo.ListCategorias(ctx)
	if err != nil {
		t.Fatalf("error listando categorías en DB real: %v", err)
	}
	if len(cats) == 0 {
		t.Fatalf("se esperaban categorías cargadas en baseline")
	}

	// 2. Insertar cotización y recuperar
	hoy := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	prom := decimal.NewFromFloat(2940.00)
	min := decimal.NewFromFloat(2850.00)
	max := decimal.NewFromFloat(3020.00)

	cot := domain.Cotizacion{
		Fecha:            hoy,
		Fuente:           domain.FuenteMAG,
		CategoriaCanonID: cats[0].ID,
		PrecioMin:        &min,
		PrecioMax:        &max,
		PrecioPromedio:   prom,
		VolumenCabezas:   2100,
	}

	if err := quotRepo.UpsertCotizacion(ctx, cot); err != nil {
		t.Fatalf("error en UpsertCotizacion real: %v", err)
	}

	ultimas, err := quotRepo.GetUltimasCotizaciones(ctx)
	if err != nil {
		t.Fatalf("error en GetUltimasCotizaciones: %v", err)
	}
	if len(ultimas) == 0 {
		t.Fatalf("se esperaba al menos 1 cotización guardada")
	}

	// 3. Probar idempotencia de alertas en PostgreSQL real
	alerta := domain.Alerta{
		Fecha:            hoy,
		CategoriaCanonID: cats[0].ID,
		TipoAlerta:       domain.AlertaVariacionDiaria,
		Canal:            domain.CanalTelegram,
		Mensaje:          "Suba en PostgreSQL real",
	}

	emitidaPre, err := alertRepo.CheckAlertaEmitida(ctx, hoy, cats[0].ID, alerta.TipoAlerta, alerta.Canal)
	if err != nil {
		t.Fatalf("error en CheckAlertaEmitida: %v", err)
	}
	if emitidaPre {
		t.Errorf("no debería estar emitida antes de registrar")
	}

	if err := alertRepo.RegistrarAlerta(ctx, alerta); err != nil {
		t.Fatalf("error registrando alerta: %v", err)
	}

	emitidaPost, err := alertRepo.CheckAlertaEmitida(ctx, hoy, cats[0].ID, alerta.TipoAlerta, alerta.Canal)
	if err != nil {
		t.Fatalf("error en CheckAlertaEmitida post: %v", err)
	}
	if !emitidaPost {
		t.Errorf("debería figurar como emitida tras registrar")
	}
}

func TestLiveRedisStreamsIntegration(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL no configurada; omitiendo test de integración con Redis real")
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		opts = &redis.Options{Addr: redisURL}
	}

	client := redis.NewClient(opts)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("error conectando a Redis real: %v", err)
	}

	bus := eventbus.NewRedisEventBus(client)

	testEvent := domain.CotizacionIngeridaEvent{
		EventID:          "live-evt-001",
		CotizacionID:     10,
		Fecha:            time.Now(),
		Fuente:           domain.FuenteMAG,
		CategoriaCanonID: 1,
		CategoriaCodigo:  "NOVILLITO",
		PrecioPromedio:   decimal.NewFromFloat(2900.00),
		Timestamp:        time.Now(),
	}

	msgID, err := bus.PublishCotizacion(ctx, testEvent)
	if err != nil {
		t.Fatalf("error publicando en Redis Streams: %v", err)
	}
	if msgID == "" {
		t.Errorf("ID de mensaje en Redis Streams vacío")
	}
}
