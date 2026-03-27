package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDeps(t *testing.T) {
	assert.Nil(t, parseDeps(""))
	assert.Nil(t, parseDeps("[]"))
	assert.Equal(t, []string{"a", "b"}, parseDeps(`["a","b"]`))
}

func TestParseDepsInvalid(t *testing.T) {
	// Invalid JSON returns nil
	assert.Nil(t, parseDeps("not json"))
}
