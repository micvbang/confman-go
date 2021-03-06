package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/micvbang/confman-go/pkg/logger"
)

type ChamberCompatibility struct {
	storage Storage
	log     logger.Logger
}

var _ Storage = &ChamberCompatibility{}

func NewChamberCompatibility(log logger.Logger, storage Storage) *ChamberCompatibility {
	log = log.WithField("storage_type", "ChamberCompatibility")

	return &ChamberCompatibility{
		storage: storage,
		log:     log,
	}
}

func (c *ChamberCompatibility) Write(ctx context.Context, servicePath string, key string, value string) error {
	c.log.Debugf("Write(ctx, \"%s\", \"%s\", \"%s\")", servicePath, key, value)

	key = c.chamberKeyToLower(key)
	return c.storage.Write(ctx, servicePath, key, value)
}

func (c *ChamberCompatibility) WriteKeys(ctx context.Context, servicePath string, config map[string]string) error {
	c.log.Debugf("WriteKeys(ctx, \"%s\", %+v)", servicePath, config)

	newConfig := c.chamberConfigToLower(config)
	return c.storage.WriteKeys(ctx, servicePath, newConfig)
}

func (c *ChamberCompatibility) Read(ctx context.Context, servicePath string, key string) (value string, _ error) {
	c.log.Debugf("Read(ctx, \"%s\", \"%s\")", servicePath, key)

	key = c.chamberKeyToLower(key)
	return c.storage.Read(ctx, servicePath, key)
}

func (c *ChamberCompatibility) ReadKeys(ctx context.Context, servicePath string, keys []string) (map[string]string, error) {
	c.log.Debugf("ReadKeys(ctx, \"%s\", %+v)", servicePath, keys)

	newKeys := c.chamberKeysToLower(keys)
	return c.storage.ReadKeys(ctx, servicePath, newKeys)
}

func (c *ChamberCompatibility) ReadAll(ctx context.Context, servicePath string) (map[string]string, error) {
	c.log.Debugf("ReadAll(ctx, \"%s\")", servicePath)

	config, err := c.storage.ReadAll(ctx, servicePath)
	return c.chamberConfigToUpper(config), err
}

func (c *ChamberCompatibility) ReadAllMetadata(ctx context.Context, servicePath string) ([]KeyMetadata, error) {
	c.log.Debugf("ReadAllMetadata(ctx, \"%s\")", servicePath)

	keyMetadata, err := c.storage.ReadAllMetadata(ctx, servicePath)
	return c.chamberKeyMetaToUpper(keyMetadata), err
}

func (c *ChamberCompatibility) Delete(ctx context.Context, servicePath string, key string) error {
	c.log.Debugf("Delete(ctx, \"%s\", \"%s\")", servicePath, key)

	key = c.chamberKeyToLower(key)
	return c.storage.Delete(ctx, servicePath, key)
}

func (c *ChamberCompatibility) DeleteKeys(ctx context.Context, servicePath string, keys []string) error {
	c.log.Debugf("DeleteKeys(ctx, \"%s\", %+v)", servicePath, keys)

	newKeys := c.chamberKeysToLower(keys)
	return c.storage.DeleteKeys(ctx, servicePath, newKeys)
}

func (c *ChamberCompatibility) MetadataKeys() []string {
	return c.storage.MetadataKeys()
}

func (c *ChamberCompatibility) String() string {
	return fmt.Sprintf("ChamberCompatibility(%s)", c.storage)
}

func (c *ChamberCompatibility) chamberKeyToLower(key string) string {
	newKey := strings.ToLower(key)
	c.log.Debugf("Translating key %s to %s", key, newKey)
	return newKey
}

func (c *ChamberCompatibility) chamberKeyToUpper(key string) string {
	newKey := strings.ToUpper(key)
	c.log.Debugf("Translating key %s to %s", key, newKey)
	return newKey
}

func (c *ChamberCompatibility) chamberKeyMetaToUpper(keyMetadata []KeyMetadata) []KeyMetadata {
	seenKeys := make(map[string]struct{}, len(keyMetadata))

	newKeyMetadata := make([]KeyMetadata, len(keyMetadata))
	for i, keyMeta := range keyMetadata {
		newKeyMetadata[i] = keyMeta
		newKey := c.chamberKeyToUpper(keyMeta.Key)
		newKeyMetadata[i].Key = newKey

		if _, seen := seenKeys[newKey]; seen {
			c.log.Warnf("overwriting %s; multiple instances of same key with different case", newKey)
		}
		seenKeys[newKey] = struct{}{}
	}

	return newKeyMetadata
}

func (c *ChamberCompatibility) chamberKeysToLower(keys []string) []string {
	newKeys := make([]string, len(keys))
	for i, key := range keys {
		newKeys[i] = c.chamberKeyToLower(key)
	}

	return newKeys
}

func (c *ChamberCompatibility) chamberConfigToLower(config map[string]string) map[string]string {
	newConfig := make(map[string]string, len(config))
	for key, value := range config {
		newConfig[c.chamberKeyToLower(key)] = value
	}

	return newConfig
}

func (c *ChamberCompatibility) chamberConfigToUpper(config map[string]string) map[string]string {
	newConfig := make(map[string]string, len(config))
	for key, value := range config {
		newConfig[c.chamberKeyToUpper(key)] = value
	}

	return newConfig
}
