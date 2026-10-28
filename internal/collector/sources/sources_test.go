package sources_test

import (
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/collector/sources"
	"github.com/matirivolta/baqueano/internal/domain"
)

func TestParseArgentineDecimal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "Precio estándar con pesos y separadores",
			input:    "$ 2.950,50",
			expected: "2950.50",
		},
		{
			name:     "Precio entero con separador de miles",
			input:    "3.200",
			expected: "3200",
		},
		{
			name:     "Precio sin separadores",
			input:    "1450,75",
			expected: "1450.75",
		},
		{
			name:        "Sin cotización s/c",
			input:       "s/c",
			expectError: true,
		},
		{
			name:        "Guión vacío",
			input:       " - ",
			expectError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d, err := sources.ParseArgentineDecimal(tt.input)
			if tt.expectError {
				if err == nil {
					t.Fatalf("se esperaba error para input %s", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("error inesperado parseando %s: %v", tt.input, err)
			}

			expectedDec, _ := decimal.NewFromString(tt.expected)
			if !d.Equal(expectedDec) {
				t.Errorf("esperado %s, obtenido %s", expectedDec, d)
			}
		})
	}
}

func TestMAGScraper_ParseHTML(t *testing.T) {
	t.Parallel()

	htmlFixture := `
	<html>
	<body>
		<table>
			<tr>
				<th>Categoría</th>
				<th>Cabezas</th>
				<th>Mínimo</th>
				<th>Máximo</th>
				<th>Promedio</th>
			</tr>
			<tr>
				<td>Novillitos</td>
				<td>1.450</td>
				<td>$ 2.800,00</td>
				<td>$ 3.050,00</td>
				<td>$ 2.940,50</td>
			</tr>
			<tr>
				<td>Vaquillonas</td>
				<td>980</td>
				<td>$ 2.750,00</td>
				<td>$ 2.980,00</td>
				<td>$ 2.870,00</td>
			</tr>
			<tr>
				<td>Toros</td>
				<td>120</td>
				<td>$ 1.900,00</td>
				<td>$ 2.200,00</td>
				<td>$ 2.050,00</td>
			</tr>
		</table>
	</body>
	</html>
	`

	scraper := sources.NewMAGScraper("", nil)
	quotes, err := scraper.ParseHTML(strings.NewReader(htmlFixture), time.Now())
	if err != nil {
		t.Fatalf("error parseando fixture MAG: %v", err)
	}

	if len(quotes) != 3 {
		t.Fatalf("se esperaban 3 cotizaciones, obtenidas %d", len(quotes))
	}

	novillito := quotes[0]
	if novillito.TextoOrigen != "Novillitos" {
		t.Errorf("esperado Novillitos, obtenido %s", novillito.TextoOrigen)
	}
	if novillito.Fuente != domain.FuenteMAG {
		t.Errorf("esperada fuente MAG, obtenida %s", novillito.Fuente)
	}
	if novillito.VolumenCabezas != 1450 {
		t.Errorf("esperadas 1450 cabezas, obtenido %d", novillito.VolumenCabezas)
	}

	expectedProm := decimal.NewFromFloat(2940.50)
	if !novillito.PrecioPromedio.Equal(expectedProm) {
		t.Errorf("precio promedio esperado %s, obtenido %s", expectedProm, novillito.PrecioPromedio)
	}
}

