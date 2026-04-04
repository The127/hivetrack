package commands

import "github.com/the127/hivetrack/internal/models"

// HoldUpdate groups the on-hold fields that always travel together
// when updating hold state on an issue (single or batch).
type HoldUpdate struct {
	OnHold     *bool              `json:"on_hold"`
	HoldReason *models.HoldReason `json:"hold_reason"`
	HoldNote   *string            `json:"hold_note"`
}
