package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zeirash/recapo/arion/common/logger"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"ip":     r.RemoteAddr,
			"took":   time.Since(start).String(),
		}).Info("request")
	})
}
