package httpapi

import (
	"github.com/micvbang/confman-go/internal/httpapi/httphelpers"
	"github.com/micvbang/confman-go/pkg/storage"
)

// AddRoutes registers and maps all HTTP endpoints to their respective routes.
func AddRoutes(r httphelpers.Router, d Dependencies) httphelpers.Router {
	r.AddRoute("GET", "/service_paths", NewServicePathConfigLister(d.Storage))
	r.AddRoute("DELETE", "/service_paths/keys", NewServicePathKeysDeleter(d.Storage))
	r.AddRoute("PUT", "/service_paths/key", NewServicePathKeyWriter(d.Storage))
	r.AddRoute("GET", "/service_paths/key", NewServicePathKeyReader(d.Storage))

	r.ConfigureOPTIONS()

	return r
}

type Dependencies struct {
	Storage storage.Storage
}
