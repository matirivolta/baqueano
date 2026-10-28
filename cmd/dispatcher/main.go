package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"

	"github.com/matirivolta/baqueano/internal/alert"
	"github.com/matirivolta/baqueano/internal/alert/notifier"
	"github.com/matirivolta/baqueano/internal/db"
	"github.com/matirivolta/baqueano/internal/platform/database"
	"github.com/matirivolta/baqueano/internal/platform/eventbus"
)

func main() {
	log.Println("[alert-dispatcher] Arrancando worker de alertas ganaderas Baqueano...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Cargar variables de entorno
	dbURL := getEnv("DATABASE_URL", "postgres://cotizador_app:cotizador_secret_pwd@localhost:5432/cotizador_db?sslmode=disable")
	redisURL := getEnv("REDIS_URL", "localhost:6379")
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	tgChatID := os.Getenv("TELEGRAM_CHAT_ID")
	streamName := getEnv("REDIS_STREAM_NAME", eventbus.DefaultStreamName)
	groupName := getEnv("REDIS_GROUP_NAME", eventbus.DefaultGroupName)
	consumerName := getEnv("CONSUMER_NAME", "dispatcher-worker-1")

	// 2. Conectar a PostgreSQL
	dbPool, err := database.NewPool(ctx, database.DefaultConfig(dbURL))
	if err != nil {
		log.Fatalf("[alert-dispatcher] Error fatal conectando a PostgreSQL: %v", err)
	}
	defer dbPool.Close()
	log.Println("[alert-dispatcher] Conexión establecida con PostgreSQL.")

	queries := db.New(dbPool)
	quotRepo := database.NewQuotationRepository(queries)
	alertRepo := database.NewAlertRepository(queries)

	// 3. Conectar a Redis
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		redisOpts = &redis.Options{Addr: redisURL}
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("[alert-dispatcher] Advertencia: Ping a Redis falló (%v). Intentando reconectar en loop...", err)
	} else {
		log.Println("[alert-dispatcher] Conexión establecida con Redis.")
	}

	bus := eventbus.NewRedisEventBus(redisClient)

	// 4. Configurar motor de reglas y adaptadores de notificación
	engine := alert.NewEngine(alert.DefaultConfig(), quotRepo, alertRepo)

	if tgToken != "" && tgChatID != "" {
		tgNotifier, err := notifier.NewTelegramNotifier(notifier.TelegramConfig{
			BotToken: tgToken,
			ChatID:   tgChatID,
		})
		if err != nil {
			log.Fatalf("[alert-dispatcher] Error configurando Telegram: %v", err)
		}
		engine.RegisterNotifier(tgNotifier)
		log.Println("[alert-dispatcher] Notificador de Telegram inicializado y registrado.")
	} else {
		log.Println("[alert-dispatcher] TELEGRAM_BOT_TOKEN o TELEGRAM_CHAT_ID no configurados. Operando en modo audit-only.")
	}

	// 5. Iniciar orquestador de consumo reactivo
	dispatcher := alert.NewDispatcher(bus, engine, streamName, groupName, consumerName)

	if err := dispatcher.Start(ctx); err != nil {
		log.Fatalf("[alert-dispatcher] Error en loop del dispatcher: %v", err)
	}

	log.Println("[alert-dispatcher] Apagado ordenado finalizado.")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

