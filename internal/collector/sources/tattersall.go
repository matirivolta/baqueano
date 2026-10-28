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
	DefaultTattersallURL = "https://monasterio-tattersall.com"
)

// TattersallScraper implementa la extracción para la consignataria Monasterio Tattersall.
type TattersallScraper struct {
	baseURL string
	client  *http.Client
}

// NewTattersallScraper construye una nueva estrategia de scraping para Tattersall.
func NewTattersallScraper(baseURL string, client *http.Client) *TattersallScraper {
	if baseURL == "" {
		baseURL = DefaultTattersallURL
	}
	if client == nil {
		client = NewHTTPClient(15 * time.Second)
	}
	return &TattersallScraper{
		baseURL: baseURL,
		client:  client,
	}
}

// Name retorna el identificador de la fuente.
func (s *TattersallScraper) Name() string {
	return domain.FuenteTattersall
}

// Scrape realiza la petición HTTP a la sección de cotizaciones de Tattersall.
func (s *TattersallScraper) Scrape(ctx context.Context) ([]collector.RawQuote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando request para Tattersall: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error conectando a Tattersall (%s): %w", s.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tattersall devolvió status no exitoso: %d", resp.StatusCode)
	}

	return s.ParseHTML(resp.Body, time.Now())
}

// ParseHTML analiza la grilla de cotizaciones de hacienda de Monasterio Tattersall.
func (s *TattersallScraper) ParseHTML(r io.Reader, fecha time.Time) ([]collector.RawQuote, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML de Tattersall: %w", err)
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

		var promStr, minStr, maxStr string
		var cabezas int

		if tds.Length() >= 4 {
			minStr = tds.Eq(1).Text()
			maxStr = tds.Eq(2).Text()
			promStr = tds.Eq(3).Text()
			if tds.Length() >= 5 {
				cabTexto := strings.ReplaceAll(strings.TrimSpace(tds.Eq(4).Text()), ".", "")
				cabezas, _ = strconv.Atoi(cabTexto)
			}
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
			Fuente:         domain.FuenteTattersall,
			Fecha:          fecha,
			TextoOrigen:    categoriaTexto,
			PrecioMin:      pMin,
			PrecioMax:      pMax,
			PrecioPromedio: promDecimal,
			VolumenCabezas: cabezas,
			Metadata: map[string]interface{}{
				"origen": "Monasterio Tattersall Remates",
			},
		})
	})

	return quotes, nil
}

