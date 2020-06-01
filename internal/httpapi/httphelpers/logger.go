package httphelpers

import (
	"context"
	"net/http"

	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/sirupsen/logrus"
)

type contextKey string

const logKey contextKey = "logger"

// GetLogger returns a logger. If the router has not been initialized with
// one, it creates a new one and emits a warning.
func GetLogger(r *http.Request) logger.Logger {
	log, ok := r.Context().Value(logKey).(logger.Logger)
	if !ok {
		log = logger.LogrusWrapper{Logger: logrus.New()}
		log.Warnf("Logger is nil - this should only happen in tests.")
	}
	return log
}

func SetLogger(r *http.Request, log logger.Logger) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), logKey, log))
}
