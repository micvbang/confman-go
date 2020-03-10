package confman

import (
	"context"
	"fmt"
	"path"

	"github.com/micvbang/go-helpy/mapy"
	"github.com/micvbang/go-helpy/stringy"
	"gitlab.com/micvbang/confman-go/pkg/logger"
	"gitlab.com/micvbang/confman-go/pkg/storage"
)

type Confman interface {
	Add(ctx context.Context, key string, value string) error
	AddKeys(ctx context.Context, config map[string]string) error

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

	// ServiceName returns the properly formatted service name
	ServiceName() string

	MetadataKeys() []string

	String() string
}

var ChamberCompatible bool = true

type confman struct {
	log               logger.Logger
	storage           storage.Storage
	serviceName       string
	chamberCompatible bool
}

func New(log logger.Logger, s storage.Storage, serviceName string) Confman {
	if ChamberCompatible {
		s = storage.NewChamberCompatibility(log, s)
	}

	return &confman{
		log:               log,
		storage:           s,
		serviceName:       FormatServiceName(serviceName),
		chamberCompatible: ChamberCompatible,
	}
}

func (c *confman) Add(ctx context.Context, key string, value string) error {
	return c.storage.Add(ctx, c.serviceName, key, value)
}

func (c *confman) AddKeys(ctx context.Context, config map[string]string) error {
	return c.storage.AddKeys(ctx, c.serviceName, config)
}

func (c *confman) Read(ctx context.Context, key string) (value string, _ error) {
	return c.storage.Read(ctx, c.serviceName, key)
}

func (c *confman) ReadKeys(ctx context.Context, keys []string) (map[string]string, error) {
	config, err := c.storage.ReadKeys(ctx, c.serviceName, keys)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *confman) ReadAll(ctx context.Context) (map[string]string, error) {
	config, err := c.storage.ReadAll(ctx, c.serviceName)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *confman) ReadAllMetadata(ctx context.Context) ([]storage.KeyMetadata, error) {
	keyMetadata, err := c.storage.ReadAllMetadata(ctx, c.serviceName)
	if err != nil {
		return nil, err
	}

	return keyMetadata, nil
}

func (c *confman) Move(ctx context.Context, dst Confman) error {
	// TODO: add chamber compatibility
	c.log.Debugf("Attempting to move %v to %v", c, dst)

	config, err := c.copy(ctx, dst)
	if err != nil {
		return err
	}

	keys := make([]string, len(config))
	for key := range config {
		keys = append(keys, key)
	}

	return c.storage.DeleteKeys(ctx, c.serviceName, keys)
}

func (c *confman) Copy(ctx context.Context, dst Confman) error {
	_, err := c.copy(ctx, dst)
	return err
}

func (c *confman) copy(ctx context.Context, dst Confman) (map[string]string, error) {
	// TODO: add chamber compatibility
	c.log.Debugf("Attempting to copy %v to %v", c, dst)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	config, err := c.storage.ReadAll(ctx, c.serviceName)
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

	err = dst.AddKeys(ctx, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *confman) Define(ctx context.Context, config map[string]string) error {
	newKeys, _ := mapy.StringKeys(config)
	newKeysLookup := stringy.ToSet(newKeys)

	currentConfig, err := c.storage.ReadAll(ctx, c.serviceName)
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

	err = c.storage.AddKeys(ctx, c.serviceName, config)
	if err != nil {
		return err
	}

	return c.storage.DeleteKeys(ctx, c.serviceName, keysToDelete)
}

func (c *confman) Delete(ctx context.Context, key string) error {
	return c.storage.Delete(ctx, c.ServiceName(), key)
}

func (c *confman) DeleteKeys(ctx context.Context, keys []string) error {
	return c.storage.DeleteKeys(ctx, c.serviceName, keys)
}

func (c *confman) DeleteAll(ctx context.Context) error {
	config, err := c.storage.ReadAll(ctx, c.serviceName)
	if err != nil {
		return err
	}

	keys, err := mapy.StringKeys(config)
	if err != nil {
		return err
	}

	return c.storage.DeleteKeys(ctx, c.serviceName, keys)
}

func (c *confman) ServiceName() string {
	return c.serviceName
}

func (c *confman) FormatKeyPath(key string) string {
	return path.Join(c.serviceName, key)
}

func (c *confman) MetadataKeys() []string {
	return c.storage.MetadataKeys()
}

func (c *confman) String() string {
	return fmt.Sprintf("Confman(service='%s', storage='%s')", c.serviceName, c.storage)
}
