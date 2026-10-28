package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/matirivolta/baqueano/internal/collector"
	"github.com/matirivolta/baqueano/internal/collector/sources"
	"github.com/matirivolta/baqueano/internal/db"
	"github.com/matirivolta/baqueano/internal/platform/database"
	"github.com/matirivolta/baqueano/internal/platform/eventbus"
	"github.com/matirivolta/baqueano/internal/platform/resilience"
)

func main() {
	log.Println("[market-collector] Iniciando servicio recolector de cotizaciones ganaderas Baqueano...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Cargar configuración desde variables de entorno
	dbURL := getEnv("DATABASE_URL", "postgres://cotizador_app:cotizador_secret_pwd@localhost:5432/cotizador_db?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "localhost:6379")

	scrapeIntervalStr := getEnv("SCRAPE_INTERVAL", "30m")
	scrapeInterval, err := time.ParseDuration(scrapeIntervalStr)
	if err != nil {
		log.Printf("[market-collector] Formato SCRAPE_INTERVAL inválido ('%s'), usando 30m por defecto", scrapeIntervalStr)
		scrapeInterval = 30 * time.Minute
	}

	runOnStartup := getEnvBool("RUN_ON_STARTUP", true)
	jitterSeconds := getEnvInt("JITTER_SECONDS", 120) // Jitter de hasta 2 minutos para polite crawling

	cbFailures := getEnvInt("CIRCUIT_BREAKER_FAILURES", 5)
	cbTimeoutStr := getEnv("CIRCUIT_BREAKER_TIMEOUT", "1m")
	cbTimeout, _ := time.ParseDuration(cbTimeoutStr)
	if cbTimeout <= 0 {
		cbTimeout = time.Minute
	}

	magURL := os.Getenv("MAG_URL")
	rosganURL := os.Getenv("ROSGAN_URL")
	decampoURL := os.Getenv("DECAMPO_URL")
	tattersallURL := os.Getenv("TATTERSALL_URL")

	log.Printf("[market-collector] Configuración activa: Intervalo=%s, InicioInmediato=%t, JitterMax=%ds",
		scrapeInterval, runOnStartup, jitterSeconds)

	// 2. Conectar a PostgreSQL
	dbPool, err := database.NewPool(ctx, database.DefaultConfig(dbURL))
	if err != nil {
		log.Fatalf("[market-collector] Error fatal conectando a PostgreSQL: %v", err)
	}
	defer dbPool.Close()
	log.Println("[market-collector] Conexión establecida con PostgreSQL.")

	queries := db.New(dbPool)
	aclRepo := database.NewACLRepository(queries)
	quotRepo := database.NewQuotationRepository(queries)

	// 3. Conectar a Redis Streams (Event Bus)
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		redisOpts = &redis.Options{Addr: redisURL}
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("[market-collector] Advertencia: Ping a Redis falló (%v). Operando con reintentos...", err)
	} else {
		log.Println("[market-collector] Conexión establecida con Redis.")
	}

	bus := eventbus.NewRedisEventBus(redisClient)

	// 4. Inicializar Anti-Corruption Layer y Motor de Ingesta
	normalizer := collector.NewNormalizer(aclRepo)
	breakerCfg := resilience.Config{
		MaxFailures: cbFailures,
		Timeout:     cbTimeout,
	}
	engine := collector.NewEngine(normalizer, quotRepo, bus, breakerCfg)

	// 5. Registrar las 4 fuentes polimórficas de mercado
	engine.RegisterScraper(sources.NewMAGScraper(magURL, nil))
	engine.RegisterScraper(sources.NewRosganScraper(rosganURL, nil))
	engine.RegisterScraper(sources.NewDeCampoScraper(decampoURL, nil))
	engine.RegisterScraper(sources.NewTattersallScraper(tattersallURL, nil))
	log.Println("[market-collector] Estrategias registradas: MAG, Rosgan, De Campo a Campo, Monasterio Tattersall.")

	// Función auxiliar para ejecutar una ronda de extracción con telemetría básica
	runCycle := func() {
		log.Println("[market-collector] ==> Disparando ronda de extracción sobre todas las fuentes...")
		start := time.Now()

		cotizaciones, err := engine.Run(ctx)
		duration := time.Since(start)

		if err != nil {
			log.Printf("[market-collector] Ronda finalizada con advertencias en %s: %v", duration, err)
		} else {
			log.Printf("[market-collector] Ronda exitosa finalizada en %s. Total cotizaciones procesadas: %d",
				duration, len(cotizaciones))
		}
	}

	// Disparo inicial al arrancar si está habilitado
	if runOnStartup {
		runCycle()
	}

	// 6. Loop del Scheduler
	ticker := time.NewTicker(scrapeInterval)
	defer ticker.Stop()

	log.Printf("[market-collector] Scheduler en espera. Próxima ronda programada cada %s.", scrapeInterval)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		select {
		case <-ctx.Done():
			log.Println("[market-collector] Apagado solicitado. Finalizando tareas y cerrando conexiones...")
			return
		case <-ticker.C:
			// Aplicar jitter aleatorio para polite crawling
			if jitterSeconds > 0 {
				sleepSec := r.Intn(jitterSeconds)
				log.Printf("[market-collector] Aplicando jitter de espera cortés (%d segundos)...", sleepSec)
				select {
				case <-time.After(time.Duration(sleepSec) * time.Second):
				case <-ctx.Done():
					return
				}
			}

			runCycle()
		}
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

