package confman

import (
	"context"
	"fmt"

	"gitlab.com/micvbang/confman-go/pkg/storage"
)

type Confman interface {
	Add(ctx context.Context, key string, value string) error
	AddKeys(ctx context.Context, config map[string]string) error
	Read(ctx context.Context, key string) (value string, _ error)
	ReadKeys(ctx context.Context, keys []string) (map[string]string, error)
	ReadAll(ctx context.Context) (map[string]string, error)
	Move(ctx context.Context, confman Confman) error
	Copy(ctx context.Context, confman Confman) error
	String() string
}

type confman struct {
	log     Logger
	storage storage.Storage
}

func New(log Logger, storage storage.Storage) Confman {
	return &confman{
		log:     log,
		storage: storage,
	}
}

func (c *confman) Add(ctx context.Context, key string, value string) error {
	return c.storage.Add(ctx, key, value)
}

func (c *confman) AddKeys(ctx context.Context, config map[string]string) error {
	return c.storage.AddKeys(ctx, config)
}

func (c *confman) Read(ctx context.Context, key string) (value string, _ error) {
	return c.storage.Read(ctx, key)
}

func (c *confman) ReadKeys(ctx context.Context, keys []string) (map[string]string, error) {
	return c.storage.ReadKeys(ctx, keys)
}

func (c *confman) ReadAll(ctx context.Context) (map[string]string, error) {
	return c.storage.ReadAll(ctx)
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

	return c.storage.DeleteKeys(ctx, keys)
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

	config, err := c.storage.ReadAll(ctx)
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

func (c *confman) String() string {
	return fmt.Sprintf("Confman(%s)", c.storage)
}
