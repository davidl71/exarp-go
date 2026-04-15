package tools

import (
	"github.com/davidl71/exarp-go/internal/models"
)

// ApprovalRequest is the payload for gotoHuman request-human-review-with-form.
// The client (e.g. Cursor) invokes the gotoHuman MCP tool with these fields.
type ApprovalRequest struct {
	FormID    string                 `json:"form_id"`
	FieldData map[string]interface{} `json:"field_data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// BuildApprovalRequestFromTask builds an ApprovalRequest from a Todo2 task for
// sending to gotoHuman. formID should be a form ID from gotoHuman list-forms;
// if empty, the client must substitute a default form ID.
func BuildApprovalRequestFromTask(task *models.Todo2Task, formID string) ApprovalRequest {
	fieldData := map[string]interface{}{
		"task_id":     task.ID,
		"title":       task.Content,
		"description": task.LongDescription,
		"status":      task.Status,
		"priority":    task.Priority,
	}
	if len(task.Tags) > 0 {
		fieldData["tags"] = task.Tags
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
