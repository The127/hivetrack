package authentication

import (
	"context"
	"fmt"
	"strings"
	"sync"

	oidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/the127/hivetrack/internal/config"
)

// Claims holds the extracted JWT claims.
type Claims struct {
	Sub       string
	Email     string
	Name      string
	AvatarURL string
}

// OIDCVerifier lazily initializes an OIDC provider and verifies tokens.
type OIDCVerifier struct {
	authority     string
	clientID      string
	claimMappings config.OIDCClaimMappings

	mu       sync.Mutex
	verifier *oidc.IDTokenVerifier
}

// NewOIDCVerifier creates a new verifier for the given authority.
func NewOIDCVerifier(authority, clientID string, claimMappings config.OIDCClaimMappings) *OIDCVerifier {
	return &OIDCVerifier{
		authority:     authority,
		clientID:      clientID,
		claimMappings: claimMappings.WithDefaults(),
	}
}

func (v *OIDCVerifier) getVerifier(ctx context.Context) (*oidc.IDTokenVerifier, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.verifier != nil {
		return v.verifier, nil
	}

	provider, err := oidc.NewProvider(ctx, v.authority)
	if err != nil {
		return nil, fmt.Errorf("creating OIDC provider: %w", err)
	}

	v.verifier = provider.Verifier(&oidc.Config{ClientID: v.clientID})
	return v.verifier, nil
}

// VerifyToken verifies the JWT and returns extracted claims.
func (v *OIDCVerifier) VerifyToken(ctx context.Context, token string) (*Claims, error) {
	verifier, err := v.getVerifier(ctx)
	if err != nil {
		return nil, err
	}

	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("verifying token: %w", err)
	}

	var raw map[string]interface{}
	if err := idToken.Claims(&raw); err != nil {
		return nil, fmt.Errorf("extracting claims: %w", err)
	}

	return &Claims{
		Sub:       stringClaim(raw, "sub"),
		Email:     stringClaim(raw, v.claimMappings.Email),
		Name:      resolveName(raw, v.claimMappings.Name),
		AvatarURL: stringClaim(raw, v.claimMappings.Avatar),
	}, nil
}

// resolveName extracts the display name from claims using the configured claim,
// falling back to common OIDC claim names if the configured one is empty.
func resolveName(claims map[string]interface{}, primaryClaim string) string {
	if v := stringClaim(claims, primaryClaim); v != "" {
		return v
	}

	// Fallback chain for common OIDC providers.
	fallbacks := []string{"name", "preferred_username"}
	for _, key := range fallbacks {
		if key == primaryClaim {
			continue
		}
		if v := stringClaim(claims, key); v != "" {
			return v
		}
	}

	// Last resort: combine given_name + family_name.
	given := stringClaim(claims, "given_name")
	family := stringClaim(claims, "family_name")
	if combined := strings.TrimSpace(given + " " + family); combined != "" {
		return combined
	}

	return ""
}

// stringClaim extracts a string value from a claims map, returning "" if missing or wrong type.
func stringClaim(claims map[string]interface{}, key string) string {
	v, ok := claims[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
