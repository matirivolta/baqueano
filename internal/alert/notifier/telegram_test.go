package notifier_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matirivolta/baqueano/internal/alert/notifier"
	"github.com/matirivolta/baqueano/internal/domain"
)

func TestTelegramNotifier_SendSuccess(t *testing.T) {
	t.Parallel()

	var capturedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("esperado POST, obtenido %s", r.Method)
		}
		if r.URL.Path != "/bot123456:ABC-DEF/sendMessage" {
			t.Errorf("path incorrecto: %s", r.URL.Path)
		}

		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true, "result": {"message_id": 999}}`))
	}))
	defer server.Close()

	n, err := notifier.NewTelegramNotifier(notifier.TelegramConfig{
		BotToken: "123456:ABC-DEF",
		ChatID:   "-100123456789",
		BaseURL:  server.URL,
		Client:   server.Client(),
	})
	if err != nil {
		t.Fatalf("error construyendo notifier: %v", err)
	}

	if n.Canal() != domain.CanalTelegram {
		t.Errorf("esperado canal TELEGRAM, obtenido %s", n.Canal())
	}

	alerta := domain.Alerta{
		Fecha:            time.Now(),
		CategoriaCanonID: 1,
		TipoAlerta:       domain.AlertaVariacionDiaria,
		Canal:            domain.CanalTelegram,
		Mensaje:          "🚨 *Suba del 5.2% en Novillito*",
	}

	if err := n.Send(context.Background(), alerta); err != nil {
		t.Fatalf("error inesperado enviando a Telegram: %v", err)
	}

	if capturedBody["chat_id"] != "-100123456789" {
		t.Errorf("chat_id capturado incorrecto: %v", capturedBody["chat_id"])
	}
	if capturedBody["text"] != alerta.Mensaje {
		t.Errorf("texto capturado incorrecto: %v", capturedBody["text"])
	}
	if capturedBody["parse_mode"] != "Markdown" {
		t.Errorf("parse_mode incorrecto: %v", capturedBody["parse_mode"])
	}
}

func TestTelegramNotifier_ErrorResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"ok": false, "description": "Unauthorized"}`))
	}))
	defer server.Close()

	n, err := notifier.NewTelegramNotifier(notifier.TelegramConfig{
		BotToken: "token_invalido",
		ChatID:   "123",
		BaseURL:  server.URL,
		Client:   server.Client(),
	})
	if err != nil {
		t.Fatalf("error construyendo notifier: %v", err)
	}

	alerta := domain.Alerta{
		Fecha:   time.Now(),
		Mensaje: "Test",
	}

	err = n.Send(context.Background(), alerta)
	if err == nil {
		t.Fatalf("se esperaba error HTTP 401 de Telegram")
	}
}

