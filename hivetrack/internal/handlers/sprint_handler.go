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

type SprintHandler struct {
	mediator mediatr.Mediator
}

func NewSprintHandler(m mediatr.Mediator) *SprintHandler {
	return &SprintHandler{mediator: m}
}

func (h *SprintHandler) ListSprints(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	result, err := mediatr.Send[*queries.GetSprintsResult](r.Context(), h.mediator, queries.GetSprintsQuery{ProjectSlug: slug})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}

func (h *SprintHandler) CreateSprint(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	var body struct {
		Name      string     `json:"name"`
		Goal      *string    `json:"goal"`
		StartDate *time.Time `json:"start_date"`
		EndDate   *time.Time `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}
	if body.Name == "" {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*commands.CreateSprintResult](r.Context(), h.mediator, commands.CreateSprintCommand{
		ProjectSlug: slug,
		Name:        body.Name,
		Goal:        body.Goal,
		StartDate:   body.StartDate,
		EndDate:     body.EndDate,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusCreated, result)
}

func (h *SprintHandler) UpdateSprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	idStr := vars["id"]

	sprintID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	// Verify sprint belongs to this project.
	sprintsResult, err := mediatr.Send[*queries.GetSprintsResult](r.Context(), h.mediator, queries.GetSprintsQuery{ProjectSlug: slug})
	if err != nil {
		RespondError(w, err)
		return
	}
	found := false
	for _, s := range sprintsResult.Sprints {
		if s.ID == sprintID {
			found = true
			break
		}
	}
	if !found {
		RespondError(w, models.ErrNotFound)
		return
	}

	var body struct {
		Name                     *string              `json:"name"`
		Goal                     *string              `json:"goal"`
		StartDate                *time.Time           `json:"start_date"`
		EndDate                  *time.Time           `json:"end_date"`
		Status                   *models.SprintStatus `json:"status"`
		MoveOpenIssuesToSprintID *uuid.UUID           `json:"move_open_issues_to_sprint_id"`
		Force                    *bool                `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.UpdateSprintResult](r.Context(), h.mediator, commands.UpdateSprintCommand{
		SprintID:                 sprintID,
		Name:                     body.Name,
		Goal:                     body.Goal,
		StartDate:                body.StartDate,
		EndDate:                  body.EndDate,
		Status:                   body.Status,
		MoveOpenIssuesToSprintID: body.MoveOpenIssuesToSprintID,
		Force:                    body.Force != nil && *body.Force,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *SprintHandler) GetSprintBurndown(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	idStr := vars["id"]

	sprintID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*queries.GetSprintBurndownResult](r.Context(), h.mediator, queries.GetSprintBurndownQuery{
		ProjectSlug: slug,
		SprintID:    sprintID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}

func (h *SprintHandler) DeleteSprint(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	idStr := vars["id"]

	sprintID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	// Verify sprint belongs to this project.
	sprintsResult, err := mediatr.Send[*queries.GetSprintsResult](r.Context(), h.mediator, queries.GetSprintsQuery{ProjectSlug: slug})
	if err != nil {
		RespondError(w, err)
		return
	}
	found := false
	for _, s := range sprintsResult.Sprints {
		if s.ID == sprintID {
			found = true
			break
		}
	}
	if !found {
		RespondError(w, models.ErrNotFound)
		return
	}

	_, err = mediatr.Send[*commands.DeleteSprintResult](r.Context(), h.mediator, commands.DeleteSprintCommand{
		SprintID: sprintID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}
