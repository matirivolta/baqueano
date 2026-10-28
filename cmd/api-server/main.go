package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/matirivolta/baqueano/api/rest"
	"github.com/matirivolta/baqueano/internal/db"
	"github.com/matirivolta/baqueano/internal/platform/database"
	"github.com/matirivolta/baqueano/internal/quotation"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("Iniciando Baqueano API Server...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Cargar configuración de entorno
	port := getEnv("PORT", getEnv("HTTP_PORT", "8080"))
	dbURL := getEnv("DATABASE_URL", "postgres://cotizador_app:cotizador_secret_pwd@localhost:5432/cotizador_db?sslmode=disable")

	// 2. Conectar a PostgreSQL (con degradación agraciada si la base aún no inició)
	var (
		quotRepo quotation.QuotationRepository
		catRepo  quotation.CategoriaRepository
		dbPing   rest.Pinger
	)

	dbPool, err := database.NewPool(ctx, database.DefaultConfig(dbURL))
	if err != nil {
		logger.Warn("PostgreSQL no está disponible al inicio. Iniciando API en modo resiliencia/offline.", "error", err)
	} else {
		defer dbPool.Close()
		logger.Info("Conexión con PostgreSQL establecida con éxito")
		queries := db.New(dbPool)
		quotRepo = database.NewQuotationRepository(queries)
		catRepo = database.NewCategoriaRepository(queries)
		dbPing = dbPool
	}

	// 3. Ensamblar repositorios y casos de uso de dominio
	svc := quotation.NewService(quotRepo, catRepo, nil)

	// 4. Configurar Handler y Router HTTP
	handler := rest.NewHandler(svc, catRepo, dbPing)
	router := rest.NewRouter(handler, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Iniciar servidor HTTP en segundo plano
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("API Server escuchando", "port", port, "url", fmt.Sprintf("http://localhost:%s", port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// 6. Esperar señal de terminación o error crítico
	select {
	case err := <-serverErr:
		logger.Error("Fallo crítico en el servidor HTTP", "error", err)
	case <-ctx.Done():
		logger.Info("Recibida señal de apagado, ejecutando graceful shutdown...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error durante el apagado del servidor", "error", err)
		} else {
			logger.Info("Servidor HTTP detenido correctamente.")
		}
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

