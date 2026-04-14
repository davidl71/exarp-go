package tools

import (
	"github.com/davidl71/exarp-go/internal/models"
)

// ApprovalRequest is a portable payload for external human review (form_id + field_data).
// Clients may map it to email, ticketing, or another MCP tool; exarp-go does not call third-party review services.
type ApprovalRequest struct {
	FormID    string                 `json:"form_id"`
	FieldData map[string]interface{} `json:"field_data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// BuildApprovalRequestFromTask builds an ApprovalRequest from a Todo2 task.
// formID is an optional template or form identifier for the consumer; if empty, the client may substitute a default.
func BuildApprovalRequestFromTask(task *models.Todo2Task, formID string) ApprovalRequest {
	fieldData := map[string]interface{}{
		"task_id":     task.ID,
		"title":       task.Content,
		"description": task.LongDescription,
		"status":      task.Status,
		"priority":    task.Priority,
	}
	if len(task.Tags) > 0 {
		// Clone to avoid leaking references to internal slices.
		fieldData["tags"] = append([]string(nil), task.Tags...)
	}

	metadata := map[string]interface{}{
		"source":  "exarp-go",
		"task_id": task.ID,
	}

	return ApprovalRequest{
		FormID:    formID,
		FieldData: fieldData,
		Metadata:  metadata,
	}
}
