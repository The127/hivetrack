package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringClaim(t *testing.T) {
	claims := map[string]interface{}{
		"sub":   "user-123",
		"email": "alice@example.com",
		"count": 42,
	}

	assert.Equal(t, "user-123", stringClaim(claims, "sub"))
	assert.Equal(t, "alice@example.com", stringClaim(claims, "email"))
	assert.Equal(t, "", stringClaim(claims, "missing"))
	assert.Equal(t, "", stringClaim(claims, "count"), "non-string values return empty")
}

func TestResolveName_PrimaryClaim(t *testing.T) {
	claims := map[string]interface{}{
		"name":               "Alice Smith",
		"preferred_username": "alice",
	}

	got := resolveName(claims, "name")
	assert.Equal(t, "Alice Smith", got)
}

func TestResolveName_CustomPrimaryClaim(t *testing.T) {
	claims := map[string]interface{}{
		"display_name":       "Custom Name",
		"name":               "Standard Name",
		"preferred_username": "alice",
	}

	got := resolveName(claims, "display_name")
	assert.Equal(t, "Custom Name", got)
}

func TestResolveName_FallbackToPreferredUsername(t *testing.T) {
	claims := map[string]interface{}{
		"sub":                "user-123",
		"preferred_username": "alice",
	}

	got := resolveName(claims, "name")
	assert.Equal(t, "alice", got)
}

func TestResolveName_FallbackToGivenAndFamilyName(t *testing.T) {
	claims := map[string]interface{}{
		"sub":         "user-123",
		"given_name":  "Alice",
		"family_name": "Smith",
	}

	got := resolveName(claims, "name")
	assert.Equal(t, "Alice Smith", got)
}

func TestResolveName_FallbackToGivenNameOnly(t *testing.T) {
	claims := map[string]interface{}{
		"sub":        "user-123",
		"given_name": "Alice",
	}

	got := resolveName(claims, "name")
	assert.Equal(t, "Alice", got)
}

func TestResolveName_EmptyClaims(t *testing.T) {
	claims := map[string]interface{}{
		"sub": "user-123",
	}

	got := resolveName(claims, "name")
	assert.Equal(t, "", got)
}

func TestResolveName_CustomClaimFallsBackToStandard(t *testing.T) {
	// Custom claim is configured but not present; should fall back to standard "name".
	claims := map[string]interface{}{
		"name": "Standard Name",
	}

	got := resolveName(claims, "custom_name")
	assert.Equal(t, "Standard Name", got)
}

func TestResolveName_PreferredUsernameAsConfigured(t *testing.T) {
	// When preferred_username is configured as primary, don't try it again in fallback.
	claims := map[string]interface{}{
		"preferred_username": "alice",
	}

	got := resolveName(claims, "preferred_username")
	assert.Equal(t, "alice", got)
}
