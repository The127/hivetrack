package client

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoadTokenFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	tc := TokenCache{
		AccessToken:  "test-token",
		RefreshToken: "refresh-token",
		IssuedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Expiry:       time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC),
		ServerURL:    "http://example.com",
	}

	if err := SaveTokenFileTo(path, tc); err != nil {
		t.Fatalf("SaveTokenFileTo failed: %v", err)
	}

	loaded, err := LoadTokenFileFrom(path)
	if err != nil {
		t.Fatalf("LoadTokenFileFrom failed: %v", err)
	}

	if loaded.AccessToken != tc.AccessToken {
		t.Errorf("expected %s, got %s", tc.AccessToken, loaded.AccessToken)
	}
	if loaded.RefreshToken != tc.RefreshToken {
		t.Errorf("expected %s, got %s", tc.RefreshToken, loaded.RefreshToken)
	}
	if loaded.ServerURL != tc.ServerURL {
		t.Errorf("expected %s, got %s", tc.ServerURL, loaded.ServerURL)
	}
}

func TestLoadTokenFileFrom_missingFile(t *testing.T) {
	_, err := LoadTokenFileFrom("/nonexistent/path/tokens.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestSaveTokenFileTo_createsDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "tokens.json")

	tc := TokenCache{AccessToken: "tok"}
	if err := SaveTokenFileTo(path, tc); err != nil {
		t.Fatalf("SaveTokenFileTo failed: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}
}

func TestSaveTokenFileTo_filePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")

	if err := SaveTokenFileTo(path, TokenCache{AccessToken: "tok"}); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected 0600 permissions, got %o", perm)
	}
}
