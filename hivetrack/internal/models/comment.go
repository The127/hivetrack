package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/change"
)

type CommentChange int

const (
	CommentChangeBody CommentChange = iota
)

type Comment struct {
	BaseModel
	change.List[CommentChange]

	issueID     uuid.UUID
	authorID    *uuid.UUID
	authorEmail *string
	authorName  *string
	body        string
}

func NewComment(issueID uuid.UUID, authorID *uuid.UUID, authorEmail, authorName *string, body string) *Comment {
	return &Comment{
		BaseModel:   NewBaseModel(),
		List:        change.NewList[CommentChange](),
		issueID:     issueID,
		authorID:    authorID,
		authorEmail: authorEmail,
		authorName:  authorName,
		body:        body,
	}
}

func NewCommentFromDB(id uuid.UUID, createdAt, updatedAt time.Time,
	issueID uuid.UUID, authorID *uuid.UUID, authorEmail, authorName *string, body string) *Comment {
	return &Comment{
		BaseModel:   NewBaseModelFromDB(id, createdAt, updatedAt, nil),
		List:        change.NewList[CommentChange](),
		issueID:     issueID,
		authorID:    authorID,
		authorEmail: authorEmail,
		authorName:  authorName,
		body:        body,
	}
}

func (c *Comment) GetIssueID() uuid.UUID    { return c.issueID }
func (c *Comment) GetAuthorID() *uuid.UUID   { return c.authorID }
func (c *Comment) GetAuthorEmail() *string   { return c.authorEmail }
func (c *Comment) GetAuthorName() *string    { return c.authorName }
func (c *Comment) GetBody() string           { return c.body }

func (c *Comment) SetBody(v string) {
	if c.body == v {
		return
	}
	c.body = v
	c.TrackChange(CommentChangeBody)
}
