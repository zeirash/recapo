package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetLangFromRequest(t *testing.T) {
	tests := []struct {
		name       string
		nilRequest bool
		acceptLang string
		want       string
	}{
		{
			name:       "nil request returns en",
			nilRequest: true,
			want:       "en",
		},
		{
			name: "no Accept-Language header returns en",
			want: "en",
		},
		{
			name:       "explicit en",
			acceptLang: "en",
			want:       "en",
		},
		{
			name:       "explicit id",
			acceptLang: "id",
			want:       "id",
		},
		{
			name:       "en-US maps to en",
			acceptLang: "en-US",
			want:       "en",
		},
		{
			name:       "id-ID maps to id",
			acceptLang: "id-ID",
			want:       "id",
		},
		{
			name:       "first preferred language wins",
			acceptLang: "id-ID,en;q=0.9",
			want:       "id",
		},
		{
			name:       "unsupported language falls back to en",
			acceptLang: "fr,de",
			want:       "en",
		},
		{
			name:       "unsupported first lang falls through to id",
			acceptLang: "fr,id;q=0.9",
			want:       "id",
		},
		{
			name:       "quality value stripped correctly",
			acceptLang: "en;q=0.8",
			want:       "en",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r *http.Request
			if !tt.nilRequest {
				r = httptest.NewRequest(http.MethodGet, "/", nil)
				if tt.acceptLang != "" {
					r.Header.Set("Accept-Language", tt.acceptLang)
				}
			}
			got := GetLangFromRequest(r)
			if got != tt.want {
				t.Errorf("GetLangFromRequest() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestT(t *testing.T) {
	tests := []struct {
		name string
		lang string
		code string
		want string
	}{
		{
			name: "english key exists",
			lang: "en",
			code: "err_email_required",
			want: "Email is required",
		},
		{
			name: "indonesian key exists",
			lang: "id",
			code: "err_email_required",
			want: "Email wajib diisi",
		},
		{
			name: "unknown lang falls back to en",
			lang: "fr",
			code: "err_email_required",
			want: "Email is required",
		},
		{
			name: "unknown code returns empty string",
			lang: "en",
			code: "this_key_does_not_exist",
			want: "",
		},
		{
			name: "unknown code unknown lang returns empty string",
			lang: "fr",
			code: "this_key_does_not_exist",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := T(tt.lang, tt.code)
			if got != tt.want {
				t.Errorf("T(%q, %q) = %q, want %q", tt.lang, tt.code, got, tt.want)
			}
		})
	}
}

func TestMessage(t *testing.T) {
	tests := []struct {
		name       string
		nilRequest bool
		acceptLang string
		code       string
		fallback   string
		want       string
	}{
		{
			name:       "nil request unknown code returns fallback",
			nilRequest: true,
			code:       "this_key_does_not_exist",
			fallback:   "fallback message",
			want:       "fallback message",
		},
		{
			name:       "returns en translation",
			acceptLang: "en",
			code:       "err_email_required",
			fallback:   "fallback",
			want:       "Email is required",
		},
		{
			name:       "returns id translation",
			acceptLang: "id",
			code:       "err_email_required",
			fallback:   "fallback",
			want:       "Email wajib diisi",
		},
		{
			name:       "unknown code returns fallback",
			acceptLang: "en",
			code:       "this_key_does_not_exist",
			fallback:   "my fallback",
			want:       "my fallback",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r *http.Request
			if !tt.nilRequest {
				r = httptest.NewRequest(http.MethodGet, "/", nil)
				if tt.acceptLang != "" {
					r.Header.Set("Accept-Language", tt.acceptLang)
				}
			}
			got := Message(r, tt.code, tt.fallback)
			if got != tt.want {
				t.Errorf("Message() = %q, want %q", got, tt.want)
			}
		})
	}
}
