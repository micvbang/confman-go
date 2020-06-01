package httphelpers

import (
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

type requestLogger struct {
	log bool
}

func NewRequestLogger(log bool) HTTPWrapper {
	return &requestLogger{log: log}
}

func (rl *requestLogger) wrap(h http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.log {
			h(w, r)
			return
		}

		log.Printf("Request: %+v -> %+v: %+v%+v\n", r.RemoteAddr, r.Method, r.Host, r.URL)

		responseRecorder := httptest.NewRecorder()
		start := time.Now()
		h(responseRecorder, r)
		elapsed := time.Since(start)

		for k, v := range responseRecorder.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(responseRecorder.Code)
		responseRecorder.Body.WriteTo(w)

		log.Printf("Response (took %s): %+v <- %+v: %+v%+v %+v:\n", elapsed, r.RemoteAddr, r.Method, r.Host, r.URL, responseRecorder.Code)
	}
}

func (rl *requestLogger) Wrap(h http.Handler) func(http.ResponseWriter, *http.Request) {
	return rl.wrap(h.ServeHTTP)
}

func (rl *requestLogger) WrapFunc(h func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return rl.wrap(h)
}
