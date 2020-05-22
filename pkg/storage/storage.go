package storage

import (
	"context"
)

//go:generate mockery -inpkg -name Storage -case=underscore

// Storage defines the methods that must be implemented by a storage driver.
type Storage interface {
	Write(ctx context.Context, servicePath string, key string, value string) error
	WriteKeys(ctx context.Context, servicePath string, config map[string]string) error

	Read(ctx context.Context, servicePath string, key string) (value string, _ error)
	ReadKeys(ctx context.Context, servicePath string, keys []string) (map[string]string, error)

	ReadAll(ctx context.Context, servicePath string) (map[string]string, error)
	ReadAllMetadata(ctx context.Context, servicePath string) ([]KeyMetadata, error)

	Delete(ctx context.Context, servicePath string, key string) error
	DeleteKeys(ctx context.Context, servicePath string, keys []string) error

	MetadataKeys() []string

	// History?
	String() string
}

type KeyMetadata struct {
	Key      string            `json:"key"`
	Value    string            `json:"value"`
	Metadata map[string]string `json:"metadata"`
}
