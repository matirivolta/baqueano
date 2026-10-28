package eventbus_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/platform/eventbus"
)

func TestCotizacionIngeridaEvent_Serialization(t *testing.T) {
	t.Parallel()

	min := decimal.NewFromInt(2800)
	max := decimal.NewFromInt(2950)
	prom := decimal.NewFromFloat(2880.50)

	evt := domain.CotizacionIngeridaEvent{
		EventID:          "evt-12345",
		CotizacionID:     99,
		Fecha:            time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC),
		Fuente:           domain.FuenteMAG,
		CategoriaCanonID: 1,
		CategoriaCodigo:  "NOVILLITO",
		PrecioPromedio:   prom,
		PrecioMin:        &min,
		PrecioMax:        &max,
		VolumenCabezas:   3200,
		Timestamp:        time.Now(),
	}

	bytes, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("error serializando evento: %v", err)
	}

	var deserialized domain.CotizacionIngeridaEvent
	if err := json.Unmarshal(bytes, &deserialized); err != nil {
		t.Fatalf("error deserializando evento: %v", err)
	}

	if deserialized.EventID != evt.EventID {
		t.Errorf("esperado EventID %s, obtenido %s", evt.EventID, deserialized.EventID)
	}
	if deserialized.CategoriaCodigo != "NOVILLITO" {
		t.Errorf("esperado CategoriaCodigo NOVILLITO, obtenido %s", deserialized.CategoriaCodigo)
	}
	if !deserialized.PrecioPromedio.Equal(prom) {
		t.Errorf("esperado precio promedio %s, obtenido %s", prom, deserialized.PrecioPromedio)
	}
	if deserialized.PrecioMin == nil || !deserialized.PrecioMin.Equal(min) {
		t.Errorf("esperado precio min %s, obtenido %v", min, deserialized.PrecioMin)
	}
}

func TestStreamConstants(t *testing.T) {
	t.Parallel()

	if eventbus.DefaultStreamName != "cotizaciones:stream" {
		t.Errorf("esperado stream cotizaciones:stream, obtenido %s", eventbus.DefaultStreamName)
	}
	if eventbus.DefaultGroupName != "alert-dispatcher-group" {
		t.Errorf("esperado grupo alert-dispatcher-group, obtenido %s", eventbus.DefaultGroupName)
	}
}

