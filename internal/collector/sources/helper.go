package sources

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const DefaultUserAgent = "BaqueanoBot/1.0 (+https://baqueano.agro.ar; cotizador ganadero)"

// NewHTTPClient retorna un cliente HTTP configurado con timeouts prudentes para scraping agropecuario.
func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &http.Client{
		Timeout: timeout,
	}
}

// ParseArgentineDecimal parsea números en formato monetario argentino ("$ 2.950,50" -> 2950.50).
func ParseArgentineDecimal(raw string) (decimal.Decimal, error) {
	limpio := strings.TrimSpace(raw)
	limpio = strings.ReplaceAll(limpio, "$", "")
	limpio = strings.ReplaceAll(limpio, " ", "")
	limpio = strings.ReplaceAll(limpio, ".", "")      // Eliminar separador de miles
	limpio = strings.ReplaceAll(limpio, ",", ".")      // Convertir coma decimal a punto

	if limpio == "" || limpio == "-" || limpio == "s/c" {
		return decimal.Zero, fmt.Errorf("valor no numérico o sin cotización: %s", raw)
	}

	d, err := decimal.NewFromString(limpio)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error parseando decimal %s: %w", raw, err)
	}

	return d, nil
}

