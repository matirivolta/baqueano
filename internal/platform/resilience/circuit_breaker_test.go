package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/matirivolta/baqueano/internal/platform/resilience"
)

func TestCircuitBreaker_Lifecycle(t *testing.T) {
	t.Parallel()

	cfg := resilience.Config{
		MaxFailures: 2,
		Timeout:     50 * time.Millisecond,
	}
	cb := resilience.NewCircuitBreaker(cfg)

	if cb.State() != resilience.StateClosed {
		t.Fatalf("esperado CLOSED, obtenido %s", cb.State())
	}

	testErr := errors.New("error de conexión con mercado")

	// 1er fallo
	_ = cb.Execute(context.Background(), func() error {
		return testErr
	})
	if cb.State() != resilience.StateClosed {
		t.Fatalf("con 1 fallo debe seguir CLOSED, obtenido %s", cb.State())
	}

	// 2do fallo -> debe abrirse
	_ = cb.Execute(context.Background(), func() error {
		return testErr
	})
	if cb.State() != resilience.StateOpen {
		t.Fatalf("con 2 fallos debe pasar a OPEN, obtenido %s", cb.State())
	}

	// Operación inmediata bloqueada por CircuitBreaker
	blockedErr := cb.Execute(context.Background(), func() error {
		return nil
	})
	if blockedErr != resilience.ErrCircuitOpen {
		t.Fatalf("se esperaba ErrCircuitOpen, obtenido %v", blockedErr)
	}

	// Esperar timeout para transición a Half-Open
	time.Sleep(60 * time.Millisecond)
	if cb.State() != resilience.StateHalfOpen {
		t.Fatalf("tras timeout debe estar HALF_OPEN, obtenido %s", cb.State())
	}

	// Operación exitosa restablece a Closed
	err := cb.Execute(context.Background(), func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("error inesperado en operación: %v", err)
	}

	if cb.State() != resilience.StateClosed {
		t.Fatalf("tras éxito debe volver a CLOSED, obtenido %s", cb.State())
	}
}

