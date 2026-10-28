package sources

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/collector"
	"github.com/matirivolta/baqueano/internal/domain"
)

const (
	DefaultRosganURL = "https://www.rosgan.com.ar"
)

// RosganScraper implementa la extracción de cotizaciones de invernada y cría de Rosgan.
type RosganScraper struct {
	baseURL string
	client  *http.Client
}

// NewRosganScraper crea una nueva estrategia de scraping para Rosgan.
func NewRosganScraper(baseURL string, client *http.Client) *RosganScraper {
	if baseURL == "" {
		baseURL = DefaultRosganURL
	}
	if client == nil {
		client = NewHTTPClient(15 * time.Second)
	}
	return &RosganScraper{
		baseURL: baseURL,
		client:  client,
	}
}

// Name retorna el identificador de la fuente Rosgan.
func (s *RosganScraper) Name() string {
	return domain.FuenteRosgan
}

// Scrape ejecuta la solicitud web a Rosgan y procesa la grilla de precios de invernada.
func (s *RosganScraper) Scrape(ctx context.Context) ([]collector.RawQuote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request para Rosgan: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error conectando a Rosgan (%s): %w", s.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Rosgan devolvió status no exitoso: %d", resp.StatusCode)
	}

	return s.ParseHTML(resp.Body, time.Now())
}

// ParseHTML analiza la grilla de categorías de invernada y cría de Rosgan.
func (s *RosganScraper) ParseHTML(r io.Reader, fecha time.Time) ([]collector.RawQuote, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML de Rosgan: %w", err)
	}

	var quotes []collector.RawQuote

	doc.Find("table tr").Each(func(i int, sel *goquery.Selection) {
		tds := sel.Find("td")
		if tds.Length() < 2 {
			return
		}

		categoriaTexto := strings.TrimSpace(tds.Eq(0).Text())
		if categoriaTexto == "" || strings.EqualFold(categoriaTexto, "categoría") || strings.EqualFold(categoriaTexto, "categoria") {
			return
		}

		// Grilla Rosgan habitual: Categoría | Precio Promedio | (Opcional: Mín | Máx | Cabezas)
		var promStr, minStr, maxStr string
		var cabezas int

		if tds.Length() >= 5 {
			minStr = tds.Eq(1).Text()
			maxStr = tds.Eq(2).Text()
			promStr = tds.Eq(3).Text()
			cabTexto := strings.ReplaceAll(strings.TrimSpace(tds.Eq(4).Text()), ".", "")
			cabezas, _ = strconv.Atoi(cabTexto)
		} else if tds.Length() >= 2 {
			promStr = tds.Eq(1).Text()
		}

		promDecimal, err := ParseArgentineDecimal(promStr)
		if err != nil || promDecimal.LessThanOrEqual(decimal.Zero) {
			return
		}

		var pMin, pMax *decimal.Decimal
		if minDec, err := ParseArgentineDecimal(minStr); err == nil && minDec.GreaterThan(decimal.Zero) {
			pMin = &minDec
		}
		if maxDec, err := ParseArgentineDecimal(maxStr); err == nil && maxDec.GreaterThan(decimal.Zero) {
			pMax = &maxDec
		}

		quotes = append(quotes, collector.RawQuote{
			Fuente:         domain.FuenteRosgan,
			Fecha:          fecha,
			TextoOrigen:    categoriaTexto,
			PrecioMin:      pMin,
			PrecioMax:      pMax,
			PrecioPromedio: promDecimal,
			VolumenCabezas: cabezas,
			Metadata: map[string]interface{}{
				"origen": "Rosgan Remates",
			},
		})
	})

	return quotes, nil
}

