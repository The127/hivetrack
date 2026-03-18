package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/The127/mediatr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/the127/hivetrack/internal/commands"
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
	slug := mux.Vars(r)["slug"]

	result, err := mediatr.Send[*queries.GetMilestonesResult](r.Context(), h.mediator, queries.GetMilestonesQuery{ProjectSlug: slug})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}

func (h *MilestoneHandler) CreateMilestone(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	var body struct {
		Title       string     `json:"title"`
		Description *string    `json:"description"`
		TargetDate  *time.Time `json:"target_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*commands.CreateMilestoneResult](r.Context(), h.mediator, commands.CreateMilestoneCommand{
		ProjectSlug: slug,
		Title:       body.Title,
		Description: body.Description,
		TargetDate:  body.TargetDate,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusCreated, result)
}

func (h *MilestoneHandler) UpdateMilestone(w http.ResponseWriter, r *http.Request) {
	milestoneIDStr := mux.Vars(r)["id"]
	milestoneID, err := uuid.Parse(milestoneIDStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	var body struct {
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		TargetDate  *time.Time `json:"target_date"`
		Close       *bool      `json:"close"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.UpdateMilestoneResult](r.Context(), h.mediator, commands.UpdateMilestoneCommand{
		MilestoneID: milestoneID,
		Title:       body.Title,
		Description: body.Description,
		TargetDate:  body.TargetDate,
		Close:       body.Close,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *MilestoneHandler) DeleteMilestone(w http.ResponseWriter, r *http.Request) {
	milestoneIDStr := mux.Vars(r)["id"]
	milestoneID, err := uuid.Parse(milestoneIDStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.DeleteMilestoneResult](r.Context(), h.mediator, commands.DeleteMilestoneCommand{
		MilestoneID: milestoneID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}
