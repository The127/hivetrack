package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOIDCClaimMappings_WithDefaults_AllEmpty(t *testing.T) {
	m := OIDCClaimMappings{}
	got := m.WithDefaults()

	assert.Equal(t, "email", got.Email)
	assert.Equal(t, "name", got.Name)
	assert.Equal(t, "picture", got.Avatar)
}

func TestOIDCClaimMappings_WithDefaults_PreservesCustom(t *testing.T) {
	m := OIDCClaimMappings{
		Email:  "mail",
		Name:   "preferred_username",
		Avatar: "photo",
	}
	got := m.WithDefaults()

	assert.Equal(t, "mail", got.Email)
	assert.Equal(t, "preferred_username", got.Name)
	assert.Equal(t, "photo", got.Avatar)
}

func TestOIDCClaimMappings_WithDefaults_PartialOverride(t *testing.T) {
	m := OIDCClaimMappings{
		Name: "preferred_username",
	}
	got := m.WithDefaults()

	assert.Equal(t, "email", got.Email)
	assert.Equal(t, "preferred_username", got.Name)
	assert.Equal(t, "picture", got.Avatar)
}
