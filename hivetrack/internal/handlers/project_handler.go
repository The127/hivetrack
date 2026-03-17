package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/The127/mediatr"
	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
)

type ProjectHandler struct {
	mediator mediatr.Mediator
}

func NewProjectHandler(m mediatr.Mediator) *ProjectHandler {
	return &ProjectHandler{mediator: m}
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	result, err := mediatr.Send[*queries.GetProjectsResult](r.Context(), h.mediator, queries.GetProjectsQuery{})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	result, err := mediatr.Send[*queries.GetProjectResult](r.Context(), h.mediator, queries.GetProjectQuery{Slug: slug})
	if err != nil {
		RespondError(w, err)
		return
	}
	if result == nil {
		RespondError(w, models.ErrNotFound)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Slug        string                  `json:"slug"`
		Name        string                  `json:"name"`
		Archetype   models.ProjectArchetype `json:"archetype"`
		Description *string                 `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*commands.CreateProjectResult](r.Context(), h.mediator, commands.CreateProjectCommand{
		Slug:        body.Slug,
		Name:        body.Name,
		Archetype:   body.Archetype,
		Description: body.Description,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusCreated, result)
}

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Archived    *bool   `json:"archived"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.UpdateProjectResult](r.Context(), h.mediator, commands.UpdateProjectCommand{
		ID:          id,
		Name:        body.Name,
		Description: body.Description,
		Archived:    body.Archived,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.DeleteProjectResult](r.Context(), h.mediator, commands.DeleteProjectCommand{ID: id})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}
