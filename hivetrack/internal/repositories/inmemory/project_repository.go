package inmemory

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type ProjectRepository struct {
	tracker        *change.Tracker
	byID           map[uuid.UUID]*models.Project
	bySlug         map[string]*models.Project
	members        map[uuid.UUID]map[uuid.UUID]*models.ProjectMember // projectID -> userID -> member
	counters       map[uuid.UUID]int
	sprintCounters map[uuid.UUID]int
}

func NewProjectRepository(tracker *change.Tracker) *ProjectRepository {
	return &ProjectRepository{
		tracker:        tracker,
		byID:           make(map[uuid.UUID]*models.Project),
		bySlug:         make(map[string]*models.Project),
		members:        make(map[uuid.UUID]map[uuid.UUID]*models.ProjectMember),
		counters:       make(map[uuid.UUID]int),
		sprintCounters: make(map[uuid.UUID]int),
	}
}

func (r *ProjectRepository) Insert(project *models.Project) {
	r.tracker.Add(change.NewEntry(0, project, change.Added))
}

func (r *ProjectRepository) Update(project *models.Project) {
	r.tracker.Add(change.NewEntry(0, project, change.Updated))
}

func (r *ProjectRepository) Delete(project *models.Project) {
	r.tracker.Add(change.NewEntry(0, project, change.Deleted))
}

func (r *ProjectRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Project, error) {
	p, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (r *ProjectRepository) GetBySlug(_ context.Context, slug string) (*models.Project, error) {
	p, ok := r.bySlug[slug]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (r *ProjectRepository) List(_ context.Context, filter *repositories.ProjectFilter) ([]*models.Project, error) {
	var result []*models.Project
	for _, p := range r.byID {
		if filter.MemberUserID != nil && !filter.IsAdmin {
			projectMembers, ok := r.members[p.GetId()]
			if !ok {
				continue
			}
			if _, isMember := projectMembers[*filter.MemberUserID]; !isMember {
				continue
			}
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *ProjectRepository) AddMember(_ context.Context, member *models.ProjectMember) error {
	if _, ok := r.members[member.ProjectID]; !ok {
		r.members[member.ProjectID] = make(map[uuid.UUID]*models.ProjectMember)
	}
	cp := *member
	r.members[member.ProjectID][member.UserID] = &cp
	return nil
}

func (r *ProjectRepository) UpdateMember(_ context.Context, member *models.ProjectMember) error {
	projectMembers, ok := r.members[member.ProjectID]
	if !ok {
		return fmt.Errorf("project %s not found: %w", member.ProjectID, models.ErrNotFound)
	}
	cp := *member
	projectMembers[member.UserID] = &cp
	return nil
}

func (r *ProjectRepository) RemoveMember(_ context.Context, projectID, userID uuid.UUID) error {
	projectMembers, ok := r.members[projectID]
	if !ok {
		return nil
	}
	delete(projectMembers, userID)
	return nil
}

func (r *ProjectRepository) GetMember(_ context.Context, projectID, userID uuid.UUID) (*models.ProjectMember, error) {
	projectMembers, ok := r.members[projectID]
	if !ok {
		return nil, nil
	}
	m, ok := projectMembers[userID]
	if !ok {
		return nil, nil
	}
	cp := *m
	return &cp, nil
}

func (r *ProjectRepository) ListMembers(_ context.Context, projectID uuid.UUID) ([]*models.ProjectMember, error) {
	projectMembers, ok := r.members[projectID]
	if !ok {
		return nil, nil
	}
	result := make([]*models.ProjectMember, 0, len(projectMembers))
	for _, m := range projectMembers {
		cp := *m
		result = append(result, &cp)
	}
	return result, nil
}

func (r *ProjectRepository) NextIssueNumber(_ context.Context, projectID uuid.UUID) (int, error) {
	n, ok := r.counters[projectID]
	if !ok {
		return 0, fmt.Errorf("project %s not found: %w", projectID, models.ErrNotFound)
	}
	r.counters[projectID] = n + 1
	return n, nil
}

func (r *ProjectRepository) NextSprintNumber(_ context.Context, projectID uuid.UUID) (int, error) {
	n, ok := r.sprintCounters[projectID]
	if !ok {
		return 0, fmt.Errorf("project %s not found: %w", projectID, models.ErrNotFound)
	}
	r.sprintCounters[projectID] = n + 1
	return n, nil
}
