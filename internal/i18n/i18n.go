// Package i18n provides multi-language support for the simplex-node server.
package i18n

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

type Messages map[string]map[string]string // lang_code -> key -> message

type Translator struct {
	mu       sync.RWMutex
	messages Messages
	path     string
}


// NewTranslator handles the NewTranslator HTTP request.
func NewTranslator(dataDir string) *Translator {
	t := &Translator{
		messages: Messages{
			"en": defaultEN,
			"ru": defaultRU,
			"es": defaultES,
		},
		path: filepath.Join(dataDir, "i18n.json"),
	}
	t.load()
	return t
}

func (t *Translator) load() {
	data, err := os.ReadFile(t.path)
	if err != nil {
		return
	}
	var custom Messages
	if err := json.Unmarshal(data, &custom); err != nil {
		slog.Warn("i18n load", "error", err)
		return
	}
	// Merge custom messages on top of defaults
	for lang, msgs := range custom {
		if t.messages[lang] == nil {
			t.messages[lang] = map[string]string{}
		}
		for key, msg := range msgs {
			t.messages[lang][key] = msg
		}
	}
	slog.Info("i18n loaded", "languages", len(t.messages))
}

func (t *Translator) save() {
	data, err := json.MarshalIndent(t.messages, "", "  ")
	if err != nil {
		slog.Error("i18n save", "error", err)
		return
	}
	if err := os.WriteFile(t.path, data, 0644); err != nil {
		slog.Error("i18n save write", "error", err)
	}
}

// T translates a key to the given language. Falls back to English, then the key itself.
func (t *Translator) T(lang, key string, args ...any) string {
	t.mu.RLock()
	langMsgs, ok := t.messages[lang]
	if !ok {
		langMsgs = t.messages["en"]
	}
	msg, ok := langMsgs[key]
	if !ok {
		msg = t.messages["en"][key]
	}
	t.mu.RUnlock()

	if msg == "" {
		return key
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// SetCustom adds or updates a custom message.
func (t *Translator) SetCustom(lang, key, msg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.messages[lang] == nil {
		t.messages[lang] = map[string]string{}
	}
	t.messages[lang][key] = msg
	t.save()
}

// Languages returns available language codes.
func (t *Translator) Languages() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	langs := make([]string, 0, len(t.messages))
	for l := range t.messages {
		langs = append(langs, l)
	}
	return langs
}

// Global translator instance
var Global *Translator


// Init handles the Init HTTP request.
func Init(dataDir string) {
	Global = NewTranslator(dataDir)
}

// T is a convenience wrapper around Global.T
func T(lang, key string, args ...any) string {
	if Global == nil {
		return key
	}
	return Global.T(lang, key, args...)
}

// ── Default Messages ─────────────────────────────────────────────────────────

var defaultEN = map[string]string{
	"welcome":              "Welcome to Saint Mary Liberty Island!",
	"goodbye":              "Farewell, citizen!",
	"error_internal":       "Internal server error",
	"error_not_found":      "Resource not found",
	"error_bad_request":    "Bad request",
	"error_unauthorized":   "Unauthorized",
	"marketplace_empty":    "No active listings",
	"marketplace_created":  "Listing created successfully",
	"marketplace_sold":     "Item marked as sold",
	"dao_proposal_created": "Proposal created successfully",
	"dao_vote_recorded":    "Vote recorded",
	"dao_proposal_passed":  "Proposal has passed!",
	"dao_proposal_rejected":"Proposal was rejected",
	"health_ok":            "System healthy",
	"health_warn":          "System degraded",
	"health_down":          "System down",
}

var defaultRU = map[string]string{
	"welcome":              "Добро пожаловать на Остров Святой Марии!",
	"goodbye":              "Прощайте, гражданин!",
	"error_internal":       "Внутренняя ошибка сервера",
	"error_not_found":      "Ресурс не найден",
	"error_bad_request":    "Неверный запрос",
	"error_unauthorized":   "Не авторизован",
	"marketplace_empty":    "Нет активных объявлений",
	"marketplace_created":  "Объявление создано",
	"marketplace_sold":     "Товар продан",
	"dao_proposal_created": "Предложение создано",
	"dao_vote_recorded":    "Голос учтён",
	"dao_proposal_passed":  "Предложение принято!",
	"dao_proposal_rejected":"Предложение отклонено",
	"health_ok":            "Система здорова",
	"health_warn":          "Система деградирована",
	"health_down":          "Система не работает",
}

var defaultES = map[string]string{
	"welcome":              "¡Bienvenido a la Isla Santa María!",
	"goodbye":              "¡Adiós, ciudadano!",
	"error_internal":       "Error interno del servidor",
	"error_not_found":      "Recurso no encontrado",
	"error_bad_request":    "Solicitud incorrecta",
	"error_unauthorized":   "No autorizado",
	"marketplace_empty":    "No hay anuncios activos",
	"marketplace_created":  "Anuncio creado con éxito",
	"marketplace_sold":     "Artículo vendido",
	"dao_proposal_created": "Propuesta creada con éxito",
	"dao_vote_recorded":    "Voto registrado",
	"dao_proposal_passed":  "¡Propuesta aprobada!",
	"dao_proposal_rejected":"Propuesta rechazada",
	"health_ok":            "Sistema saludable",
	"health_warn":          "Sistema degradado",
	"health_down":          "Sistema caído",
}
