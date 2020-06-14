package httphelpers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/go-helpy/stringy"
)

type Router interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)

	AddRoute(method string, endpoint string, h http.HandlerFunc, mw ...HTTPWrapper) string
	ConfigureOPTIONS()
}

func NewDefaultRouter(log logger.Logger) Router {
	return &router{
		httpLogger: NewRequestLogger(true),
		log:        log,
		options:    make(map[string][]cors),
	}
}

type router struct {
	httprouter.Router

	httpLogger HTTPWrapper
	log        logger.Logger

	options map[string][]cors
}

type cors struct {
	domains []string
	methods []string
	headers []string
}

// httpMethodFunc represents an HTTP method endpoint handler, e.g. GET on the endpoint /messages/.
type httpMethodFunc func(path string, handle httprouter.Handle)

func (r *router) AddRoute(method string, route string, h http.HandlerFunc, mw ...HTTPWrapper) string {
	corsValues := cors{
		domains: []string{"*"},
		methods: []string{method},
		headers: []string{"Authorization", "Content-Type"},
	}
	r.options[route] = append(r.options[route], corsValues)

	corsHandler := NewCORSHandler(
		strings.Join(corsValues.domains, ", "),
		strings.Join(corsValues.methods, ", "),
		strings.Join(corsValues.headers, ", "),
	)

	return r.addRoute(method, route, corsHandler.WrapFunc(h), mw...)
}

func (r *router) addRoute(method string, route string, h http.HandlerFunc, mw ...HTTPWrapper) string {
	logger := r.httpLogger.WrapFunc

	// Add custom middlewares in first-in-last-executed order, i.e. the first
	// middleware of the slice is executed _after_ the following one, and so on.
	for _, m := range mw {
		h = m.Wrap(h)
	}

	var m httpMethodFunc

	switch method {
	case "GET":
		m = r.GET
	case "DELETE":
		m = r.DELETE
	case "OPTIONS":
		m = r.OPTIONS
	default:
		panic(fmt.Sprintf("Router: HTTP method %s not handled yet", method))
	}

	m(route, httpParametersWrapper(logger(h)))
	r.log.Infof("Registering route %s", route)

	return route
}

func (r *router) ConfigureOPTIONS() {
	for endpointPath, corsList := range r.options {
		domains := make([]string, 0, len(corsList))
		methods := make([]string, 0, len(corsList))
		headers := make([]string, 0, len(corsList))

		for _, cors := range corsList {
			domains = append(domains, cors.domains...)
			methods = append(methods, cors.methods...)
			headers = append(headers, cors.headers...)
		}

		domains = stringy.Unique(domains)
		methods = stringy.Unique(methods)
		headers = stringy.Unique(headers)
		corsHandler := NewCORSHandler(
			strings.Join(domains, ", "),
			strings.Join(methods, ", "),
			strings.Join(headers, ", "),
		)

		r.addRoute("OPTIONS", endpointPath, corsHandler.Handle())
	}
}

// httpParametersWrapper wraps an http.HandlerFunc in an httprouter.Handle to comply with the type signature of the
// httprouter.Router. This allows us to retrieve http route paramters using GetRouteParameterByName, without having
// to use the httprouter.Router throughout the codebase.
func httpParametersWrapper(h http.HandlerFunc) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h(w, r)
	})
}
