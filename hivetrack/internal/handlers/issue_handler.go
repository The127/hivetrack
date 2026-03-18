package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/The127/mediatr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
)

type IssueHandler struct {
	mediator mediatr.Mediator
}

func NewIssueHandler(m mediatr.Mediator) *IssueHandler {
	return &IssueHandler{mediator: m}
}

func (h *IssueHandler) ListIssues(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	q := queries.GetIssuesQuery{ProjectSlug: slug}

	if s := r.URL.Query().Get("status"); s != "" {
		status := models.IssueStatus(s)
		q.Status = &status
	}
	if p := r.URL.Query().Get("priority"); p != "" {
		priority := models.IssuePriority(p)
		q.Priority = &priority
	}
	if t := r.URL.Query().Get("triaged"); t != "" {
		triaged := t == "true"
		q.Triaged = &triaged
	}
	if b := r.URL.Query().Get("backlog"); b != "" {
		inBacklog := b == "true"
		q.InBacklog = &inBacklog
	}
	if text := r.URL.Query().Get("text"); text != "" {
		q.Text = &text
	}
	if tp := r.URL.Query().Get("type"); tp != "" {
		issueType := models.IssueType(tp)
		q.Type = &issueType
	}
	if pid := r.URL.Query().Get("parent_id"); pid != "" {
		if parentID, err := uuid.Parse(pid); err == nil {
			q.ParentID = &parentID
		}
	}
	if np := r.URL.Query().Get("no_parent"); np == "true" {
		v := true
		q.HasNoParent = &v
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ := strconv.Atoi(l)
		q.Limit = limit
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ := strconv.Atoi(o)
		q.Offset = offset
	}

	result, err := mediatr.Send[*queries.GetIssuesResult](r.Context(), h.mediator, q)
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}

func (h *IssueHandler) GetIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	numberStr := vars["number"]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*queries.IssueDetail](r.Context(), h.mediator, queries.GetIssueQuery{
		ProjectSlug: slug,
		Number:      number,
	})
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

