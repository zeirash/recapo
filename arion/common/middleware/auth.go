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

type Middleware func(http.Handler) http.Handler

// MiddlewareWrapper takes Handler funcs and chains them to the main handler.
func MiddlewareWrapper(handler http.Handler, middlewares ...Middleware) http.Handler {
	// The loop is reversed so the middleware gets executed in the same
	// order as provided in the array.
	for i := len(middlewares); i > 0; i-- {
			handler = middlewares[i-1](handler)
	}
	return handler
}

func JwtMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					secret := config.GetConfig().SecretKey
					tokenStore := store.NewTokenStore()

					authHeader := r.Header.Get("Authorization")
					t := strings.Split(authHeader, " ")
					if len(t) != 2 {
						handler.WriteErrorJson(w, http.StatusUnauthorized, errors.New("unauthorized"), "invalid token format")
						return
					}

					authToken := t[1]
					authorized, err := tokenStore.IsAuthorized(authToken, secret)
					if err != nil {
						handler.WriteErrorJson(w, http.StatusUnauthorized, err, "is not authorized")
						return
					}

					if !authorized {
						handler.WriteErrorJson(w, http.StatusUnauthorized, errors.New("not authorzed"), "is not authorized")
						return
					}

					userID, err := tokenStore.ExtractIDFromToken(authToken, secret)
					if err != nil {
						handler.WriteErrorJson(w, http.StatusUnauthorized, err, "extract id from token")
						return
					}

					ctx := context.WithValue(context.Background(), common.UserIDKey, userID)
					r = r.WithContext(ctx)

					next.ServeHTTP(w, r)
			})
	}
}
