package handlers

import (
	"net/http"

	"github.com/the127/hivetrack/internal/authentication"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := authentication.MustGetCurrentUser(r.Context())
	RespondJSON(w, http.StatusOK, map[string]any{
		"id":      user.ID,
		"sub":     user.Sub,
		"email":   user.Email,
		"is_admin": user.IsAdmin,
	})
}
