package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/handler"
	"github.com/zeirash/recapo/arion/store"
)

// ChainMiddleware takes Handler funcs and chains them to the main handler.
func ChainMiddleware(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		// The loop is reversed so the middleware gets executed in the same
		// order as provided in the array.
		for i := len(middlewares); i > 0; i-- {
			final = middlewares[i-1](final)
		}
		return final
	}
}

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := config.GetConfig().SecretKey
		tokenStore := store.NewTokenStore()

		authHeader := r.Header.Get("Authorization")
		t := strings.Split(authHeader, " ")
		if len(t) != 2 {
			handler.WriteErrorJson(w, http.StatusUnauthorized, errors.New("invalid token format"), "unauthorized")
			return
		}

		authToken := t[1]
		authorized, err := tokenStore.IsAuthorized(authToken, secret)
		if err != nil {
			handler.WriteErrorJson(w, http.StatusUnauthorized, err, "unauthorized")
			return
		}

		if !authorized {
			handler.WriteErrorJson(w, http.StatusUnauthorized, errors.New("is not authorzed"), "unauthorized")
			return
		}

		tokenData, err := tokenStore.ExtractDataFromToken(authToken, secret)
		if err != nil {
			handler.WriteErrorJson(w, http.StatusInternalServerError, err, "extract_data")
			return
		}

		ctx := context.WithValue(r.Context(), common.UserIDKey, tokenData.UserID)
		ctx = context.WithValue(ctx, common.SystemModeKey, tokenData.SystemMode)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func CheckSystemMode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		isSystemMode := ctx.Value(common.SystemModeKey).(bool)

		if !isSystemMode {
			handler.WriteErrorJson(w, http.StatusUnauthorized, errors.New("doesn't have system mode access"), "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}
