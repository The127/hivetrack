package inmemory

import (
	"context"

	"github.com/google/uuid"

	"github.com/the127/hivetrack/internal/change"
	"github.com/the127/hivetrack/internal/models"
	"github.com/the127/hivetrack/internal/repositories"
)

type MilestoneRepository struct {
	tracker *change.Tracker
	byID    map[uuid.UUID]*models.Milestone
	issues  *IssueRepository
}

func NewMilestoneRepository(tracker *change.Tracker, issues *IssueRepository) *MilestoneRepository {
	return &MilestoneRepository{
		tracker: tracker,
		byID:    make(map[uuid.UUID]*models.Milestone),
		issues:  issues,
	}
}

func (r *MilestoneRepository) Insert(milestone *models.Milestone) {
	r.tracker.Add(change.NewEntry(0, milestone, change.Added))
}

func (r *MilestoneRepository) Update(milestone *models.Milestone) {
	r.tracker.Add(change.NewEntry(0, milestone, change.Updated))
}

func (r *MilestoneRepository) Delete(milestone *models.Milestone) {
	r.tracker.Add(change.NewEntry(0, milestone, change.Deleted))
}

func (r *MilestoneRepository) GetByID(_ context.Context, id uuid.UUID) (*models.Milestone, error) {
	m, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return m, nil
}

func (r *MilestoneRepository) List(_ context.Context, projectID uuid.UUID) ([]*models.Milestone, error) {
	var result []*models.Milestone
	for _, m := range r.byID {
		if m.GetProjectID() == projectID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (r *MilestoneRepository) CountByMilestone(_ context.Context, projectID uuid.UUID) (map[uuid.UUID]repositories.MilestoneProgress, error) {
	result := make(map[uuid.UUID]repositories.MilestoneProgress)
	for _, issue := range r.issues.byID {
		if issue.GetProjectID() != projectID {
			continue
		}
		mid := issue.GetMilestoneID()
		if mid == nil {
			continue
		}
		p := result[*mid]
		p.IssueCount++
		if issue.IsTerminal() {
			p.ClosedIssueCount++
		}
		result[*mid] = p
	}
	return result, nil
}
