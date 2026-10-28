package integration_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/db"
	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/platform/database"
	"github.com/matirivolta/baqueano/internal/platform/eventbus"
	"github.com/matirivolta/baqueano/internal/quotation"
)

// inMemoryDBStore simula el almacenamiento transaccional en PostgreSQL con sqlc Querier.
type inMemoryDBStore struct {
	db.Querier
	categorias    map[int32]db.CategoriasCanon
	categoriasCod map[string]db.CategoriasCanon
	mapeos        map[string]db.MapeosFuente // key: fuente + ":" + texto
	cotizaciones  map[string]db.Cotizacione  // key: fecha + ":" + fuente + ":" + catID
	alertasLog    map[string]db.AlertasLog   // key: fecha + ":" + catID + ":" + tipo + ":" + canal
}

func newInMemoryDBStore() *inMemoryDBStore {
	store := &inMemoryDBStore{
		categorias:    make(map[int32]db.CategoriasCanon),
		categoriasCod: make(map[string]db.CategoriasCanon),
		mapeos:        make(map[string]db.MapeosFuente),
		cotizaciones:  make(map[string]db.Cotizacione),
		alertasLog:    make(map[string]db.AlertasLog),
	}

	// Semilla inicial simulada de categorías canónicas
	novillito := db.CategoriasCanon{
		ID:            2,
		Codigo:        "NOVILLITO",
		NombreVisible: "Novillitos (Consumo)",
		TipoDestino:   "FAENA",
		PesoMinimoKg:  pgtype.Int4{Int32: 300, Valid: true},
		PesoMaximoKg:  pgtype.Int4{Int32: 390, Valid: true},
		Activo:        true,
	}
	store.categorias[2] = novillito
	store.categoriasCod["NOVILLITO"] = novillito

	// Semilla inicial de mapeos ACL
	store.mapeos["MAG:Novillitos"] = db.MapeosFuente{
		ID:                  1,
		Fuente:              "MAG",
		TextoOrigen:         "Novillitos",
		CategoriaCanonID:    pgtype.Int4{Int32: 2, Valid: true},
		VerificadoPorHumano: true,
	}

	return store
}

func (s *inMemoryDBStore) ListCategoriasCanon(ctx context.Context) ([]db.CategoriasCanon, error) {
	res := make([]db.CategoriasCanon, 0, len(s.categorias))
	for _, c := range s.categorias {
		res = append(res, c)
	}
	return res, nil
}

func (s *inMemoryDBStore) GetCategoriaCanonByCodigo(ctx context.Context, codigo string) (db.CategoriasCanon, error) {
	c, ok := s.categoriasCod[codigo]
	if !ok {
		return db.CategoriasCanon{}, database.ErrCategoriaNotFound
	}
	return c, nil
}

func (s *inMemoryDBStore) GetCategoriaCanonByID(ctx context.Context, id int32) (db.CategoriasCanon, error) {
	c, ok := s.categorias[id]
	if !ok {
		return db.CategoriasCanon{}, database.ErrCategoriaNotFound
	}
	return c, nil
}

func (s *inMemoryDBStore) GetMapeoFuente(ctx context.Context, arg db.GetMapeoFuenteParams) (db.MapeosFuente, error) {
	key := arg.Fuente + ":" + arg.TextoOrigen
	m, ok := s.mapeos[key]
	if !ok {
		return db.MapeosFuente{}, database.ErrCategoriaNotFound
	}
	return m, nil
}

