package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

type State string

const (
	StateClosed   State = "CLOSED"
	StateOpen     State = "OPEN"
	StateHalfOpen State = "HALF_OPEN"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker está ABIERTO: solicitudes bloqueadas preventivamente")
)

// Config define los umbrales de activación y recuperación del Circuit Breaker.
type Config struct {
	MaxFailures int           // Número de fallos consecutivos para abrir el circuito
	Timeout     time.Duration // Tiempo que permanece abierto antes de intentar medio-abrirse
}

// DefaultConfig retorna valores de resiliencia estándar para scrapers web.
func DefaultConfig() Config {
	return Config{
		MaxFailures: 5,
		Timeout:     30 * time.Second,
	}
}

// CircuitBreaker implementa el patrón de aislamiento de fallas para prevenir saturación de fuentes externas.
type CircuitBreaker struct {
	mu          sync.RWMutex
	state       State
	failures    int
	maxFailures int
	timeout     time.Duration
	lastFailure time.Time
}

// NewCircuitBreaker construye una nueva instancia del Circuit Breaker.
func NewCircuitBreaker(cfg Config) *CircuitBreaker {
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: cfg.MaxFailures,
		timeout:     cfg.Timeout,
	}
}

// State retorna el estado actual del disyuntor de forma concurrente y segura.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.currentState()
}

func (cb *CircuitBreaker) currentState() State {
	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.timeout {
		return StateHalfOpen
	}
	return cb.state
}

// Execute envuelve una operación con protección de circuito y timeout.
func (cb *CircuitBreaker) Execute(ctx context.Context, op func() error) error {
	cb.mu.Lock()
	st := cb.currentState()

	if st == StateOpen {
		cb.mu.Unlock()
		return ErrCircuitOpen
	}

	if st == StateHalfOpen {
		cb.state = StateHalfOpen
	}
	cb.mu.Unlock()

	// Ejecutar la operación externa
	err := op()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.maxFailures || cb.state == StateHalfOpen {
			cb.state = StateOpen
		}
		return err
	}

	// Éxito: restablecer contadores
	cb.failures = 0
	cb.state = StateClosed
	return nil
}

