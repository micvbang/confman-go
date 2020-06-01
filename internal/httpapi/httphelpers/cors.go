package httphelpers

import "net/http"

type CORSHandler interface {
	HTTPWrapper
	Handle() http.HandlerFunc
}

type corsHandler struct {
	origin  string
	methods string
	headers string
}

func NewCORSHandler(origin, methods, headers string) CORSHandler {
	return (&corsHandler{
		origin:  origin,
		methods: methods,
		headers: headers,
	})
}

func (c *corsHandler) wrap(h http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Origin", c.origin)
		(w).Header().Set("Access-Control-Allow-Methods", c.methods)
		(w).Header().Set("Access-Control-Allow-Headers", c.headers)

		if h != nil {
			h(w, r)
		}
	}
}

func (c *corsHandler) Wrap(h http.Handler) func(http.ResponseWriter, *http.Request) {
	return c.wrap(h.ServeHTTP)
}

func (c *corsHandler) WrapFunc(h func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return c.wrap(h)
}

func (c *corsHandler) Handle() http.HandlerFunc {
	return c.wrap(func(w http.ResponseWriter, r *http.Request) {})
}
