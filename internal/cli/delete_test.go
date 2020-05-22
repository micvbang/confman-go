package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/micvbang/confman-go/pkg/confman"
	"github.com/micvbang/confman-go/pkg/logger"
	"github.com/micvbang/confman-go/pkg/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestDeleteCommandSpecificKeys verifies that DeleteCommand deletes exactly
// the keys requested when DeleteAll is false.
func TestDeleteCommandSpecificKeys(t *testing.T) {
	ctx := context.Background()
	logger := logger.LogrusWrapper{Logger: logrus.New()}
	servicePath := "/email-dispatch/runtime/development"
	expectedKeysDeleted := []string{"key1", "key2", "key3"}

	s := &storage.MockStorage{}
	s.On("DeleteKeys", ctx, servicePath, expectedKeysDeleted).Return(nil)
	defer s.AssertExpectations(t)

	deleteCommandInput := DeleteCommandInput{
		ServicePath: servicePath,
		Keys:        expectedKeysDeleted,
		DeleteAll:   false,
		Format:      formatText,
	}

	outputBuf := bytes.NewBuffer(nil)
	err := DeleteCommand(ctx, deleteCommandInput, outputBuf, logger, s)
	require.NoError(t, err)
	output := outputBuf.String()

	require.Equal(t, len(expectedKeysDeleted), strings.Count(output, servicePath))
	for _, key := range expectedKeysDeleted {
		require.Equal(t, 1, strings.Count(output, key), fmt.Sprintf("expected to find key '%s'", key))
	}
}

// TestDeleteCommandAllKeys verifies that DeleteCommand ignores the input keys
// and deletes all keys on the given service path when DeleteAll is true.
func TestDeleteCommandAllKeys(t *testing.T) {
	confman.ChamberCompatible = false

	ctx := context.Background()
	logger := logger.LogrusWrapper{Logger: logrus.New()}
	servicePath := "/email-dispatch/runtime/development"
	servicePathConfig := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	s := &storage.MockStorage{}
	s.On("ReadAll", ctx, servicePath).Return(servicePathConfig, nil)

	// Since maps are iterated in random order, we can't know what the exact
	// input to DeleteKeys. We therefore have to do some more work to assert
	// that exactly the expected keys are deleted.
	s.On("DeleteKeys", ctx, servicePath, mock.Anything).Return(nil).Once()
	defer func() {
		var call mock.Call
		for _, c := range s.Calls {
			if c.Method == "DeleteKeys" {
				call = c
			}
		}

		deletedKeys := call.Arguments.Get(2).([]string)
		require.Equal(t, len(deletedKeys), len(servicePathConfig))
		for _, key := range deletedKeys {
			_, exists := servicePathConfig[key]
			require.True(t, exists)
		}
	}()
	defer s.AssertExpectations(t)

	deleteCommandInput := DeleteCommandInput{
		ServicePath: servicePath,
		Keys:        []string{"ignore-me", "ignore-me-too"},
		DeleteAll:   true,
		Format:      formatText,
	}

	outputBuf := bytes.NewBuffer(nil)
	err := DeleteCommand(ctx, deleteCommandInput, outputBuf, logger, s)
	require.NoError(t, err)
	output := outputBuf.String()

	require.Equal(t, len(servicePathConfig), strings.Count(output, servicePath))
	for key := range servicePathConfig {
		require.Equal(t, 1, strings.Count(output, key), fmt.Sprintf("expected to find key '%s' in output '%s'", key, output))
	}
}