func (s *inMemoryDBStore) UpsertCotizacion(ctx context.Context, arg db.UpsertCotizacionParams) (db.Cotizacione, error) {
	key := arg.Fecha.Time.Format("2006-01-02") + ":" + arg.Fuente + ":" + string(rune(arg.CategoriaCanonID))
	cot := db.Cotizacione{
		ID:               int64(len(s.cotizaciones) + 1),
		Fecha:            arg.Fecha,
		Fuente:           arg.Fuente,
		CategoriaCanonID: arg.CategoriaCanonID,
		PrecioMin:        arg.PrecioMin,
		PrecioMax:        arg.PrecioMax,
		PrecioPromedio:   arg.PrecioPromedio,
		VolumenCabezas:   arg.VolumenCabezas,
		Metadata:         arg.Metadata,
		CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	s.cotizaciones[key] = cot
	return cot, nil
}

func (s *inMemoryDBStore) GetCotizacionPrevia(ctx context.Context, arg db.GetCotizacionPreviaParams) (db.Cotizacione, error) {
	// Retornamos la cotización del día anterior simulada ($2.780,00)
	return db.Cotizacione{
		ID:               100,
		Fecha:            pgtype.Date{Time: arg.Fecha.Time.AddDate(0, 0, -1), Valid: true},
		Fuente:           arg.Fuente,
		CategoriaCanonID: arg.CategoriaCanonID,
		PrecioPromedio:   decimal.NewFromFloat(2780.00),
		VolumenCabezas:   pgtype.Int4{Int32: 1500, Valid: true},
	}, nil
}

func (s *inMemoryDBStore) CheckAlertaEmitida(ctx context.Context, arg db.CheckAlertaEmitidaParams) (bool, error) {
	key := arg.Fecha.Time.Format("2006-01-02") + ":" + string(rune(arg.CategoriaCanonID)) + ":" + arg.TipoAlerta + ":" + arg.Canal
	_, exists := s.alertasLog[key]
	return exists, nil
}

func (s *inMemoryDBStore) RegistrarAlerta(ctx context.Context, arg db.RegistrarAlertaParams) (db.AlertasLog, error) {
	key := arg.Fecha.Time.Format("2006-01-02") + ":" + string(rune(arg.CategoriaCanonID)) + ":" + arg.TipoAlerta + ":" + arg.Canal
	alerta := db.AlertasLog{
		Fecha:            arg.Fecha,
		CategoriaCanonID: arg.CategoriaCanonID,
		TipoAlerta:       arg.TipoAlerta,
		Canal:            arg.Canal,
		Mensaje:          arg.Mensaje,
		Metadata:         arg.Metadata,
		DisparadoAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	s.alertasLog[key] = alerta
	return alerta, nil
}

// TestEndToEndPipeline_SimulatedIntegration ejecuta el flujo completo de extremo a extremo:
// 1. ACL lookup (texto heterogéneo -> CategoriaCanon).
// 2. Persistencia de Cotización normalizada en PostgreSQL (sqlc adapter).
// 3. Emisión de Evento CotizacionIngeridaEvent al EventBus.
// 4. Recepción en Worker de Alertas, cálculo de variación diaria (Δ% >= 4%).
// 5. Garantía de Idempotencia al registrar en alertas_log.
// 6. Simulación de liquidación neta ganadera con rinde por kg para el productor.
func TestEndToEndPipeline_SimulatedIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newInMemoryDBStore()

	// Inicializar adaptadores de persistencia
	catRepo := database.NewCategoriaRepository(store)
	quotRepo := database.NewQuotationRepository(store)
	aclRepo := database.NewACLRepository(store)
	alertRepo := database.NewAlertRepository(store)

	// PASO 1: Anti-Corruption Layer (Scraper extrae "Novillitos" del MAG)
	textoScrapeado := "Novillitos"
	fuente := domain.FuenteMAG

	mapeo, err := aclRepo.GetMapeo(ctx, fuente, textoScrapeado)
	if err != nil {
		t.Fatalf("falló resolución en ACL: %v", err)
	}
	if mapeo.CategoriaCanonID == nil || *mapeo.CategoriaCanonID != 2 {
		t.Fatalf("categoría canon resuelta inválida: %+v", mapeo)
	}

	catCanon, err := catRepo.GetCategoriaByID(ctx, *mapeo.CategoriaCanonID)
	if err != nil {
		t.Fatalf("error obteniendo categoría canónica: %v", err)
	}
	if catCanon.Codigo != "NOVILLITO" {
		t.Errorf("esperado NOVILLITO, obtenido %s", catCanon.Codigo)
	}

	// PASO 2: Persistir Cotización con precisión decimal estricta
	fechaHoy := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	precioPromHoy := decimal.NewFromFloat(2920.50) // Suba frente a $2.780 de ayer (+5.05%)
	precioMinHoy := decimal.NewFromFloat(2850.00)
	precioMaxHoy := decimal.NewFromFloat(3000.00)

	cot := domain.Cotizacion{
		Fecha:            fechaHoy,
		Fuente:           fuente,
		CategoriaCanonID: catCanon.ID,
		CategoriaCodigo:  catCanon.Codigo,
		CategoriaNombre:  catCanon.NombreVisible,
		TipoDestino:      catCanon.TipoDestino,
		PrecioMin:        &precioMinHoy,
		PrecioMax:        &precioMaxHoy,
		PrecioPromedio:   precioPromHoy,
		VolumenCabezas:   1850,
		Metadata: map[string]interface{}{
			"mercado":   "MAG Cañuelas",
			"operador":  "Consignataria de Prueba",
			"condicion": "Plaza firme",
		},
	}

	if err := quotRepo.UpsertCotizacion(ctx, cot); err != nil {
		t.Fatalf("falló UpsertCotizacion: %v", err)
	}

	// PASO 3: Publicación al Event Bus (Redis Streams contract)
	if eventbus.DefaultStreamName != "cotizaciones:stream" {
		t.Fatalf("stream name mismatch: %s", eventbus.DefaultStreamName)
	}

	event := domain.CotizacionIngeridaEvent{
		EventID:          "evt-integration-001",
		CotizacionID:     1,
		Fecha:            fechaHoy,
		Fuente:           fuente,
		CategoriaCanonID: catCanon.ID,
		CategoriaCodigo:  catCanon.Codigo,
		PrecioPromedio:   precioPromHoy,
		PrecioMin:        &precioMinHoy,
		PrecioMax:        &precioMaxHoy,
		VolumenCabezas:   1850,
		Timestamp:        time.Now(),
	}

	// Verificamos serialización del contrato Redis Streams
	eventBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("error serializando evento para Redis Streams: %v", err)
	}
	if len(eventBytes) == 0 {
		t.Fatalf("payload vacío para Redis Streams")
	}

	// PASO 4: Motor de Alertas (Subscriber procesa el evento)
	cotPrevia, err := quotRepo.GetCotizacionPrevia(ctx, event.CategoriaCanonID, event.Fuente, event.Fecha)
	if err != nil {
		t.Fatalf("error obteniendo cotización previa: %v", err)
	}
	if cotPrevia == nil {
		t.Fatalf("se esperaba cotización previa registrada")
	}

	// Evaluamos regla: Variación diaria |Δ%| >= 4%
	criterio := domain.CriterioVariacionDiaria{
		UmbralPorcentaje: decimal.NewFromInt(4), // 4%
	}
	cumpleAlerta, variacionPct := criterio.Evaluar(cotPrevia.PrecioPromedio, event.PrecioPromedio)
	if !cumpleAlerta {
		t.Fatalf("se esperaba que cumpliera umbral de alerta (obtenido Δ%%: %s)", variacionPct)
	}

	// PASO 5: Garantía de Idempotencia y Auditoría
	tipoAlerta := domain.AlertaVariacionDiaria
	canal := domain.CanalTelegram

	// Verificación 1: Primera pasada (no debe estar emitida)
	yaEmitida, err := alertRepo.CheckAlertaEmitida(ctx, event.Fecha, event.CategoriaCanonID, tipoAlerta, canal)
	if err != nil {
		t.Fatalf("error verificando idempotencia: %v", err)
	}
	if yaEmitida {
		t.Fatalf("alerta no debería estar emitida aún")
	}

	// Despacho y registro en alertas_log
	alerta := domain.Alerta{
		Fecha:            event.Fecha,
		CategoriaCanonID: event.CategoriaCanonID,
		TipoAlerta:       tipoAlerta,
		Canal:            canal,
		Mensaje:          "Suba en Novillito MAG: " + variacionPct.StringFixed(2) + "%",
		Metadata: map[string]interface{}{
			"precio_previo": cotPrevia.PrecioPromedio.StringFixed(2),
			"precio_nuevo":  event.PrecioPromedio.StringFixed(2),
		},
	}
	if err := alertRepo.RegistrarAlerta(ctx, alerta); err != nil {
		t.Fatalf("error registrando alerta: %v", err)
	}

	// Verificación 2: Segunda pasada (idempotencia en reintento)
	yaEmitidaSegundaVez, err := alertRepo.CheckAlertaEmitida(ctx, event.Fecha, event.CategoriaCanonID, tipoAlerta, canal)
	if err != nil {
		t.Fatalf("error verificando idempotencia en 2da pasada: %v", err)
	}
	if !yaEmitidaSegundaVez {
		t.Fatalf("idempotencia falló: la alerta debería figurar como YA EMITIDA")
	}

	// PASO 6: Productor ganadero utiliza la cotización ingerida para simular su liquidación neta
	calc := quotation.NewCalculator()
	liqParams := domain.ParametrosLiquidacion{
		Cabezas:               45,
		PesoBrutoKg:           decimal.NewFromInt(360), // 360 kg por novillito
		EsPesoIndividual:      true,
		PrecioPorKg:           event.PrecioPromedio, // Precio observado en MAG: $2.920,50
		PorcentajeDesbaste:    decimal.NewFromInt(3), // 3%
		PorcentajeComision:    decimal.NewFromInt(2), // 2%
		PorcentajeIVAComision: decimal.NewFromInt(21),
		CostoFleteTotal:       decimal.NewFromInt(350000), // Flete jaula a frigorífico
		OtrosCostosFijos:      decimal.NewFromInt(45000),  // Guías y balanza
		PlazoDias:             14,                         // 14 días
		TasaInteresAnual:      decimal.NewFromInt(40),     // 40% TNA
	}

	liq, err := calc.CalcularLiquidacion(liqParams)
	if err != nil {
		t.Fatalf("error en cálculo de liquidación: %v", err)
	}

	// Verificaciones de consistencia financiera:
	// Kilos brutos = 45 * 360 = 16.200 kg
	if !liq.KilosBrutos.Equal(decimal.NewFromInt(16200)) {
		t.Errorf("kilos brutos incorrectos: %s", liq.KilosBrutos)
	}
	// Kilos netos = 16.200 - 486 (3%) = 15.714 kg
	if !liq.KilosNetos.Equal(decimal.NewFromInt(15714)) {
		t.Errorf("kilos netos incorrectos: %s", liq.KilosNetos)
	}
	// Importe Bruto = 15.714 * $2.920,50 = $45.892.737,00
	expectedBruto := decimal.NewFromFloat(45892737.00)
	if !liq.ImporteBruto.Equal(expectedBruto) {
		t.Errorf("importe bruto incorrecto: esperado %s, obtenido %s", expectedBruto, liq.ImporteBruto)
	}
	// Cobro neto efectivo debe ser menor al importe bruto y estrictamente positivo
	if liq.CobroNetoEfectivo.GreaterThan(liq.ImporteBruto) {
		t.Errorf("el cobro neto no puede superar el importe bruto")
	}
	if liq.CobroNetoEfectivo.LessThanOrEqual(decimal.Zero) {
		t.Errorf("cobro neto no positivo: %s", liq.CobroNetoEfectivo)
	}
	// Rinde neto por kilo en pie de campo
	if liq.PrecioNetoPorKiloBruto.GreaterThan(event.PrecioPromedio) {
		t.Errorf("el rinde de bolsillo por kilo bruto no puede superar el precio bruto de plaza")
	}
}
