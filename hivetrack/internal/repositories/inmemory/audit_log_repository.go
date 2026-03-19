package inmemory

import (
	"context"

	"github.com/the127/hivetrack/internal/models"
)

type AuditLogRepository struct {
	entries []*models.AuditLogEntry
}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{}
}

func (r *AuditLogRepository) Insert(_ context.Context, entry *models.AuditLogEntry) error {
	r.entries = append(r.entries, entry)
	return nil
}

// Entries returns all recorded audit log entries (for test assertions).
func (r *AuditLogRepository) Entries() []*models.AuditLogEntry {
	return r.entries
}
