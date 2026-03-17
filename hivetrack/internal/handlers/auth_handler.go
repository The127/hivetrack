package handlers

import (
	"net/http"

	"github.com/the127/hivetrack/internal/config"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

func (h *AuthHandler) GetOIDCConfig(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, map[string]string{
		"authority": h.cfg.OIDC.Authority,
		"client_id": h.cfg.OIDC.ClientID,
	})
}
