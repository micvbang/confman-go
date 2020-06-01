package httphelpers

import (
	"net/http"
)

// HTTPWrapper provides wrappers for HTTP methods
type HTTPWrapper interface {
	Wrap(h http.Handler) func(http.ResponseWriter, *http.Request)
	WrapFunc(h func(http.ResponseWriter, *http.Request)) http.HandlerFunc
}
