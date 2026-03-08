package i18n

import (
	"embed"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

//go:embed messages/*.json
var messagesFS embed.FS

var (
	mu       sync.RWMutex
	messages = make(map[string]map[string]string) // lang -> code -> message
)

const defaultLang = "en"

func init() {
	loadLang("en")
	loadLang("id")
}

func loadLang(lang string) {
	data, err := messagesFS.ReadFile("messages/" + lang + ".json")
	if err != nil {
		return
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return
	}
	mu.Lock()
	messages[lang] = m
	mu.Unlock()
}

// GetLangFromRequest returns the preferred language from Accept-Language header.
// Supports "en" and "id"; defaults to "en".
func GetLangFromRequest(r *http.Request) string {
	if r == nil {
		return defaultLang
	}
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang == "" {
		return defaultLang
	}
	// Accept-Language can be "en-US,en;q=0.9,id;q=0.8" – take first preferred.
	for _, part := range strings.Split(acceptLang, ",") {
		part = strings.TrimSpace(part)
		if idx := strings.Index(part, ";"); idx > 0 {
			part = part[:idx]
		}
		part = strings.TrimSpace(part)
		if len(part) >= 2 {
			lang := strings.ToLower(part[:2])
			if lang == "id" || lang == "en" {
				return lang
			}
		}
	}
	return defaultLang
}

// T returns the translated message for the given language and code.
// Falls back to English if the key is missing in the requested language.
// Returns empty string if not found in any language.
func T(lang, code string) string {
	mu.RLock()
	langMap := messages[lang]
	mu.RUnlock()
	if msg, ok := langMap[code]; ok {
		return msg
	}
	// Fall back to English if key missing in requested language.
	if lang != defaultLang {
		mu.RLock()
		enMap := messages[defaultLang]
		mu.RUnlock()
		if msg, ok := enMap[code]; ok {
			return msg
		}
	}
	return ""
}

// Message returns the translated message for the request's language and code.
// If no translation exists, returns fallback.
func Message(r *http.Request, code string, fallback string) string {
	lang := GetLangFromRequest(r)
	if msg := T(lang, code); msg != "" {
		return msg
	}
	return fallback
}
