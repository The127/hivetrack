package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// DeviceFlow holds the state for an in-progress OIDC device authorization flow.
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

// InitDeviceFlow starts the OIDC device authorization flow.
// Returns immediately with the verification URL for the user to open.
// Call WaitForToken to poll for completion.
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

// WaitForToken polls the token endpoint until the user completes authentication.
// Respects context cancellation.
func (f *DeviceFlow) WaitForToken(ctx context.Context) (TokenCache, error) {
	interval := f.interval
	for {
		select {
		case <-ctx.Done():
			return TokenCache{}, ctx.Err()
		case <-time.After(time.Duration(interval) * time.Second):
		}

		var tr tokenResponse
		if err := postFormJSON(f.tokenEndpoint, url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"device_code": {f.deviceCode},
			"client_id":   {f.clientID},
		}, &tr); err != nil {
			return TokenCache{}, fmt.Errorf("polling token endpoint: %w", err)
		}

		switch tr.Error {
		case "":
			now := time.Now()
			tc := TokenCache{
				AccessToken:  tr.AccessToken,
				RefreshToken: tr.RefreshToken,
				IssuedAt:     now,
				Expiry:       now.Add(time.Duration(tr.ExpiresIn) * time.Second),
				ServerURL:    f.serverURL,
			}
			_ = SaveTokenFile(tc)
			return tc, nil
		case "authorization_pending":
			// keep waiting
		case "slow_down":
			interval += 5
		case "access_denied":
			return TokenCache{}, fmt.Errorf("device flow: access denied")
		case "expired_token":
			return TokenCache{}, fmt.Errorf("device flow: device code expired")
		default:
			return TokenCache{}, fmt.Errorf("device flow error: %s", tr.Error)
		}
	}
}

// TryRefresh attempts to refresh a token using the refresh token.
func TryRefresh(serverURL, refreshTok string) (TokenCache, error) {
	providerCfg, discovery, err := fetchOIDCEndpoints(serverURL)
	if err != nil {
		return TokenCache{}, err
	}
	var tr tokenResponse
	if err := postFormJSON(discovery.TokenEndpoint, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshTok},
		"client_id":     {providerCfg.ClientID},
	}, &tr); err != nil {
		return TokenCache{}, err
	}
	if tr.Error != "" {
		return TokenCache{}, fmt.Errorf("token refresh error: %s", tr.Error)
	}
	now := time.Now()
	return TokenCache{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		IssuedAt:     now,
		Expiry:       now.Add(time.Duration(tr.ExpiresIn) * time.Second),
		ServerURL:    serverURL,
	}, nil
}

// Login runs the OIDC device flow interactively. If a valid cached token exists,
// returns immediately. Otherwise prints the auth URL and waits for completion.
func Login(serverURL string) error {
	tc, err := LoadTokenFile()
	if err == nil && tc.ServerURL == serverURL && time.Now().Before(tc.Expiry) {
		fmt.Fprintf(os.Stderr, "[hivetrack] already authenticated (token is valid)\n")
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

// OIDC types

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

