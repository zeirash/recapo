package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	sentry "github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/handler"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.WithFields(logrus.Fields{
					"error":       rec,
					"stack_trace": string(debug.Stack()),
				}).Error("panic recovered")

				var panicErr error
				switch v := rec.(type) {
				case error:
					panicErr = v
				default:
					panicErr = fmt.Errorf("%v", v)
				}
				sentry.CaptureException(panicErr)

				handler.WriteErrorJson(w, r, http.StatusInternalServerError, nil, "internal_server_error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
