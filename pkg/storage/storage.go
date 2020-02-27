package storage

import "context"

// Storage defines the methods that must be implemented by a storage driver.
type Storage interface {
	Add(ctx context.Context, key string, value string) error
	Read(ctx context.Context, key string) (value string, _ error)
	ReadAll(ctx context.Context) (map[string]string, error)
	Delete(ctx context.Context, key string) error
	DeleteKeys(ctx context.Context, keys []string) error

	// History?
}