func (h *IssueHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	var body struct {
		Title       string                `json:"title"`
		Type        models.IssueType      `json:"type"`
		Priority    *models.IssuePriority `json:"priority"`
		Estimate    *models.IssueEstimate `json:"estimate"`
		Status      *models.IssueStatus   `json:"status"`
		Description *string               `json:"description"`
		AssigneeIDs []uuid.UUID           `json:"assignee_ids"`
		LabelIDs    []uuid.UUID           `json:"label_ids"`
		SprintID    *uuid.UUID            `json:"sprint_id"`
		MilestoneID *uuid.UUID            `json:"milestone_id"`
		ParentID    *uuid.UUID            `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*commands.CreateIssueResult](r.Context(), h.mediator, commands.CreateIssueCommand{
		ProjectSlug: slug,
		Title:       body.Title,
		Type:        body.Type,
		Priority:    body.Priority,
		Estimate:    body.Estimate,
		Status:      body.Status,
		Description: body.Description,
		AssigneeIDs: body.AssigneeIDs,
		LabelIDs:    body.LabelIDs,
		SprintID:    body.SprintID,
		MilestoneID: body.MilestoneID,
		ParentID:    body.ParentID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusCreated, result)
}

func (h *IssueHandler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	numberStr := vars["number"]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	// Get issue ID by number first
	issueResult, err := mediatr.Send[*queries.IssueDetail](r.Context(), h.mediator, queries.GetIssueQuery{
		ProjectSlug: slug,
		Number:      number,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	if issueResult == nil {
		RespondError(w, models.ErrNotFound)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	var body struct {
		Title       *string                 `json:"title"`
		Description *string                 `json:"description"`
		Status      *models.IssueStatus     `json:"status"`
		Priority    *models.IssuePriority   `json:"priority"`
		Estimate    *models.IssueEstimate   `json:"estimate"`
		AssigneeIDs []uuid.UUID             `json:"assignee_ids"`
		LabelIDs    []uuid.UUID             `json:"label_ids"`
		SprintID    *uuid.UUID              `json:"sprint_id"`
		MilestoneID *uuid.UUID              `json:"milestone_id"`
		ParentID    *uuid.UUID              `json:"parent_id"`
		OnHold      *bool                   `json:"on_hold"`
		HoldReason  *models.HoldReason      `json:"hold_reason"`
		HoldNote    *string                 `json:"hold_note"`
		Visibility  *models.IssueVisibility `json:"visibility"`
		Rank        *string                 `json:"rank"`
	}
	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	// Detect explicit null for sprint_id/parent_id to distinguish "clear" from "not provided".
	clearSprintID := false
	clearParentID := false
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &rawFields); err == nil {
		if raw, ok := rawFields["sprint_id"]; ok && string(raw) == "null" {
			clearSprintID = true
		}
		if raw, ok := rawFields["parent_id"]; ok && string(raw) == "null" {
			clearParentID = true
		}
	}

	_, err = mediatr.Send[*commands.UpdateIssueResult](r.Context(), h.mediator, commands.UpdateIssueCommand{
		IssueID:       issueResult.ID,
		Title:         body.Title,
		Description:   body.Description,
		Status:        body.Status,
		Priority:      body.Priority,
		Estimate:      body.Estimate,
		AssigneeIDs:   body.AssigneeIDs,
		LabelIDs:      body.LabelIDs,
		SprintID:      body.SprintID,
		ClearSprintID: clearSprintID,
		MilestoneID:   body.MilestoneID,
		ParentID:      body.ParentID,
		ClearParentID: clearParentID,
		OnHold:        body.OnHold,
		HoldReason:    body.HoldReason,
		HoldNote:      body.HoldNote,
		Visibility:    body.Visibility,
		Rank:          body.Rank,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *IssueHandler) DeleteIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	numberStr := vars["number"]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	issueResult, err := mediatr.Send[*queries.IssueDetail](r.Context(), h.mediator, queries.GetIssueQuery{
		ProjectSlug: slug,
		Number:      number,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	if issueResult == nil {
		RespondError(w, models.ErrNotFound)
		return
	}

	_, err = mediatr.Send[*commands.DeleteIssueResult](r.Context(), h.mediator, commands.DeleteIssueCommand{
		IssueID: issueResult.ID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *IssueHandler) TriageIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	numberStr := vars["number"]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	issueResult, err := mediatr.Send[*queries.IssueDetail](r.Context(), h.mediator, queries.GetIssueQuery{
		ProjectSlug: slug,
		Number:      number,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	if issueResult == nil {
		RespondError(w, models.ErrNotFound)
		return
	}

	var body struct {
		Status      models.IssueStatus `json:"status"`
		SprintID    *uuid.UUID         `json:"sprint_id"`
		MilestoneID *uuid.UUID         `json:"milestone_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.TriageIssueResult](r.Context(), h.mediator, commands.TriageIssueCommand{
		IssueID:     issueResult.ID,
		Status:      body.Status,
		SprintID:    body.SprintID,
		MilestoneID: body.MilestoneID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *IssueHandler) AddChecklistItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	numberStr := vars["number"]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	issueResult, err := mediatr.Send[*queries.IssueDetail](r.Context(), h.mediator, queries.GetIssueQuery{
		ProjectSlug: slug,
		Number:      number,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	if issueResult == nil {
		RespondError(w, models.ErrNotFound)
		return
	}

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*commands.AddChecklistItemResult](r.Context(), h.mediator, commands.AddChecklistItemCommand{
		IssueID: issueResult.ID,
		Text:    body.Text,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusCreated, result)
}

func (h *IssueHandler) UpdateChecklistItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	numberStr := vars["number"]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}
	itemID, err := uuid.Parse(vars["item_id"])
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	issueResult, err := mediatr.Send[*queries.IssueDetail](r.Context(), h.mediator, queries.GetIssueQuery{
		ProjectSlug: slug,
		Number:      number,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	if issueResult == nil {
		RespondError(w, models.ErrNotFound)
		return
	}

	var body struct {
		Text *string `json:"text"`
		Done *bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.UpdateChecklistItemResult](r.Context(), h.mediator, commands.UpdateChecklistItemCommand{
		IssueID: issueResult.ID,
		ItemID:  itemID,
		Text:    body.Text,
		Done:    body.Done,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *IssueHandler) RemoveChecklistItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	numberStr := vars["number"]
	number, err := strconv.Atoi(numberStr)
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}
	itemID, err := uuid.Parse(vars["item_id"])
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	issueResult, err := mediatr.Send[*queries.IssueDetail](r.Context(), h.mediator, queries.GetIssueQuery{
		ProjectSlug: slug,
		Number:      number,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	if issueResult == nil {
		RespondError(w, models.ErrNotFound)
		return
	}

	_, err = mediatr.Send[*commands.RemoveChecklistItemResult](r.Context(), h.mediator, commands.RemoveChecklistItemCommand{
		IssueID: issueResult.ID,
		ItemID:  itemID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *IssueHandler) GetMyIssues(w http.ResponseWriter, r *http.Request) {
	result, err := mediatr.Send[*queries.GetMyIssuesResult](r.Context(), h.mediator, queries.GetMyIssuesQuery{})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}
