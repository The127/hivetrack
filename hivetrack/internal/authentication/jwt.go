package authentication

import (
	"context"
	"fmt"
	"sync"

	oidc "github.com/coreos/go-oidc/v3/oidc"
)

// Claims holds the extracted JWT claims.
type Claims struct {
	Sub   string
	Email string
	Name  string
}

// OIDCVerifier lazily initializes an OIDC provider and verifies tokens.
type OIDCVerifier struct {
	authority string
	clientID  string

	mu       sync.Mutex
	verifier *oidc.IDTokenVerifier
}

// NewOIDCVerifier creates a new verifier for the given authority.
func NewOIDCVerifier(authority, clientID string) *OIDCVerifier {
	return &OIDCVerifier{
		authority: authority,
		clientID:  clientID,
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

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("extracting claims: %w", err)
	}

	return &Claims{
		Sub:   claims.Sub,
		Email: claims.Email,
		Name:  claims.Name,
	}, nil
}
