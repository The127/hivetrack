package handlers

import (
	"net/http"

	"github.com/The127/mediatr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
)

type MilestoneHandler struct {
	mediator mediatr.Mediator
}

func NewMilestoneHandler(m mediatr.Mediator) *MilestoneHandler {
	return &MilestoneHandler{mediator: m}
}

func (h *MilestoneHandler) ListMilestones(w http.ResponseWriter, r *http.Request) {
	projectIDStr := mux.Vars(r)["project_id"]
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*queries.GetMilestonesResult](r.Context(), h.mediator, queries.GetMilestonesQuery{ProjectID: projectID})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}
