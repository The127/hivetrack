package handlers

import (
	"net/http"

	"github.com/The127/mediatr"

	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/queries"
)

type UserHandler struct {
	mediator mediatr.Mediator
}

func NewUserHandler(m mediatr.Mediator) *UserHandler {
	return &UserHandler{mediator: m}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := authentication.MustGetCurrentUser(r.Context())
	RespondJSON(w, http.StatusOK, map[string]any{
		"id":       user.ID,
		"sub":      user.Sub,
		"email":    user.Email,
		"is_admin": user.IsAdmin,
	})
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	result, err := mediatr.Send[*queries.GetUsersResult](r.Context(), h.mediator, queries.GetUsersQuery{})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}
