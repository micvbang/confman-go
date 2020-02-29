package storage

import "context"

// Storage defines the methods that must be implemented by a storage driver.
type Storage interface {
	Add(ctx context.Context, key string, value string) error
	AddKeys(ctx context.Context, config map[string]string) error
	Read(ctx context.Context, key string) (value string, _ error)
	ReadKeys(ctx context.Context, keys []string) (map[string]string, error)
	ReadAll(ctx context.Context) (map[string]string, error)
	Delete(ctx context.Context, key string) error
	DeleteKeys(ctx context.Context, keys []string) error

	// History?
	String() string
}
