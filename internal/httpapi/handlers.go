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

func NewServicePathKeysDeleter(s storage.Storage) http.HandlerFunc {
	return httphelpers.StatusHandler(func(w http.ResponseWriter, r *http.Request) httphelpers.Status {
		input := ServicePathConfigDeleteInput{}
		err := httphelpers.ParseJSON(r, &input)
		if err != nil {
			return httphelpers.StatusBadRequest
		}

		ctx := context.Background()
		return s.DeleteKeys(ctx, input.ServicePath, input.Keys)
	})
}

func NewServicePathKeyWriter(s storage.Storage) http.HandlerFunc {
	return httphelpers.StatusHandler(func(w http.ResponseWriter, r *http.Request) httphelpers.Status {
		input := ServicePathConfigWriteKeyInput{}
		err := httphelpers.ParseJSON(r, &input)
		if err != nil {
			return httphelpers.StatusBadRequest
		}

		ctx := context.Background()
		return s.Write(ctx, input.ServicePath, input.Key, input.Value)
	})
}

func NewServicePathKeyReader(s storage.Storage) http.HandlerFunc {
	return httphelpers.StatusHandler(func(w http.ResponseWriter, r *http.Request) httphelpers.Status {
		query := r.URL.Query()
		servicePath := query.Get("service-path")
		key := query.Get("key")

		ctx := context.Background()
		value, err := s.Read(ctx, servicePath, key)
		if err != nil {
			return err
		}

		return httphelpers.WriteJSON(w, value)
	})
}

type ServicePathConfigDeleteInput struct {
	ServicePath string   `json:"path"`
	Keys        []string `json:"keys"`
}

type ServicePathConfigWriteKeyInput struct {
	ServicePath string `json:"path"`
	Key         string `json:"key"`
	Value       string `json:"value"`
}
