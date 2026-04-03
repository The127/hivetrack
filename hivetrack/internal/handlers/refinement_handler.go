package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/The127/mediatr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
)

type RefinementHandler struct {
	mediator mediatr.Mediator
}

func NewRefinementHandler(m mediatr.Mediator) *RefinementHandler {
	return &RefinementHandler{mediator: m}
}

func (h *RefinementHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	issueID, err := h.resolveIssueID(r)
	if err != nil {
		RespondError(w, err)
		return
	}

	result, err := mediatr.Send[*commands.StartRefinementSessionResult](r.Context(), h.mediator, commands.StartRefinementSessionCommand{
		IssueID: issueID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusCreated, result)
}

func (h *RefinementHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	issueID, err := h.resolveIssueID(r)
	if err != nil {
		RespondError(w, err)
		return
	}

	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}
	if body.Content == "" {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.SendRefinementMessageResult](r.Context(), h.mediator, commands.SendRefinementMessageCommand{
		IssueID: issueID,
		Content: body.Content,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *RefinementHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	number, err := strconv.Atoi(vars["number"])
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*queries.RefinementSessionDetail](r.Context(), h.mediator, queries.GetRefinementSessionQuery{
		ProjectSlug: slug,
		IssueNumber: number,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	if result == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("null"))
		return
	}
	RespondJSON(w, http.StatusOK, result)
}

func (h *RefinementHandler) AcceptProposal(w http.ResponseWriter, r *http.Request) {
	issueID, err := h.resolveIssueID(r)
	if err != nil {
		RespondError(w, err)
		return
	}

	_, err = mediatr.Send[*commands.AcceptRefinementProposalResult](r.Context(), h.mediator, commands.AcceptRefinementProposalCommand{
		IssueID: issueID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

// resolveIssueID resolves the issue UUID from route params (slug + number).
func (h *RefinementHandler) resolveIssueID(r *http.Request) (uuid.UUID, error) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	number, err := strconv.Atoi(vars["number"])
	if err != nil {
		return uuid.UUID{}, models.ErrBadRequest
	}

	issue, err := mediatr.Send[*queries.IssueDetail](r.Context(), h.mediator, queries.GetIssueQuery{
		ProjectSlug: slug,
		Number:      number,
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	if issue == nil {
		return uuid.UUID{}, models.ErrNotFound
	}
	return issue.ID, nil
}
