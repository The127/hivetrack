package main

import (
	"context"
	"fmt"
	"os"

	"github.com/the127/hivetrack/client"
	"github.com/urfave/cli/v2"
)

// mustClient returns an authenticated client or exits with a helpful message.
// Commands that require auth call this at the top of their Action.
func mustClient(c *cli.Context) (*client.Client, error) {
	serverURL := c.String("server")
	if serverURL == "" {
		cfg, err := loadConfig()
		if err != nil || cfg.ServerURL == "" {
			return nil, cli.Exit("not configured: run 'ht login <server-url>'", 1)
		}
		serverURL = cfg.ServerURL
	}

	var tokenFn func(context.Context) (string, error)

	if tok := os.Getenv("HIVETRACK_TOKEN"); tok != "" {
		tokenFn = func(_ context.Context) (string, error) { return tok, nil }
	} else {
		tc, _ := client.LoadTokenFile()
		if tc.AccessToken == "" && tc.RefreshToken == "" {
			return nil, cli.Exit("not authenticated: run 'ht login'", 1)
		}
		provider := client.NewCachingTokenProvider(
			&noopProvider{},
			client.RealClock,
			serverURL,
			tc,
			0.1,
		)
		tokenFn = func(ctx context.Context) (string, error) {
			t, err := provider.ProvideToken(ctx)
			if err != nil {
				return "", fmt.Errorf("authentication failed (run 'ht login'): %w", err)
			}
			return t.AccessToken, nil
		}
	}

	return client.New(serverURL, tokenFn), nil
}

// noopProvider is used as the fallback inner provider for CachingTokenProvider.
// If the cached token is expired and can't be refreshed, the user must re-login manually.
type noopProvider struct{}

func (n *noopProvider) ProvideToken(_ context.Context) (client.TokenCache, error) {
	return client.TokenCache{}, fmt.Errorf("session expired: run 'ht login'")
}

var loginCmd = &cli.Command{
	Name:      "login",
	Usage:     "Authenticate with a Hivetrack instance",
	ArgsUsage: "<server-url>",
	Action: func(c *cli.Context) error {
		serverURL := c.Args().First()
		if serverURL == "" {
			serverURL = c.String("server")
		}
		if serverURL == "" {
			return cli.Exit("usage: ht login <server-url>", 1)
		}

		flow, err := client.InitDeviceFlow(serverURL)
		if err != nil {
			return cli.Exit(fmt.Sprintf("failed to start login: %v", err), 1)
		}

		authURL := flow.VerificationURIComplete
		if authURL == "" {
			authURL = flow.VerificationURI
		}
		fmt.Fprintf(c.App.Writer, "Open this URL to authenticate:\n\n  %s\n\nWaiting...\n", authURL)

		if _, err := flow.WaitForToken(c.Context); err != nil {
			return cli.Exit(fmt.Sprintf("login failed: %v", err), 1)
		}

		if err := saveConfig(Config{ServerURL: serverURL}); err != nil {
			fmt.Fprintf(c.App.ErrWriter, "warning: could not save config: %v\n", err)
		}

		fmt.Fprintln(c.App.Writer, "Authenticated.")
		return nil
	},
}

var logoutCmd = &cli.Command{
	Name:  "logout",
	Usage: "Clear stored credentials",
	Action: func(c *cli.Context) error {
		p, err := client.DefaultTokenPath()
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return cli.Exit(fmt.Sprintf("could not remove credentials: %v", err), 1)
		}
		fmt.Fprintln(c.App.Writer, "Logged out.")
		return nil
	},
}
