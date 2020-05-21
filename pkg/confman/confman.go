package confman

import (
	"context"
	"fmt"
	"path"

	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"github.com/micvbang/go-helpy/mapy"
	"github.com/micvbang/go-helpy/stringy"
)

type Confman interface {
	Write(ctx context.Context, key string, value string) error
	WriteKeys(ctx context.Context, config map[string]string) error

	Read(ctx context.Context, key string) (value string, _ error)
	ReadKeys(ctx context.Context, keys []string) (map[string]string, error)
	ReadAll(ctx context.Context) (map[string]string, error)
	ReadAllMetadata(ctx context.Context) ([]storage.KeyMetadata, error)

	Delete(ctx context.Context, key string) error
	DeleteKeys(ctx context.Context, keys []string) error
	DeleteAll(ctx context.Context) error

	Move(ctx context.Context, confman Confman) error
	Copy(ctx context.Context, confman Confman) error

	// Define enforces the current config, i.e. keys in config will be
	// added/updated and keys not in config will be deleted.
	Define(ctx context.Context, config map[string]string) error

	// FormatKeyPath returns the full path of the given key, i.e. including
	// the service name.
	FormatKeyPath(key string) string

	// ServicePath returns the properly formatted service name
	ServicePath() string

	MetadataKeys() []string

	String() string
}

var ChamberCompatible bool = true

type confman struct {
	log         logger.Logger
	storage     storage.Storage
	servicePath string
}

func New(log logger.Logger, s storage.Storage, servicePath string) Confman {
	if ChamberCompatible {
		s = storage.NewChamberCompatibility(log, s)
	}

	return &confman{
		log:         log,
		storage:     s,
		servicePath: FormatServicePath(servicePath),
	}
}

func (c *confman) Write(ctx context.Context, key string, value string) error {
	return c.storage.Write(ctx, c.servicePath, key, value)
}

func (c *confman) WriteKeys(ctx context.Context, config map[string]string) error {
	return c.storage.WriteKeys(ctx, c.servicePath, config)
}

func (c *confman) Read(ctx context.Context, key string) (value string, _ error) {
	return c.storage.Read(ctx, c.servicePath, key)
}

func (c *confman) ReadKeys(ctx context.Context, keys []string) (map[string]string, error) {
	config, err := c.storage.ReadKeys(ctx, c.servicePath, keys)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *confman) ReadAll(ctx context.Context) (map[string]string, error) {
	config, err := c.storage.ReadAll(ctx, c.servicePath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *confman) ReadAllMetadata(ctx context.Context) ([]storage.KeyMetadata, error) {
	keyMetadata, err := c.storage.ReadAllMetadata(ctx, c.servicePath)
	if err != nil {
		return nil, err
	}

	return keyMetadata, nil
}

func (c *confman) Move(ctx context.Context, dst Confman) error {
	c.log.Debugf("Attempting to move %v to %v", c, dst)

	config, err := c.copy(ctx, dst)
	if err != nil {
		return err
	}

	keys := make([]string, len(config))
	for key := range config {
		keys = append(keys, key)
	}

	return c.storage.DeleteKeys(ctx, c.servicePath, keys)
}

func (c *confman) Copy(ctx context.Context, dst Confman) error {
	_, err := c.copy(ctx, dst)
	return err
}

func (c *confman) copy(ctx context.Context, dst Confman) (map[string]string, error) {
	c.log.Debugf("Attempting to copy %v to %v", c, dst)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	config, err := c.storage.ReadAll(ctx, c.servicePath)
	if err != nil {
		return nil, err
	}

	if len(config) == 0 {
		c.log.Warnf("No keys copied")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	err = dst.WriteKeys(ctx, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *confman) Define(ctx context.Context, config map[string]string) error {
	newKeys, _ := mapy.StringKeys(config)
	newKeysLookup := stringy.ToSet(newKeys)

	currentConfig, err := c.storage.ReadAll(ctx, c.servicePath)
	if err != nil {
		return err
	}

	currentKeys, _ := mapy.StringKeys(currentConfig)
	keysToDelete := make([]string, 0, len(currentKeys))
	for _, currentKey := range currentKeys {
		if !newKeysLookup.Contains(currentKey) {
			keysToDelete = append(keysToDelete, currentKey)
		}
	}

	c.log.Debugf("Defining new keys to be %v, deleting keys %v", newKeys, keysToDelete)

	// TODO: ask user before deleting

	err = c.storage.WriteKeys(ctx, c.servicePath, config)
	if err != nil {
		return err
	}

	return c.storage.DeleteKeys(ctx, c.servicePath, keysToDelete)
}

func (c *confman) Delete(ctx context.Context, key string) error {
	return c.storage.Delete(ctx, c.ServicePath(), key)
}

func (c *confman) DeleteKeys(ctx context.Context, keys []string) error {
	return c.storage.DeleteKeys(ctx, c.servicePath, keys)
}

func (c *confman) DeleteAll(ctx context.Context) error {
	config, err := c.storage.ReadAll(ctx, c.servicePath)
	if err != nil {
		return err
	}

	keys, err := mapy.StringKeys(config)
	if err != nil {
		return err
	}

	return c.storage.DeleteKeys(ctx, c.servicePath, keys)
}

func (c *confman) ServicePath() string {
	return c.servicePath
}

func (c *confman) FormatKeyPath(key string) string {
	return path.Join(c.servicePath, key)
}

func (c *confman) MetadataKeys() []string {
	return c.storage.MetadataKeys()
}

func (c *confman) String() string {
	return fmt.Sprintf("Confman(service='%s', storage='%s')", c.servicePath, c.storage)
}
