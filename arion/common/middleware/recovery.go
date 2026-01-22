package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/zeirash/recapo/arion/handler"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n", err)
				log.Printf("stack trace:\n%s\n", debug.Stack())

				handler.WriteErrorJson(w, http.StatusInternalServerError, nil, "internal_server_error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
