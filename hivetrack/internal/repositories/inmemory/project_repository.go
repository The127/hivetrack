package inmemory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type ProjectRepository struct {
	byID     map[uuid.UUID]*models.Project
	bySlug   map[string]*models.Project
	members  map[uuid.UUID]map[uuid.UUID]*models.ProjectMember // projectID -> userID -> member
	counters map[uuid.UUID]int
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		byID:     make(map[uuid.UUID]*models.Project),
		bySlug:   make(map[string]*models.Project),
		members:  make(map[uuid.UUID]map[uuid.UUID]*models.ProjectMember),
		counters: make(map[uuid.UUID]int),
	}
}

func (r *ProjectRepository) Insert(_ context.Context, project *models.Project) error {
	if _, exists := r.bySlug[project.Slug]; exists {
		return fmt.Errorf("project with slug %q already exists: %w", project.Slug, models.ErrConflict)
	}
	cp := *project
	r.byID[project.ID] = &cp
	r.bySlug[project.Slug] = &cp
	r.counters[project.ID] = 1
	return nil
}

func (r *ProjectRepository) Update(_ context.Context, project *models.Project) error {
	existing, ok := r.byID[project.ID]
	if !ok {
		return fmt.Errorf("project %s not found: %w", project.ID, models.ErrNotFound)
	}
	// Remove old slug index if slug changed
	if existing.Slug != project.Slug {
		delete(r.bySlug, existing.Slug)
	}
	cp := *project
	r.byID[project.ID] = &cp
	r.bySlug[project.Slug] = &cp
	return nil
}

func (r *ProjectRepository) Delete(_ context.Context, id uuid.UUID) error {
	project, ok := r.byID[id]
	if !ok {
		return fmt.Errorf("project %s not found: %w", id, models.ErrNotFound)
	}
	delete(r.bySlug, project.Slug)
	delete(r.byID, id)
	delete(r.members, id)
	delete(r.counters, id)
	return nil
}

func (r *ProjectRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Project, error) {
	p, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *ProjectRepository) GetBySlug(_ context.Context, slug string) (*models.Project, error) {
	p, ok := r.bySlug[slug]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *ProjectRepository) List(_ context.Context, filter *repositories.ProjectFilter) ([]*models.Project, error) {
	var result []*models.Project
	for _, p := range r.byID {
		if filter.MemberUserID != nil && !filter.IsAdmin {
			projectMembers, ok := r.members[p.ID]
			if !ok {
				continue
			}
			if _, isMember := projectMembers[*filter.MemberUserID]; !isMember {
				continue
			}
		}
		cp := *p
		result = append(result, &cp)
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
