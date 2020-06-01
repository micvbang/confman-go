package httpapi

import (
	"context"
	"net/http"

	"github.com/micvbang/confman-go/internal/httpapi/httphelpers"
	"github.com/micvbang/confman-go/pkg/storage"
	"github.com/micvbang/go-helpy/booly"
)

func NewServicePathConfigLister(s storage.Storage) http.HandlerFunc {
	return httphelpers.StatusHandler(func(w http.ResponseWriter, r *http.Request) httphelpers.Status {
		query := r.URL.Query()

		path := query.Get("path")
		if len(path) == 0 {
			path = "/"
		}

		recursive := booly.FromString(query.Get("recursive"))

		ctx := context.Background()
		servicePathConfigs, err := s.PathRead(ctx, path, recursive)
		if err != nil {
			return err
		}

		return httphelpers.WriteJSON(w, servicePathConfigs)
	})
}
