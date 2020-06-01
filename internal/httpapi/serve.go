package httpapi

import (
	"fmt"
	"net/http"

	"github.com/micvbang/confman-go/internal/httpapi/httphelpers"
	"github.com/micvbang/confman-go/pkg/logger"
)

// ListenAndServe starts an HTTP(s) server according to the given configuration
func ListenAndServe(flags Flags, log logger.Logger, addRoutes func(httphelpers.Router) httphelpers.Router) error {
	r := httphelpers.NewDefaultRouter(log)

	r = addRoutes(r)

	addr := fmt.Sprintf("%s:%s", flags.ListenAddr, flags.ListenPort)
	log.Printf("listening on %s", addr)

	return http.ListenAndServe(addr, r)
}
