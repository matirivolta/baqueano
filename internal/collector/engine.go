package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/matirivolta/baqueano/internal/domain"
	"github.com/matirivolta/baqueano/internal/platform/eventbus"
	"github.com/matirivolta/baqueano/internal/platform/resilience"
	"github.com/matirivolta/baqueano/internal/quotation"
)

// DefaultEngine implementa la orquestación de ingesta concurrente con Circuit Breakers y Anti-Corruption Layer.
type DefaultEngine struct {
	mu              sync.RWMutex
	scrapers        []SourceScraper
	breakers        map[string]*resilience.CircuitBreaker
	normalizer      Normalizer
	quotationRepo   quotation.QuotationRepository
	producer        eventbus.Producer
	breakerConfig   resilience.Config
}

// NewEngine construye una nueva instancia del motor de scraping e ingesta.
func NewEngine(
	normalizer Normalizer,
	repo quotation.QuotationRepository,
	producer eventbus.Producer,
	breakerCfg resilience.Config,
) *DefaultEngine {
	return &DefaultEngine{
		scrapers:      make([]SourceScraper, 0),
		breakers:      make(map[string]*resilience.CircuitBreaker),
		normalizer:    normalizer,
		quotationRepo: repo,
		producer:      producer,
		breakerConfig: breakerCfg,
	}
}

// RegisterScraper añade una estrategia de mercado al motor y le asigna su propio Circuit Breaker aislado.
func (e *DefaultEngine) RegisterScraper(scraper SourceScraper) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.scrapers = append(e.scrapers, scraper)
	if _, exists := e.breakers[scraper.Name()]; !exists {
		e.breakers[scraper.Name()] = resilience.NewCircuitBreaker(e.breakerConfig)
	}
}

// Run dispara la ronda de extracción concurrente sobre todas las fuentes configuradas.
func (e *DefaultEngine) Run(ctx context.Context) ([]domain.Cotizacion, error) {
	e.mu.RLock()
	scrapers := make([]SourceScraper, len(e.scrapers))
	copy(scrapers, e.scrapers)
	e.mu.RUnlock()

	var wg sync.WaitGroup
	type scrapeResult struct {
		fuente string
		quotes []RawQuote
		err    error
	}

	resultsChan := make(chan scrapeResult, len(scrapers))

	for _, s := range scrapers {
		s := s
		wg.Add(1)

		go func() {
			defer wg.Done()

			e.mu.RLock()
			cb := e.breakers[s.Name()]
			e.mu.RUnlock()

			var raw []RawQuote
			err := cb.Execute(ctx, func() error {
				var scrapeErr error
				raw, scrapeErr = s.Scrape(ctx)
				return scrapeErr
			})

			resultsChan <- scrapeResult{
				fuente: s.Name(),
				quotes: raw,
				err:    err,
			}
		}()
	}

	wg.Wait()
	close(resultsChan)

	var procesadas []domain.Cotizacion
	var collectorErrors []error

	for res := range resultsChan {
		if res.err != nil {
			collectorErrors = append(collectorErrors, fmt.Errorf("fuente %s falló: %w", res.fuente, res.err))
			continue
		}

		for _, raw := range res.quotes {
			cot, err := e.normalizer.Normalize(ctx, raw)
			if err != nil {
				// Término nuevo o error de normalización: continúa procesando las demás
				continue
			}

			// Persistir cotización normalizada mediante sqlc
			if e.quotationRepo != nil {
				if err := e.quotationRepo.UpsertCotizacion(ctx, *cot); err != nil {
					collectorErrors = append(collectorErrors, fmt.Errorf("error guardando cotización (%s: %d): %w", cot.Fuente, cot.CategoriaCanonID, err))
					continue
				}
			}

			// Publicar evento en Redis Streams si el bus está habilitado
			if e.producer != nil {
				evt := domain.CotizacionIngeridaEvent{
					EventID:          fmt.Sprintf("evt-%s-%d-%d", cot.Fuente, cot.CategoriaCanonID, time.Now().UnixNano()),
					CotizacionID:     cot.ID,
					Fecha:            cot.Fecha,
					Fuente:           cot.Fuente,
					CategoriaCanonID: cot.CategoriaCanonID,
					CategoriaCodigo:  cot.CategoriaCodigo,
					PrecioPromedio:   cot.PrecioPromedio,
					PrecioMin:        cot.PrecioMin,
					PrecioMax:        cot.PrecioMax,
					VolumenCabezas:   cot.VolumenCabezas,
					Timestamp:        time.Now(),
				}
				_, _ = e.producer.PublishCotizacion(ctx, evt)
			}

			procesadas = append(procesadas, *cot)
		}
	}

	if len(collectorErrors) > 0 && len(procesadas) == 0 {
		return nil, fmt.Errorf("todas las fuentes fallaron en la ingesta: %v", collectorErrors)
	}

	return procesadas, nil
}

