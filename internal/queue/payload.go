package queue

import (
	"encoding/json"
)

// TaskExecutePayload is the payload for TypeTaskExecute jobs.
// Worker will claim task_id in Todo2, run task_execute, then release.
type TaskExecutePayload struct {
	TaskID      string `json:"task_id"`
	ProjectRoot string `json:"project_root"`
}

// Marshal serializes the payload to JSON for Asynq.
func (p *TaskExecutePayload) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

// UnmarshalTaskExecutePayload deserializes JSON into TaskExecutePayload.
func UnmarshalTaskExecutePayload(data []byte) (*TaskExecutePayload, error) {
	var p TaskExecutePayload
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
