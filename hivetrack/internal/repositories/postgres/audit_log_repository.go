package postgres

import (
	"context"
	"fmt"

	"github.com/the127/hivetrack/internal/models"
)

type AuditLogRepository struct {
	db *DbContext
}

func NewAuditLogRepository(db *DbContext) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Insert(_ context.Context, _ *models.AuditLogEntry) error {
	return fmt.Errorf("postgres AuditLogRepository.Insert: not implemented")
}
