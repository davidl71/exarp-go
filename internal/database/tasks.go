// tasks.go — Task types, metadata serialization, and ID helpers.
// CRUD in tasks_crud.go; list/query in tasks_list.go; FixTaskDates/GetDependencies/etc. in tasks_misc.go.
//
// Package database provides SQLite storage for tasks, comments, activities, and agent locks.
package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/davidl71/exarp-go/internal/logging"
	"github.com/davidl71/exarp-go/internal/models"
)

// ErrVersionMismatch is returned when an update fails because the task was modified by another agent.
var ErrVersionMismatch = errors.New("task version mismatch")

// Todo2Task is an alias for models.Todo2Task (for convenience).
type Todo2Task = models.Todo2Task

// TaskFilters represents filters for querying tasks.
type TaskFilters struct {
	Status     *string
	Priority   *string
	Tag        *string
	ProjectID  *string
	AssignedTo *string
	Host       *string
	Agent      *string
}

// SanitizeTaskMetadata parses JSON metadata; on failure returns a map with "raw" key and logs.
// Callers never see invalid character parse errors—invalid JSON is coerced to {"raw": "..."}.
// Use when loading tasks from DB or JSON so code that expects metadata as a map never fails.
func SanitizeTaskMetadata(s string) map[string]interface{} {
	if s == "" {
		return nil
	}

	var out map[string]interface{}
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		logging.Warn("database: invalid task metadata JSON, coercing to raw: %v", err)
		return map[string]interface{}{"raw": s}
	}

	return out
}

// unmarshalTaskMetadata is an alias for SanitizeTaskMetadata (internal use).
func unmarshalTaskMetadata(s string) map[string]interface{} {
	return SanitizeTaskMetadata(s)
}

// SerializeTaskMetadata serializes task metadata for DB storage: JSON string, optional protobuf bytes, and format.
// Returns (metadataJSON, metadataProtobuf, metadataFormat, nil) or error if JSON marshal fails.
func SerializeTaskMetadata(task *Todo2Task) (metadataJSON string, metadataProtobuf []byte, metadataFormat string, err error) {
	metadataFormat = "protobuf"

	protobufData, err := models.SerializeTaskToProtobuf(task)
	if err != nil {
		metadataFormat = "json"
	} else {
		metadataProtobuf = protobufData
	}

	if len(task.Metadata) > 0 {
		sanitized := SanitizeMetadataForWrite(task.Metadata)

		metadataBytes, jsonErr := json.Marshal(sanitized)
		if jsonErr != nil {
			return "", nil, "", fmt.Errorf("failed to marshal metadata: %w", jsonErr)
		}

		metadataJSON = string(metadataBytes)
	}

	return metadataJSON, metadataProtobuf, metadataFormat, nil
}

// DeserializeTaskMetadata deserializes metadata from DB: prefers protobuf when format is "protobuf", else JSON.
// metadataJSON and metadataFormat may be empty (e.g. from old schema); invalid JSON is coerced via SanitizeTaskMetadata.
func DeserializeTaskMetadata(metadataJSON string, metadataProtobuf []byte, metadataFormat string) map[string]interface{} {
	if metadataFormat == "protobuf" && len(metadataProtobuf) > 0 {
		deserialized, err := models.DeserializeTaskFromProtobuf(metadataProtobuf)
		if err == nil {
			return deserialized.Metadata
		}
	}

	if metadataJSON != "" {
		return SanitizeTaskMetadata(metadataJSON)
	}

	return nil
}

// SanitizeMetadataForWrite returns a copy of metadata with all values coerced to types
// that encoding/json can marshal (string, float64, int, int64, bool, nil, []interface{},
// map[string]interface{}). Use when writing DB or state JSON so non-scalar values never break marshaling.
func SanitizeMetadataForWrite(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return nil
	}

	out := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		out[k] = jsonSafeValue(v)
	}

	return out
}

func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}

	return sql.NullString{String: s, Valid: true}
}

func jsonSafeValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch x := v.(type) {
	case string:
		return x
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case bool:
		return x
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, e := range x {
			out[i] = jsonSafeValue(e)
		}

		return out
	case []string:
		out := make([]interface{}, len(x))
		for i, s := range x {
			out[i] = s
		}
		return out
	case map[string]interface{}:
		out := make(map[string]interface{}, len(x))
		for k2, val := range x {
			out[k2] = jsonSafeValue(val)
		}

		return out
	default:
		return fmt.Sprint(x)
	}
}

// IsValidTaskID delegates to models.IsValidTaskID.
func IsValidTaskID(id string) bool {
	return models.IsValidTaskID(id)
}

// GenerateTaskID delegates to models.GenerateTaskID.
func GenerateTaskID() string {
	return models.GenerateTaskID()
}