func TestRosganScraper_ParseHTML(t *testing.T) {
	t.Parallel()

	htmlFixture := `
	<html>
	<body>
		<table>
			<tr>
				<td>Categoría</td>
				<td>Mínimo</td>
				<td>Máximo</td>
				<td>Promedio</td>
				<td>Cabezas</td>
			</tr>
			<tr>
				<td>Terneros</td>
				<td>3.100,00</td>
				<td>3.450,00</td>
				<td>3.280,50</td>
				<td>2.300</td>
			</tr>
			<tr>
				<td>Terneras</td>
				<td>2.900,00</td>
				<td>3.200,00</td>
				<td>3.050,00</td>
				<td>1.850</td>
			</tr>
		</table>
	</body>
	</html>
	`

	scraper := sources.NewRosganScraper("", nil)
	quotes, err := scraper.ParseHTML(strings.NewReader(htmlFixture), time.Now())
	if err != nil {
		t.Fatalf("error parseando fixture Rosgan: %v", err)
	}

	if len(quotes) != 2 {
		t.Fatalf("se esperaban 2 cotizaciones, obtenidas %d", len(quotes))
	}

	terneros := quotes[0]
	if terneros.TextoOrigen != "Terneros" {
		t.Errorf("esperado Terneros, obtenido %s", terneros.TextoOrigen)
	}
	if terneros.Fuente != domain.FuenteRosgan {
		t.Errorf("esperada fuente ROSGAN, obtenida %s", terneros.Fuente)
	}
	expectedProm := decimal.NewFromFloat(3280.50)
	if !terneros.PrecioPromedio.Equal(expectedProm) {
		t.Errorf("esperado %s, obtenido %s", expectedProm, terneros.PrecioPromedio)
	}
}

func TestDeCampoScraper_ParseHTML(t *testing.T) {
	t.Parallel()

	htmlFixture := `
	<html>
	<body>
		<table>
			<tr>
				<td>Categoría</td>
				<td>Mínimo</td>
				<td>Máximo</td>
				<td>Promedio</td>
				<td>Cabezas</td>
			</tr>
			<tr>
				<td>Terneros 160-180kg</td>
				<td>3.150,00</td>
				<td>3.400,00</td>
				<td>3.260,00</td>
				<td>540</td>
			</tr>
			<tr>
				<td>Novillitos recría</td>
				<td>2.700,00</td>
				<td>2.950,00</td>
				<td>2.840,00</td>
				<td>320</td>
			</tr>
		</table>
	</body>
	</html>
	`

	scraper := sources.NewDeCampoScraper("", nil)
	quotes, err := scraper.ParseHTML(strings.NewReader(htmlFixture), time.Now())
	if err != nil {
		t.Fatalf("error parseando fixture De Campo a Campo: %v", err)
	}

	if len(quotes) != 2 {
		t.Fatalf("se esperaban 2 cotizaciones, obtenidas %d", len(quotes))
	}

	if quotes[0].Fuente != domain.FuenteDeCampo {
		t.Errorf("esperada fuente DECAMPO, obtenida %s", quotes[0].Fuente)
	}
	if quotes[0].TextoOrigen != "Terneros 160-180kg" {
		t.Errorf("esperado Terneros 160-180kg, obtenido %s", quotes[0].TextoOrigen)
	}
}

func TestTattersallScraper_ParseHTML(t *testing.T) {
	t.Parallel()

	htmlFixture := `
	<html>
	<body>
		<table>
			<tr>
				<td>Categoría</td>
				<td>Mínimo</td>
				<td>Máximo</td>
				<td>Promedio</td>
				<td>Cabezas</td>
			</tr>
			<tr>
				<td>Novillos Especiales</td>
				<td>2.850,00</td>
				<td>3.100,00</td>
				<td>2.980,00</td>
				<td>650</td>
			</tr>
		</table>
	</body>
	</html>
	`

	scraper := sources.NewTattersallScraper("", nil)
	quotes, err := scraper.ParseHTML(strings.NewReader(htmlFixture), time.Now())
	if err != nil {
		t.Fatalf("error parseando fixture Tattersall: %v", err)
	}

	if len(quotes) != 1 {
		t.Fatalf("se esperaba 1 cotización, obtenidas %d", len(quotes))
	}

	if quotes[0].Fuente != domain.FuenteTattersall {
		t.Errorf("esperada fuente TATTERSALL, obtenida %s", quotes[0].Fuente)
	}
	if quotes[0].TextoOrigen != "Novillos Especiales" {
		t.Errorf("esperado Novillos Especiales, obtenido %s", quotes[0].TextoOrigen)
	}
}

