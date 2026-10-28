package db_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/matirivolta/baqueano/internal/db"
)

func TestQuerierInterfaceImplementation(t *testing.T) {
	t.Parallel()

	// Verifica en tiempo de ejecución que *db.Queries satisface db.Querier
	var q interface{} = db.New(nil)
	if _, ok := q.(db.Querier); !ok {
		t.Fatalf("expected *db.Queries to implement db.Querier")
	}
}

func TestDecimalPrecisionForLivestockQuotes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		priceStr string
		expected string
	}{
		{
			name:     "Precio novillo consumo MAG",
			priceStr: "2850.50",
			expected: "2850.50",
		},
		{
			name:     "Precio ternero invernada Rosgan",
			priceStr: "3200.75",
			expected: "3200.75",
		},
		{
			name:     "Vaca conserva inferior",
			priceStr: "1450.00",
			expected: "1450.00",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d, err := decimal.NewFromString(tt.priceStr)
			if err != nil {
				t.Fatalf("unexpected error parsing decimal: %v", err)
			}

			if d.StringFixed(2) != tt.expected {
				t.Errorf("got %s, want %s", d.StringFixed(2), tt.expected)
			}
		})
	}
}

func TestCotizacionParamsModel(t *testing.T) {
	t.Parallel()

	precioProm := decimal.NewFromFloat(2950.50)
	precioMin := decimal.NullDecimal{
		Decimal: decimal.NewFromFloat(2800.00),
		Valid:   true,
	}

	params := db.UpsertCotizacionParams{
		Fuente:           "MAG",
		CategoriaCanonID: 1,
		PrecioPromedio:   precioProm,
		PrecioMin:        precioMin,
		VolumenCabezas:   pgtype.Int4{Int32: 1250, Valid: true},
		Metadata:         []byte(`{"rematador": "Monasterio Tattersall"}`),
	}

	if params.Fuente != "MAG" {
		t.Errorf("expected fuente MAG, got %s", params.Fuente)
	}
	if !params.PrecioPromedio.Equal(precioProm) {
		t.Errorf("expected %s, got %s", precioProm, params.PrecioPromedio)
	}
	if !params.PrecioMin.Valid || !params.PrecioMin.Decimal.Equal(precioMin.Decimal) {
		t.Errorf("expected valid precioMin %s, got %v", precioMin.Decimal, params.PrecioMin)
	}
}

