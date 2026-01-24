package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/handler"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.WithFields(logrus.Fields{
					"error":       err,
					"stack_trace": string(debug.Stack()),
				}).Error("panic recovered")

				handler.WriteErrorJson(w, http.StatusInternalServerError, nil, "internal_server_error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
