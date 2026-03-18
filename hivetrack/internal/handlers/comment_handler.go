package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/The127/mediatr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/the127/hivetrack/internal/authentication"
	"github.com/the127/hivetrack/internal/commands"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/queries"
)

type CommentHandler struct {
	mediator mediatr.Mediator
}

func NewCommentHandler(m mediatr.Mediator) *CommentHandler {
	return &CommentHandler{mediator: m}
}

func (h *CommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	number, err := strconv.Atoi(vars["number"])
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	q := queries.GetCommentsQuery{
		ProjectSlug: slug,
		IssueNumber: number,
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		q.Limit, _ = strconv.Atoi(l)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		q.Offset, _ = strconv.Atoi(o)
	}

	result, err := mediatr.Send[*queries.GetCommentsResult](r.Context(), h.mediator, q)
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusOK, result)
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	number, err := strconv.Atoi(vars["number"])
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	// Resolve issue ID via query
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
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Body == "" {
		RespondError(w, models.ErrBadRequest)
		return
	}

	actor := authentication.MustGetCurrentUser(r.Context())

	result, err := mediatr.Send[*commands.CreateCommentResult](r.Context(), h.mediator, commands.CreateCommentCommand{
		IssueID:  issueResult.ID,
		AuthorID: actor.ID,
		Body:     body.Body,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusCreated, result)
}

func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID, err := uuid.Parse(vars["comment_id"])
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Body == "" {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.UpdateCommentResult](r.Context(), h.mediator, commands.UpdateCommentCommand{
		CommentID: commentID,
		Body:      body.Body,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentID, err := uuid.Parse(vars["comment_id"])
	if err != nil {
		RespondError(w, models.ErrBadRequest)
		return
	}

	_, err = mediatr.Send[*commands.DeleteCommentResult](r.Context(), h.mediator, commands.DeleteCommentCommand{
		CommentID: commentID,
	})
	if err != nil {
		RespondError(w, err)
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}
