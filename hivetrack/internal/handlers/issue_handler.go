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
	params := r.URL.Query()

	q := queries.GetIssuesQuery{
		ProjectSlug:    slug,
		Status:         typedParam[models.IssueStatus](params, "status"),
		Priority:       typedParam[models.IssuePriority](params, "priority"),
		Type:           typedParam[models.IssueType](params, "type"),
		Triaged:        queryParamBool(params, "triaged"),
		Refined:        queryParamBool(params, "refined"),
		InBacklog:      queryParamBool(params, "backlog"),
		OnHold:         queryParamBool(params, "on_hold"),
		HasNoParent:    queryParamBool(params, "no_parent"),
		Text:           queryParam(params, "text"),
		ParentID:       queryParamUUID(params, "parent_id"),
		SprintID:       queryParamUUID(params, "sprint_id"),
		AssigneeID:     queryParamUUID(params, "assignee_id"),
		LabelID:        queryParamUUID(params, "label_id"),
		ExcludeLabelID: queryParamUUID(params, "exclude_label_id"),
		Limit:          queryParamInt(params, "limit"),
		Offset:         queryParamInt(params, "offset"),
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
		Title        *string                 `json:"title"`
		Description  *string                 `json:"description"`
		Status       *models.IssueStatus     `json:"status"`
		Priority     *models.IssuePriority   `json:"priority"`
		Estimate     *models.IssueEstimate   `json:"estimate"`
		AssigneeIDs  []uuid.UUID             `json:"assignee_ids"`
		LabelIDs     []uuid.UUID             `json:"label_ids"`
		SprintID     *uuid.UUID              `json:"sprint_id"`
		MilestoneID  *uuid.UUID              `json:"milestone_id"`
		ParentID     *uuid.UUID              `json:"parent_id"`
		OnHold       *bool                   `json:"on_hold"`
		HoldReason   *models.HoldReason      `json:"hold_reason"`
		HoldNote     *string                 `json:"hold_note"`
		Visibility   *models.IssueVisibility `json:"visibility"`
		Rank         *string                 `json:"rank"`
		OwnerID      *uuid.UUID              `json:"owner_id"`
		CancelReason *string                 `json:"cancel_reason"`
	}
	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	// Detect explicit null for sprint_id/parent_id/owner_id to distinguish "clear" from "not provided".
	clearSprintID := false
	clearParentID := false
	clearOwnerID := false
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &rawFields); err == nil {
		if raw, ok := rawFields["sprint_id"]; ok && string(raw) == "null" {
			clearSprintID = true
		}
		if raw, ok := rawFields["parent_id"]; ok && string(raw) == "null" {
			clearParentID = true
		}
		if raw, ok := rawFields["owner_id"]; ok && string(raw) == "null" {
			clearOwnerID = true
		}
	}

	ctx := commands.ContextWithMediator(r.Context(), h.mediator)

	_, err = mediatr.Send[*commands.UpdateIssueResult](ctx, h.mediator, commands.UpdateIssueCommand{
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
		OwnerID:       body.OwnerID,
		ClearOwnerID:  clearOwnerID,
		CancelReason:  body.CancelReason,
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
		Status      models.IssueStatus    `json:"status"`
		SprintID    *uuid.UUID            `json:"sprint_id"`
		MilestoneID *uuid.UUID            `json:"milestone_id"`
		Priority    *models.IssuePriority `json:"priority"`
		Estimate    *models.IssueEstimate `json:"estimate"`
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
		Priority:    body.Priority,
		Estimate:    body.Estimate,
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

func (h *IssueHandler) SplitIssue(w http.ResponseWriter, r *http.Request) {
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
		Titles []string `json:"titles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*commands.SplitIssueResult](r.Context(), h.mediator, commands.SplitIssueCommand{
		IssueID:   issueResult.ID,
		NewTitles: body.Titles,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusCreated, result)
}

func (h *IssueHandler) AddIssueLink(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	number, err := strconv.Atoi(vars["number"])
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	sourceIssue, err := mediatr.Send[*queries.IssueDetail](r.Context(), h.mediator, queries.GetIssueQuery{
		ProjectSlug: slug,
		Number:      number,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	if sourceIssue == nil {
		RespondError(w, models.ErrNotFound)
		return
	}

	var body struct {
		LinkType     models.LinkType `json:"link_type"`
		TargetNumber int             `json:"target_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	targetIssue, err := mediatr.Send[*queries.IssueDetail](r.Context(), h.mediator, queries.GetIssueQuery{
		ProjectSlug: slug,
		Number:      body.TargetNumber,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	if targetIssue == nil {
		RespondError(w, models.ErrNotFound)
		return
	}

	_, err = mediatr.Send[*commands.CreateIssueLinkResult](r.Context(), h.mediator, commands.CreateIssueLinkCommand{
		SourceIssueID: sourceIssue.ID,
		TargetIssueID: targetIssue.ID,
		LinkType:      body.LinkType,
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

func (h *IssueHandler) BatchUpdateIssues(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	projectResult, err := mediatr.Send[*queries.GetProjectResult](r.Context(), h.mediator, queries.GetProjectQuery{Slug: slug})
	if err != nil {
		RespondError(w, err)
		return
	}
	if projectResult == nil {
		RespondError(w, models.ErrNotFound)
		return
	}

	var body struct {
		Numbers       []int                 `json:"numbers"`
		Status        *models.IssueStatus   `json:"status"`
		Priority      *models.IssuePriority `json:"priority"`
		Estimate      *models.IssueEstimate `json:"estimate"`
		AssigneeIDs   []uuid.UUID           `json:"assignee_ids"`
		LabelIDs      []uuid.UUID           `json:"label_ids"`
		SprintID      *uuid.UUID            `json:"sprint_id"`
		ClearSprintID bool                  `json:"clear_sprint_id"`
		MilestoneID   *uuid.UUID            `json:"milestone_id"`
		OnHold        *bool                 `json:"on_hold"`
		HoldReason    *models.HoldReason    `json:"hold_reason"`
		HoldNote      *string               `json:"hold_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	result, err := mediatr.Send[*commands.BatchUpdateIssuesResult](r.Context(), h.mediator, commands.BatchUpdateIssuesCommand{
		ProjectID:     projectResult.ID,
		IssueNumbers:  body.Numbers,
		Status:        body.Status,
		Priority:      body.Priority,
		Estimate:      body.Estimate,
		AssigneeIDs:   body.AssigneeIDs,
		LabelIDs:      body.LabelIDs,
		SprintID:      body.SprintID,
		ClearSprintID: body.ClearSprintID,
		MilestoneID:   body.MilestoneID,
		OnHold:        body.OnHold,
		HoldReason:    body.HoldReason,
		HoldNote:      body.HoldNote,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}
