package collector_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/collector"
	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/platform/database"
)

type mockACLRepo struct {
	mapeos        map[string]*database.MapeoFuenteResult
	capturedUpsert []struct {
		fuente      string
		textoOrigen string
		catID       *int32
		verificado  bool
	}
}

func newMockACLRepo() *mockACLRepo {
	return &mockACLRepo{
		mapeos: make(map[string]*database.MapeoFuenteResult),
	}
}

func (m *mockACLRepo) GetMapeo(ctx context.Context, fuente, textoOrigen string) (*database.MapeoFuenteResult, error) {
	key := fuente + ":" + textoOrigen
	return m.mapeos[key], nil
}

func (m *mockACLRepo) UpsertMapeo(ctx context.Context, fuente, textoOrigen string, categoriaID *int32, verificado bool) error {
	m.capturedUpsert = append(m.capturedUpsert, struct {
		fuente      string
		textoOrigen string
		catID       *int32
		verificado  bool
	}{
		fuente:      fuente,
		textoOrigen: textoOrigen,
		catID:       categoriaID,
		verificado:  verificado,
	})

	key := fuente + ":" + textoOrigen
	m.mapeos[key] = &database.MapeoFuenteResult{
		Fuente:              fuente,
		TextoOrigen:         textoOrigen,
		CategoriaCanonID:    categoriaID,
		VerificadoPorHumano: verificado,
	}
	return nil
}

func TestNormalizer_MappedCategory(t *testing.T) {
	t.Parallel()

	repo := newMockACLRepo()
	catID := int32(2)
	repo.mapeos["MAG:Novillitos"] = &database.MapeoFuenteResult{
		ID:                  1,
		Fuente:              "MAG",
		TextoOrigen:         "Novillitos",
		CategoriaCanonID:    &catID,
		VerificadoPorHumano: true,
	}

	norm := collector.NewNormalizer(repo)

	prom := decimal.NewFromFloat(2940.50)
	raw := collector.RawQuote{
		Fuente:         domain.FuenteMAG,
		Fecha:          time.Now(),
		TextoOrigen:    "Novillitos",
		PrecioPromedio: prom,
		VolumenCabezas: 1200,
	}

	cot, err := norm.Normalize(context.Background(), raw)
	if err != nil {
		t.Fatalf("error inesperado en normalización: %v", err)
	}

	if cot.CategoriaCanonID != 2 {
		t.Errorf("esperado CategoriaCanonID 2, obtenido %d", cot.CategoriaCanonID)
	}
	if !cot.PrecioPromedio.Equal(prom) {
		t.Errorf("esperado precio promedio %s, obtenido %s", prom, cot.PrecioPromedio)
	}
}

func TestNormalizer_UnknownCategoryRecordedAsPending(t *testing.T) {
	t.Parallel()

	repo := newMockACLRepo()
	norm := collector.NewNormalizer(repo)

	raw := collector.RawQuote{
		Fuente:         domain.FuenteMAG,
		Fecha:          time.Now(),
		TextoOrigen:    "Categoría Totalmente Desconocida",
		PrecioPromedio: decimal.NewFromInt(3000),
	}

	_, err := norm.Normalize(context.Background(), raw)
	if err == nil {
		t.Fatalf("se esperaba error de mapeo pendiente")
	}
	if !errors.Is(err, collector.ErrMapeoPendiente) {
		t.Errorf("se esperaba error que envuelva ErrMapeoPendiente, obtenido: %v", err)
	}

	// Verificar que el término fue guardado como pendiente para revisión humana
	if len(repo.capturedUpsert) != 1 {
		t.Fatalf("se esperaba 1 registro capturado en ACL, obtenidos %d", len(repo.capturedUpsert))
	}

	capt := repo.capturedUpsert[0]
	if capt.fuente != domain.FuenteMAG {
		t.Errorf("esperada fuente MAG, obtenida %s", capt.fuente)
	}
	if capt.textoOrigen != "Categoría Totalmente Desconocida" {
		t.Errorf("texto capturado incorrecto: %s", capt.textoOrigen)
	}
	if capt.catID != nil {
		t.Errorf("el catID debe ser nil (pendiente de homologación)")
	}
	if capt.verificado != false {
		t.Errorf("verificado debe ser false")
	}
}

