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
	DefaultMAGURL = "https://www.mercadoagroganadero.com.ar"
)

// MAGScraper implementa el extractor de precios para el Mercado Agroganadero de Cañuelas.
type MAGScraper struct {
	baseURL string
	client  *http.Client
}

// NewMAGScraper construye un nuevo scraper para el MAG.
func NewMAGScraper(baseURL string, client *http.Client) *MAGScraper {
	if baseURL == "" {
		baseURL = DefaultMAGURL
	}
	if client == nil {
		client = NewHTTPClient(15 * time.Second)
	}
	return &MAGScraper{
		baseURL: baseURL,
		client:  client,
	}
}

// Name retorna el identificador canónico de la fuente.
func (s *MAGScraper) Name() string {
	return domain.FuenteMAG
}

// Scrape ejecuta la petición HTTP con cabeceras amigables y parsea la tabla de cotizaciones del MAG.
func (s *MAGScraper) Scrape(ctx context.Context) ([]collector.RawQuote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error construyendo request para MAG: %w", err)
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error en petición HTTP a MAG (%s): %w", s.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MAG devolvió status inesperado: %d", resp.StatusCode)
	}

	return s.ParseHTML(resp.Body, time.Now())
}

// ParseHTML analiza el documento HTML y extrae las filas de cotización por categoría.
func (s *MAGScraper) ParseHTML(r io.Reader, fecha time.Time) ([]collector.RawQuote, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("error parseando HTML de MAG: %w", err)
	}

	var quotes []collector.RawQuote

	// Busca tablas que contengan columnas de categorías y precios
	doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
		tds := s.Find("td")
		if tds.Length() < 4 {
			return
		}

		categoriaTexto := strings.TrimSpace(tds.Eq(0).Text())
		if categoriaTexto == "" || strings.EqualFold(categoriaTexto, "categoría") || strings.EqualFold(categoriaTexto, "categoria") {
			return
		}

		// Estructura habitual MAG: Categoría | Cabezas | Mínimo | Máximo | Promedio
		var cabezas int
		var minStr, maxStr, promStr string

		if tds.Length() >= 5 {
			cabTexto := strings.ReplaceAll(strings.TrimSpace(tds.Eq(1).Text()), ".", "")
			cabezas, _ = strconv.Atoi(cabTexto)
			minStr = tds.Eq(2).Text()
			maxStr = tds.Eq(3).Text()
			promStr = tds.Eq(4).Text()
		} else {
			// Alternativa 4 columnas: Categoría | Mínimo | Máximo | Promedio
			minStr = tds.Eq(1).Text()
			maxStr = tds.Eq(2).Text()
			promStr = tds.Eq(3).Text()
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
			Fuente:         domain.FuenteMAG,
			Fecha:          fecha,
			TextoOrigen:    categoriaTexto,
			PrecioMin:      pMin,
			PrecioMax:      pMax,
			PrecioPromedio: promDecimal,
			VolumenCabezas: cabezas,
			Metadata: map[string]interface{}{
				"origen": "MAG Cañuelas Oficial",
			},
		})
	})

	return quotes, nil
}

