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
	DefaultDeCampoURL = "https://decampoacampo.com"
)

// DeCampoScraper implementa la extracción para el marketplace digital De Campo a Campo.
type DeCampoScraper struct {
	baseURL string
	client  *http.Client
}

// NewDeCampoScraper construye un nuevo scraper para De Campo a Campo.
func NewDeCampoScraper(baseURL string, client *http.Client) *DeCampoScraper {
	if baseURL == "" {
		baseURL = DefaultDeCampoURL
	}
	if client == nil {
		client = NewHTTPClient(15 * time.Second)
	}
	return &DeCampoScraper{
		baseURL: baseURL,
		client:  client,
	}
}

// Name retorna el identificador de la fuente.
func (s *DeCampoScraper) Name() string {
	return domain.FuenteDeCampo
}

// Scrape realiza la petición HTTP a la grilla de precios de De Campo a Campo.
func (s *DeCampoScraper) Scrape(ctx context.Context) ([]collector.RawQuote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error construyendo request para De Campo a Campo: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error conectando a De Campo a Campo (%s): %w", s.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("De Campo a Campo devolvió status no exitoso: %d", resp.StatusCode)
	}

	return s.ParseHTML(resp.Body, time.Now())
}

// ParseHTML procesa las categorías y tramos de peso de De Campo a Campo.
func (s *DeCampoScraper) ParseHTML(r io.Reader, fecha time.Time) ([]collector.RawQuote, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML de De Campo a Campo: %w", err)
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
			// Categoría | Mínimo | Máximo | Promedio
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
			Fuente:         domain.FuenteDeCampo,
			Fecha:          fecha,
			TextoOrigen:    categoriaTexto,
			PrecioMin:      pMin,
			PrecioMax:      pMax,
			PrecioPromedio: promDecimal,
			VolumenCabezas: cabezas,
			Metadata: map[string]interface{}{
				"origen": "De Campo a Campo Digital",
			},
		})
	})

	return quotes, nil
}

