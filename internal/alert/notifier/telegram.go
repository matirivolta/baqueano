package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/matirivolta/baqueano/internal/domain"
)

const (
	DefaultTelegramBaseURL = "https://api.telegram.org"
)

var (
	ErrTelegramTokenVacio  = errors.New("el token del bot de Telegram no puede estar vacío")
	ErrTelegramChatIDVacio = errors.New("el chat_id de Telegram no puede estar vacío")
)

// TelegramConfig contiene las credenciales y parámetros del cliente Telegram Bot API.
type TelegramConfig struct {
	BotToken string
	ChatID   string
	BaseURL  string
	Timeout  time.Duration
	Client   *http.Client
}

// TelegramNotifier implementa la interfaz alert.Notifier despachando mensajes vía Telegram Bot API.
type TelegramNotifier struct {
	botToken string
	chatID   string
	baseURL  string
	client   *http.Client
}

// NewTelegramNotifier construye una nueva instancia del adaptador de Telegram.
func NewTelegramNotifier(cfg TelegramConfig) (*TelegramNotifier, error) {
	if strings.TrimSpace(cfg.BotToken) == "" {
		return nil, ErrTelegramTokenVacio
	}
	if strings.TrimSpace(cfg.ChatID) == "" {
		return nil, ErrTelegramChatIDVacio
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultTelegramBaseURL
	}

	client := cfg.Client
	if client == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 10 * time.Second
		}
		client = &http.Client{Timeout: timeout}
	}

	return &TelegramNotifier{
		botToken: cfg.BotToken,
		chatID:   cfg.ChatID,
		baseURL:  strings.TrimRight(baseURL, "/"),
		client:   client,
	}, nil
}

// Canal retorna el identificador del canal TELEGRAM.
func (t *TelegramNotifier) Canal() domain.CanalNotificacion {
	return domain.CanalTelegram
}

type telegramPayload struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	ParseMode             string `json:"parse_mode"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

type telegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// Send realiza el POST a /bot<TOKEN>/sendMessage con formateo Markdown.
func (t *TelegramNotifier) Send(ctx context.Context, alerta domain.Alerta) error {
	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", t.baseURL, t.botToken)

	payload := telegramPayload{
		ChatID:                t.chatID,
		Text:                  alerta.Mensaje,
		ParseMode:             "Markdown",
		DisableWebPagePreview: true,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializando payload de Telegram: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("error construyendo HTTP request para Telegram: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("error ejecutando HTTP POST hacia Telegram: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error leyendo respuesta de Telegram: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram Bot API devolvió error HTTP %d: %s", resp.StatusCode, string(respBytes))
	}

	var tgResp telegramResponse
	if err := json.Unmarshal(respBytes, &tgResp); err == nil && !tgResp.OK {
		return fmt.Errorf("Telegram rechazó el mensaje: %s", tgResp.Description)
	}

	return nil
}

