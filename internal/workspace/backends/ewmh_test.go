package backends

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryCreate_NoDisplayReturnsNil(t *testing.T) {
	// Invalidate DISPLAY so XGB fails to connect.
	t.Setenv("DISPLAY", ":999")
	be, err := TryCreate()
	assert.Nil(t, be)
	assert.NoError(t, err)
}

func TestTryCreate_NoXauthReturnsNil(t *testing.T) {
	// Ensure DISPLAY is set to a valid-looking display but no authority available.
	t.Setenv("DISPLAY", ":0")
	be, err := TryCreate()
	assert.Nil(t, be)
	assert.NoError(t, err)
}

func TestTryCreate_BackendIsNil(t *testing.T) {
	t.Setenv("DISPLAY", ":999")
	be, err := TryCreate()
	assert.Nil(t, be)
	assert.NoError(t, err)
}
