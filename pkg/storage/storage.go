package storage

import (
	"context"
)

// Storage defines the methods that must be implemented by a storage driver.
type Storage interface {
	Add(ctx context.Context, serviceName string, key string, value string) error
	AddKeys(ctx context.Context, serviceName string, config map[string]string) error

	Read(ctx context.Context, serviceName string, key string) (value string, _ error)
	ReadKeys(ctx context.Context, serviceName string, keys []string) (map[string]string, error)

	ReadAll(ctx context.Context, serviceName string) (map[string]string, error)
	ReadAllMetadata(ctx context.Context, serviceName string) ([]KeyMetadata, error)

	Delete(ctx context.Context, serviceName string, key string) error
	DeleteKeys(ctx context.Context, serviceName string, keys []string) error

	MetadataKeys() []string

	// History?
	String() string
}

type KeyMetadata struct {
	Key      string            `json:"key"`
	Value    string            `json:"value"`
	Metadata map[string]string `json:"metadata"`
}
