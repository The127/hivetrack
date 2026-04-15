package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/The127/mediatr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/events"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
)

type RefinementHandler struct {
	mediator mediatr.Mediator
	broker   *events.RefinementBroker
}

func NewRefinementHandler(m mediatr.Mediator, broker *events.RefinementBroker) *RefinementHandler {
	return &RefinementHandler{mediator: m, broker: broker}
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

func (h *RefinementHandler) AdvancePhase(w http.ResponseWriter, r *http.Request) {
	issueID, err := h.resolveIssueID(r)
	if err != nil {
		RespondError(w, err)
		return
	}

	var body struct {
		TargetPhase string `json:"target_phase"`
	}
	// Body is optional — empty body means advance to next phase
	_ = json.NewDecoder(r.Body).Decode(&body)

	result, err := mediatr.Send[*commands.AdvanceRefinementPhaseResult](r.Context(), h.mediator, commands.AdvanceRefinementPhaseCommand{
		IssueID:     issueID,
		TargetPhase: body.TargetPhase,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}

// StreamSession opens a Server-Sent Events stream that notifies the client
// whenever the refinement session for this issue changes. The payload is just
// an opaque "updated" nudge — the client is expected to refetch the session
// through GetSession on every tick. Polling on the client side remains as a
// safety net in case a tick is missed.
func (h *RefinementHandler) StreamSession(w http.ResponseWriter, r *http.Request) {
	issueID, err := h.resolveIssueID(r)
	if err != nil {
		RespondError(w, err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Disable proxy buffering (nginx in particular) so events arrive live.
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	ch, unsub := h.broker.Subscribe(issueID)
	defer unsub()

	writeEvent := func(payload string) bool {
		if _, err := fmt.Fprint(w, payload); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	// Initial tick: the client connected and should fetch state once right away.
	if !writeEvent("data: updated\n\n") {
		return
	}

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case _, open := <-ch:
			if !open {
				return
			}
			if !writeEvent("data: updated\n\n") {
				return
			}
		case <-heartbeat.C:
			if !writeEvent(": ping\n\n") {
				return
			}
		}
	}
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
