package handlers

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNullableInt(t *testing.T) {
	t.Run("absent field returns nil (no update)", func(t *testing.T) {
		result, err := parseNullableInt(json.RawMessage(nil))
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("explicit null clears the value", func(t *testing.T) {
		result, err := parseNullableInt(json.RawMessage(`null`))
		require.NoError(t, err)
		require.NotNil(t, result) // outer pointer present
		assert.Nil(t, *result)    // inner pointer nil = clear
	})

	t.Run("integer sets the value", func(t *testing.T) {
		result, err := parseNullableInt(json.RawMessage(`5`))
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, *result)
		assert.Equal(t, 5, **result)
	})

	t.Run("invalid value returns error", func(t *testing.T) {
		_, err := parseNullableInt(json.RawMessage(`"not-a-number"`))
		assert.Error(t, err)
	})
}

func TestParseNullableUUID(t *testing.T) {
	t.Run("absent field returns no update", func(t *testing.T) {
		value, clear, err := parseNullableUUID(json.RawMessage(nil))
		require.NoError(t, err)
		assert.Nil(t, value)
		assert.False(t, clear)
	})

	t.Run("explicit null clears the value", func(t *testing.T) {
		value, clear, err := parseNullableUUID(json.RawMessage(`null`))
		require.NoError(t, err)
		assert.Nil(t, value)
		assert.True(t, clear)
	})

	t.Run("valid UUID sets the value", func(t *testing.T) {
		id := uuid.New()
		value, clear, err := parseNullableUUID(json.RawMessage(`"` + id.String() + `"`))
		require.NoError(t, err)
		require.NotNil(t, value)
		assert.Equal(t, id, *value)
		assert.False(t, clear)
	})

	t.Run("invalid value returns error", func(t *testing.T) {
		_, _, err := parseNullableUUID(json.RawMessage(`"not-a-uuid"`))
		assert.Error(t, err)
	})
}
