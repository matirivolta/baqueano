package alert_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/alert"
	"github.com/matirivolta/baqueano/internal/domain"
)

type mockHistoricalRepo struct {
	cotizacionPrevia *domain.Cotizacion
	err              error
}

func (m *mockHistoricalRepo) GetCotizacionPrevia(ctx context.Context, categoriaID int32, fuente string, fecha time.Time) (*domain.Cotizacion, error) {
	return m.cotizacionPrevia, m.err
}

type mockAlertRepo struct {
	alertasEmitidas map[string]bool
	alertasLog      []domain.Alerta
}

func newMockAlertRepo() *mockAlertRepo {
	return &mockAlertRepo{
		alertasEmitidas: make(map[string]bool),
		alertasLog:      make([]domain.Alerta, 0),
	}
}

func (m *mockAlertRepo) CheckAlertaEmitida(ctx context.Context, fecha time.Time, categoriaID int32, tipo domain.TipoAlerta, canal domain.CanalNotificacion) (bool, error) {
	key := fmtKey(fecha, categoriaID, tipo, canal)
	return m.alertasEmitidas[key], nil
}

func (m *mockAlertRepo) RegistrarAlerta(ctx context.Context, alerta domain.Alerta) error {
	key := fmtKey(alerta.Fecha, alerta.CategoriaCanonID, alerta.TipoAlerta, alerta.Canal)
	m.alertasEmitidas[key] = true
	m.alertasLog = append(m.alertasLog, alerta)
	return nil
}

func fmtKey(fecha time.Time, categoriaID int32, tipo domain.TipoAlerta, canal domain.CanalNotificacion) string {
	return fecha.Format("2006-01-02") + ":" + string(rune(categoriaID)) + ":" + string(tipo) + ":" + string(canal)
}

type mockNotifier struct {
	canal           domain.CanalNotificacion
	mensajesEnviados []domain.Alerta
}

func (m *mockNotifier) Canal() domain.CanalNotificacion {
	return m.canal
}

func (m *mockNotifier) Send(ctx context.Context, alerta domain.Alerta) error {
	m.mensajesEnviados = append(m.mensajesEnviados, alerta)
	return nil
}

func TestEngine_EvaluateSignificantIncrease(t *testing.T) {
	t.Parallel()

	// Cotización previa: $2.800
	// Nueva cotización: $2.950 (+5.36% >= 4%)
	precioPrevio := decimal.NewFromFloat(2800.00)
	precioNuevo := decimal.NewFromFloat(2950.00)

	histRepo := &mockHistoricalRepo{
		cotizacionPrevia: &domain.Cotizacion{
			ID:               1,
			Fecha:            time.Now().AddDate(0, 0, -1),
			Fuente:           domain.FuenteMAG,
			CategoriaCanonID: 2,
			PrecioPromedio:   precioPrevio,
		},
	}
	alertRepo := newMockAlertRepo()
	notifier := &mockNotifier{canal: domain.CanalTelegram}

	engine := alert.NewEngine(alert.DefaultConfig(), histRepo, alertRepo)
	engine.RegisterNotifier(notifier)

	event := domain.CotizacionIngeridaEvent{
		EventID:          "evt-100",
		Fecha:            time.Now(),
		Fuente:           domain.FuenteMAG,
		CategoriaCanonID: 2,
		CategoriaCodigo:  "NOVILLITO",
		PrecioPromedio:   precioNuevo,
		VolumenCabezas:   1450,
	}

	err := engine.ProcessEvent(context.Background(), event)
	if err != nil {
		t.Fatalf("error procesando evento: %v", err)
	}

	// Verificar despacho a Telegram
	if len(notifier.mensajesEnviados) != 1 {
		t.Fatalf("se esperaba 1 mensaje enviado a Telegram, obtenidos %d", len(notifier.mensajesEnviados))
	}

	msg := notifier.mensajesEnviados[0].Mensaje
	if !strings.Contains(msg, "Suba Diaria") {
		t.Errorf("el mensaje no contiene 'Suba Diaria': %s", msg)
	}
	if !strings.Contains(msg, "NOVILLITO") {
		t.Errorf("el mensaje no contiene la categoría NOVILLITO: %s", msg)
	}

	// Verificar registro en BD
	if len(alertRepo.alertasLog) != 1 {
		t.Fatalf("se esperaba 1 alerta registrada en auditoría, obtenidas %d", len(alertRepo.alertasLog))
	}
}

func TestEngine_IdempotencyPreventsDuplicateDispatch(t *testing.T) {
	t.Parallel()

	precioPrevio := decimal.NewFromFloat(2800.00)
	precioNuevo := decimal.NewFromFloat(2950.00)

	histRepo := &mockHistoricalRepo{
		cotizacionPrevia: &domain.Cotizacion{
			PrecioPromedio: precioPrevio,
		},
	}
	alertRepo := newMockAlertRepo()
	notifier := &mockNotifier{canal: domain.CanalTelegram}

	engine := alert.NewEngine(alert.DefaultConfig(), histRepo, alertRepo)
	engine.RegisterNotifier(notifier)

	event := domain.CotizacionIngeridaEvent{
		Fecha:            time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC),
		Fuente:           domain.FuenteMAG,
		CategoriaCanonID: 2,
		CategoriaCodigo:  "NOVILLITO",
		PrecioPromedio:   precioNuevo,
	}

	// Primera ejecución: debe despachar
	if err := engine.ProcessEvent(context.Background(), event); err != nil {
		t.Fatalf("error en primera pasada: %v", err)
	}
	if len(notifier.mensajesEnviados) != 1 {
		t.Fatalf("esperado 1 mensaje enviado, obtenidos %d", len(notifier.mensajesEnviados))
	}

	// Segunda ejecución con evento idéntico (reintento de scraper): no debe duplicar despacho
	if err := engine.ProcessEvent(context.Background(), event); err != nil {
		t.Fatalf("error en segunda pasada: %v", err)
	}
	if len(notifier.mensajesEnviados) != 1 {
		t.Fatalf("la idempotencia falló: se duplicó el mensaje enviado a Telegram (%d)", len(notifier.mensajesEnviados))
	}
}

func TestEngine_SmallVariationIgnored(t *testing.T) {
	t.Parallel()

	// Variación leve: $2.800 a $2.830 (+1.07% < 4%)
	histRepo := &mockHistoricalRepo{
		cotizacionPrevia: &domain.Cotizacion{
			PrecioPromedio: decimal.NewFromFloat(2800.00),
		},
	}
	alertRepo := newMockAlertRepo()
	notifier := &mockNotifier{canal: domain.CanalTelegram}

	engine := alert.NewEngine(alert.DefaultConfig(), histRepo, alertRepo)
	engine.RegisterNotifier(notifier)

	event := domain.CotizacionIngeridaEvent{
		Fecha:            time.Now(),
		Fuente:           domain.FuenteMAG,
		CategoriaCanonID: 2,
		PrecioPromedio:   decimal.NewFromFloat(2830.00),
	}

	if err := engine.ProcessEvent(context.Background(), event); err != nil {
		t.Fatalf("error procesando evento: %v", err)
	}

	// No debe haber enviado alertas
	if len(notifier.mensajesEnviados) != 0 {
		t.Errorf("no se esperaba alerta por variación menor al 4%%, enviadas: %d", len(notifier.mensajesEnviados))
	}
}

