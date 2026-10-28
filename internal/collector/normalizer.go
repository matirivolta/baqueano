package collector

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/matirivolta/baqueano/internal/domain"
)

var (
	ErrTextoOrigenVacio = errors.New("el texto de origen está vacío")
	ErrMapeoPendiente   = errors.New("la categoría de origen no posee mapeo canónico verificado; registrado como pendiente")
)

// DefaultNormalizer implementa el Anti-Corruption Layer (ACL) desacoplando fuentes externas del modelo canónico.
type DefaultNormalizer struct {
	aclRepo ACLRepository
}

// NewNormalizer construye una nueva instancia del normalizador ACL.
func NewNormalizer(aclRepo ACLRepository) *DefaultNormalizer {
	return &DefaultNormalizer{aclRepo: aclRepo}
}

// Normalize traduce una cotización cruda a una canónica resolviendo su identificador en mapeos_fuente.
// Si el término es desconocido, lo persiste automáticamente con categoria_canon_id = NULL para revisión humana.
func (n *DefaultNormalizer) Normalize(ctx context.Context, raw RawQuote) (*domain.Cotizacion, error) {
	textoLimpio := strings.TrimSpace(raw.TextoOrigen)
	if textoLimpio == "" {
		return nil, ErrTextoOrigenVacio
	}

	mapeo, err := n.aclRepo.GetMapeo(ctx, raw.Fuente, textoLimpio)
	if err != nil {
		return nil, fmt.Errorf("error consultando mapeos_fuente en ACL: %w", err)
	}

	// Caso 1: Término no registrado previamente en el sistema
	if mapeo == nil {
		_ = n.aclRepo.UpsertMapeo(ctx, raw.Fuente, textoLimpio, nil, false)
		return nil, fmt.Errorf("%w: %s (%s)", ErrMapeoPendiente, textoLimpio, raw.Fuente)
	}

	// Caso 2: Término registrado pero aún no asignado a una categoría canónica por el operador
	if mapeo.CategoriaCanonID == nil {
		return nil, fmt.Errorf("%w: %s (%s)", ErrMapeoPendiente, textoLimpio, raw.Fuente)
	}

	fecha := raw.Fecha
	if fecha.IsZero() {
		fecha = time.Now().Truncate(24 * time.Hour)
	}

	cot := &domain.Cotizacion{
		Fecha:            fecha,
		Fuente:           raw.Fuente,
		CategoriaCanonID: *mapeo.CategoriaCanonID,
		PrecioMin:        raw.PrecioMin,
		PrecioMax:        raw.PrecioMax,
		PrecioPromedio:   raw.PrecioPromedio,
		VolumenCabezas:   raw.VolumenCabezas,
		Metadata:         raw.Metadata,
		CreatedAt:        time.Now(),
	}

	if err := cot.Validate(); err != nil {
		return nil, fmt.Errorf("la cotización normalizada violó invariantes de dominio: %w", err)
	}

	return cot, nil
}

