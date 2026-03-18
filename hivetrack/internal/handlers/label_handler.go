package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/The127/mediatr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
)

type LabelHandler struct {
	mediator mediatr.Mediator
}

func NewLabelHandler(m mediatr.Mediator) *LabelHandler {
	return &LabelHandler{mediator: m}
}

func (h *LabelHandler) ListLabels(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	result, err := mediatr.Send[*queries.GetLabelsResult](r.Context(), h.mediator, queries.GetLabelsQuery{ProjectSlug: slug})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}

func (h *LabelHandler) CreateLabel(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	var body struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*commands.CreateLabelResult](r.Context(), h.mediator, commands.CreateLabelCommand{
		ProjectSlug: slug,
		Name:        body.Name,
		Color:       body.Color,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusCreated, result)
}

func (h *LabelHandler) UpdateLabel(w http.ResponseWriter, r *http.Request) {
	labelIDStr := mux.Vars(r)["label_id"]
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	var body struct {
		Name  *string `json:"name"`
		Color *string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.UpdateLabelResult](r.Context(), h.mediator, commands.UpdateLabelCommand{
		LabelID: labelID,
		Name:    body.Name,
		Color:   body.Color,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *LabelHandler) DeleteLabel(w http.ResponseWriter, r *http.Request) {
	labelIDStr := mux.Vars(r)["label_id"]
	labelID, err := uuid.Parse(labelIDStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.DeleteLabelResult](r.Context(), h.mediator, commands.DeleteLabelCommand{
		LabelID: labelID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}
