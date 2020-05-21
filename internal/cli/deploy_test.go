package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/micvbang/confman-go/pkg/confman"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestHandleDefine verifies that handleDefine only deletes keys when it's
// supposed to, i.e. when user accepts it on the user prompt or when user
// prompt is disabled.
func TestHandleDefine(t *testing.T) {
	tests := map[string]struct {
		existingConfig map[string]string
		newConfig      map[string]string
		userInput      io.Reader
		skipPrompt     bool
		err            error
	}{
		"no deletions, skip user prompt": {
			existingConfig: map[string]string{
				"key1": "value1",
			},
			newConfig: map[string]string{
				"key1": "value1",
			},
			skipPrompt: true,
			err:        nil,
		},
		"no deletions, prompt user": {
			existingConfig: map[string]string{
				"key1": "value1",
			},
			newConfig: map[string]string{
				"key1": "value1",
			},
			skipPrompt: false,
			err:        nil,
		},
		"has deletions, user aborts": {
			existingConfig: map[string]string{
				"key1": "value1",
			},
			newConfig: map[string]string{
				"key2": "value2",
			},
			userInput:  strings.NewReader("no"),
			skipPrompt: false,
			err:        ErrUserAbortedKeyDeletion,
		},
		"has deletions, user aborts with random input": {
			existingConfig: map[string]string{
				"key1": "value1",
			},
			newConfig: map[string]string{
				"key2": "value2",
			},
			userInput:  strings.NewReader("lemon"),
			skipPrompt: false,
			err:        ErrUserAbortedKeyDeletion,
		},
		"has deletions, user accepts": {
			existingConfig: map[string]string{
				"key1": "value1",
			},
			newConfig: map[string]string{
				"key2": "value2",
			},
			userInput:  strings.NewReader("yes"),
			skipPrompt: false,
			err:        nil,
		},
		"has deletions, skip user prompt": {
			existingConfig: map[string]string{
				"key1": "value1",
			},
			newConfig: map[string]string{
				"key2": "value2",
			},
			skipPrompt: true,
			err:        nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			cm := confman.MockConfman{}
			cm.On("ReadAll", ctx).Return(test.existingConfig, nil)
			cm.On("DeleteKeys", ctx, mock.Anything).Return(nil)

			userOutput := bytes.NewBuffer(nil)
			err := handleDefine(ctx, &cm, test.userInput, userOutput, test.newConfig, test.skipPrompt)
			require.Equal(t, test.err, err)
		})
	}

}
