package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig(t *testing.T) {
	w := &bytes.Buffer{}
	err := InitConfig(w)
	require.NoError(t, err)
	assert.Equal(t, `error_response: "default.Error"
success_response: "default.Success"`, w.String())
}
