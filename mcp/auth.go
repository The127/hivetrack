package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type tokenCache struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IssuedAt     time.Time `json:"issued_at"`
	Expiry       time.Time `json:"expiry"`
	ServerURL    string    `json:"server_url"`
}

// DeviceFlow holds the state needed to complete an in-progress device authorization flow.
type DeviceFlow struct {
	VerificationURI         string
	VerificationURIComplete string
	UserCode                string

	deviceCode    string
	interval      int
	tokenEndpoint string
	clientID      string
	serverURL     string
}

func cachePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "hivetrack", "mcp-credentials.json"), nil
}

func loadCache() (tokenCache, bool) {
	p, err := cachePath()
	if err != nil {
		return tokenCache{}, false
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return tokenCache{}, false
	}
	var c tokenCache
	if err := json.Unmarshal(data, &c); err != nil {
		return tokenCache{}, false
	}
	return c, true
}

func saveCache(c tokenCache) error {
	p, err := cachePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

// TryToken returns a valid tokenCache from cache or via token refresh.
// Returns (tokenCache{}, nil) if no valid token is available and device flow is needed.
func TryToken(serverURL string) (tokenCache, error) {
	cached, ok := loadCache()
	if !ok || cached.ServerURL != serverURL {
		return tokenCache{}, nil
	}
	if time.Now().Before(cached.Expiry) {
		return cached, nil
	}
	if cached.RefreshToken != "" {
		refreshed, err := tryRefresh(serverURL, cached.RefreshToken)
		if err == nil {
			if saveErr := saveCache(refreshed); saveErr != nil {
				fmt.Fprintf(os.Stderr, "[mcp] warning: failed to save token cache: %v\n", saveErr)
			}
			return refreshed, nil
		}
		fmt.Fprintf(os.Stderr, "[mcp] token refresh failed, device flow required: %v\n", err)
	}
	return tokenCache{}, nil
}

// InitDeviceFlow starts the OIDC device authorization flow and returns immediately
// with the verification URL for the user to open. Call WaitForToken to poll for completion.
func InitDeviceFlow(serverURL string) (*DeviceFlow, error) {
	providerCfg, discovery, err := fetchOIDCEndpoints(serverURL)
	if err != nil {
		return nil, err
	}

	var daResp deviceAuthResponse
	if err := postFormJSON(discovery.DeviceAuthorizationEndpoint, url.Values{
		"client_id": {providerCfg.ClientID},
		"scope":     {"openid offline_access"},
	}, &daResp); err != nil {
		return nil, fmt.Errorf("device authorization request: %w", err)
	}

	interval := daResp.Interval
	if interval <= 0 {
		interval = 5
	}

	return &DeviceFlow{
		VerificationURI:         daResp.VerificationURI,
		VerificationURIComplete: daResp.VerificationURIComplete,
		UserCode:                daResp.UserCode,
		deviceCode:              daResp.DeviceCode,
		interval:                interval,
		tokenEndpoint:           discovery.TokenEndpoint,
		clientID:                providerCfg.ClientID,
		serverURL:               serverURL,
	}, nil
}

// WaitForToken polls the token endpoint until the user completes authentication,
// then saves the token to the cache and returns the full tokenCache.
// Respects context cancellation.
func (f *DeviceFlow) WaitForToken(ctx context.Context) (tokenCache, error) {
	interval := f.interval
	for {
		select {
		case <-ctx.Done():
			return tokenCache{}, ctx.Err()
		case <-time.After(time.Duration(interval) * time.Second):
		}

		var tr tokenResponse
		if err := postFormJSON(f.tokenEndpoint, url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"device_code": {f.deviceCode},
			"client_id":   {f.clientID},
		}, &tr); err != nil {
			return tokenCache{}, fmt.Errorf("polling token endpoint: %w", err)
		}

		switch tr.Error {
		case "":
			now := time.Now()
			c := tokenCache{
				AccessToken:  tr.AccessToken,
				RefreshToken: tr.RefreshToken,
				IssuedAt:     now,
				Expiry:       now.Add(time.Duration(tr.ExpiresIn) * time.Second),
				ServerURL:    f.serverURL,
			}
			if err := saveCache(c); err != nil {
				fmt.Fprintf(os.Stderr, "[mcp] warning: failed to save token cache: %v\n", err)
			}
			return c, nil
		case "authorization_pending":
			// keep waiting
		case "slow_down":
			interval += 5
		case "access_denied":
			return tokenCache{}, fmt.Errorf("device flow: access denied")
		case "expired_token":
			return tokenCache{}, fmt.Errorf("device flow: device code expired")
		default:
			return tokenCache{}, fmt.Errorf("device flow error: %s", tr.Error)
		}
	}
}

// Login runs the OIDC device flow interactively, printing the auth URL to stderr.
// If a valid token already exists, it reports that and returns immediately.
func Login(serverURL string) error {
	tc, err := TryToken(serverURL)
	if err != nil {
		return err
	}
	if tc.AccessToken != "" {
		fmt.Fprintln(os.Stderr, "[mcp] already authenticated (token is valid)")
		return nil
	}

	flow, err := InitDeviceFlow(serverURL)
	if err != nil {
		return err
	}

	authURL := flow.VerificationURIComplete
	if authURL == "" {
		authURL = flow.VerificationURI
	}
	fmt.Fprintf(os.Stderr, "Open this URL to authenticate:\n\n  %s\n\n", authURL)

	_, err = flow.WaitForToken(context.Background())
	return err
}

type oidcProviderConfig struct {
	Authority string `json:"authority"`
	ClientID  string `json:"client_id"`
}

type oidcDiscovery struct {
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
	TokenEndpoint               string `json:"token_endpoint"`
}

type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	Interval                int    `json:"interval"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
}

func getJSON(u string, out any) error {
	resp, err := http.Get(u) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, u)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func postFormJSON(endpoint string, values url.Values, out any) error {
	resp, err := http.PostForm(endpoint, values) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func fetchOIDCEndpoints(serverURL string) (oidcProviderConfig, oidcDiscovery, error) {
	var providerCfg oidcProviderConfig
	if err := getJSON(strings.TrimRight(serverURL, "/")+"/api/v1/auth/oidc-config", &providerCfg); err != nil {
		return providerCfg, oidcDiscovery{}, fmt.Errorf("fetching OIDC config: %w", err)
	}

	var discovery oidcDiscovery
	discoveryURL := strings.TrimRight(providerCfg.Authority, "/") + "/.well-known/openid-configuration"
	if err := getJSON(discoveryURL, &discovery); err != nil {
		return providerCfg, discovery, fmt.Errorf("fetching OIDC discovery: %w", err)
	}

	if discovery.DeviceAuthorizationEndpoint == "" {
		return providerCfg, discovery, fmt.Errorf("OIDC provider does not support device authorization grant")
	}

	return providerCfg, discovery, nil
}

func tryRefresh(serverURL, refreshTok string) (tokenCache, error) {
	providerCfg, discovery, err := fetchOIDCEndpoints(serverURL)
	if err != nil {
		return tokenCache{}, err
	}
	var tr tokenResponse
	if err := postFormJSON(discovery.TokenEndpoint, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshTok},
		"client_id":     {providerCfg.ClientID},
	}, &tr); err != nil {
		return tokenCache{}, err
	}
	if tr.Error != "" {
		return tokenCache{}, fmt.Errorf("token refresh error: %s", tr.Error)
	}
	now := time.Now()
	return tokenCache{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		IssuedAt:     now,
		Expiry:       now.Add(time.Duration(tr.ExpiresIn) * time.Second),
		ServerURL:    serverURL,
	}, nil
}
