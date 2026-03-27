package client

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// DefaultTokenPath returns the default path for the token cache file.
func DefaultTokenPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "hivetrack", "credentials.json"), nil
}

// LoadTokenFile loads a cached token from the default path.
func LoadTokenFile() (TokenCache, error) {
	p, err := DefaultTokenPath()
	if err != nil {
		return TokenCache{}, err
	}
	return LoadTokenFileFrom(p)
}

// LoadTokenFileFrom loads a cached token from a specific path.
func LoadTokenFileFrom(path string) (TokenCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TokenCache{}, err
	}
	var tc TokenCache
	if err := json.Unmarshal(data, &tc); err != nil {
		return TokenCache{}, err
	}
	return tc, nil
}

// SaveTokenFile saves a token to the default path.
func SaveTokenFile(tc TokenCache) error {
	p, err := DefaultTokenPath()
	if err != nil {
		return err
	}
	return SaveTokenFileTo(p, tc)
}

// SaveTokenFileTo saves a token to a specific path.
func SaveTokenFileTo(path string, tc TokenCache) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(tc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
