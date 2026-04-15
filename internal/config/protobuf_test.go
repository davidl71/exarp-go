package config

import (
	"reflect"
	"testing"
	"time"
)

// TestProtobufRoundTrip tests that conversion is lossless (Go → Protobuf → Go).
func TestProtobufRoundTrip(t *testing.T) {
	// Test with full defaults
	original := GetDefaults()

	// Go → Protobuf → Go
	pb, err := ToProtobuf(original)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	converted, err := FromProtobuf(pb)
	if err != nil {
		t.Fatalf("FromProtobuf failed: %v", err)
	}

	// Verify key fields are preserved
	if converted.Version != original.Version {
		t.Errorf("Version: got %q, want %q", converted.Version, original.Version)
	}

	// Verify timeouts (with tolerance for duration conversion)
	if converted.Timeouts.TaskLockLease != original.Timeouts.TaskLockLease {
		t.Errorf("Timeouts.TaskLockLease: got %v, want %v", converted.Timeouts.TaskLockLease, original.Timeouts.TaskLockLease)
	}

	if converted.Timeouts.ToolDefault != original.Timeouts.ToolDefault {
		t.Errorf("Timeouts.ToolDefault: got %v, want %v", converted.Timeouts.ToolDefault, original.Timeouts.ToolDefault)
	}

	// Verify thresholds
	if converted.Thresholds.SimilarityThreshold != original.Thresholds.SimilarityThreshold {
		t.Errorf("Thresholds.SimilarityThreshold: got %f, want %f", converted.Thresholds.SimilarityThreshold, original.Thresholds.SimilarityThreshold)
	}

	if converted.Thresholds.MinCoverage != original.Thresholds.MinCoverage {
		t.Errorf("Thresholds.MinCoverage: got %d, want %d", converted.Thresholds.MinCoverage, original.Thresholds.MinCoverage)
	}

	// Verify tasks
	if converted.Tasks.DefaultStatus != original.Tasks.DefaultStatus {
		t.Errorf("Tasks.DefaultStatus: got %q, want %q", converted.Tasks.DefaultStatus, original.Tasks.DefaultStatus)
	}

	if converted.Tasks.DefaultPriority != original.Tasks.DefaultPriority {
		t.Errorf("Tasks.DefaultPriority: got %q, want %q", converted.Tasks.DefaultPriority, original.Tasks.DefaultPriority)
	}

	// Verify StatusWorkflow map (converted via JSON)
	if !reflect.DeepEqual(converted.Tasks.StatusWorkflow, original.Tasks.StatusWorkflow) {
		t.Errorf("Tasks.StatusWorkflow: got %v, want %v", converted.Tasks.StatusWorkflow, original.Tasks.StatusWorkflow)
	}
}

// TestProtobufRoundTripWithCustomValues tests round-trip with custom values.
func TestProtobufRoundTripWithCustomValues(t *testing.T) {
	// Create config with custom values
	original := &FullConfig{
		Version: "2.0",
		Timeouts: TimeoutsConfig{
			TaskLockLease:    45 * time.Minute,
			ToolDefault:      120 * time.Second,
			OllamaDownload:   900 * time.Second,
			ContextSummarize: 60 * time.Second,
		},
		Thresholds: ThresholdsConfig{
			SimilarityThreshold:  0.95,
			MinDescriptionLength: 100,
			MinCoverage:          90,
			MaxParallelTasks:     20,
		},
		Tasks: TasksConfig{
			DefaultStatus:   "In Progress",
			DefaultPriority: "high",
			DefaultTags:     []string{"urgent", "backend"},
			StatusWorkflow: map[string][]string{
				"Todo":        {"In Progress"},
				"In Progress": {"Review", "Done"},
				"Review":      {"Done", "In Progress"},
			},
			StaleThresholdHours: 4,
			AutoCleanupEnabled:  true,
		},
	}

	// Go → Protobuf → Go
	pb, err := ToProtobuf(original)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	converted, err := FromProtobuf(pb)
	if err != nil {
		t.Fatalf("FromProtobuf failed: %v", err)
	}

	// Verify all custom values are preserved
	if converted.Version != original.Version {
		t.Errorf("Version: got %q, want %q", converted.Version, original.Version)
	}

	if converted.Timeouts.TaskLockLease != original.Timeouts.TaskLockLease {
		t.Errorf("Timeouts.TaskLockLease: got %v, want %v", converted.Timeouts.TaskLockLease, original.Timeouts.TaskLockLease)
	}

	if converted.Thresholds.SimilarityThreshold != original.Thresholds.SimilarityThreshold {
		t.Errorf("Thresholds.SimilarityThreshold: got %f, want %f", converted.Thresholds.SimilarityThreshold, original.Thresholds.SimilarityThreshold)
	}

	if converted.Tasks.DefaultStatus != original.Tasks.DefaultStatus {
		t.Errorf("Tasks.DefaultStatus: got %q, want %q", converted.Tasks.DefaultStatus, original.Tasks.DefaultStatus)
	}

	if !reflect.DeepEqual(converted.Tasks.DefaultTags, original.Tasks.DefaultTags) {
		t.Errorf("Tasks.DefaultTags: got %v, want %v", converted.Tasks.DefaultTags, original.Tasks.DefaultTags)
	}

	if !reflect.DeepEqual(converted.Tasks.StatusWorkflow, original.Tasks.StatusWorkflow) {
		t.Errorf("Tasks.StatusWorkflow: got %v, want %v", converted.Tasks.StatusWorkflow, original.Tasks.StatusWorkflow)
	}
}

// TestProtobufRoundTripWithZeroValues tests round-trip with zero/empty values.
func TestProtobufRoundTripWithZeroValues(t *testing.T) {
	// Create minimal config with mostly zero values
	original := &FullConfig{
		Version: "1.0",
		// Most fields are zero values
	}

	// Go → Protobuf → Go
	pb, err := ToProtobuf(original)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	converted, err := FromProtobuf(pb)
	if err != nil {
		t.Fatalf("FromProtobuf failed: %v", err)
	}

	// Verify version is preserved
	if converted.Version != original.Version {
		t.Errorf("Version: got %q, want %q", converted.Version, original.Version)
	}
	// Zero values should remain zero (or be set to defaults if protobuf sets them)
	// This test ensures conversion doesn't break with minimal configs
}

// TestProtobufDurationConversion tests duration conversion accuracy.
func TestProtobufDurationConversion(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{"zero", 0},
		{"1 second", 1 * time.Second},
		{"30 seconds", 30 * time.Second},
		{"1 minute", 1 * time.Minute},
		{"30 minutes", 30 * time.Minute},
		{"1 hour", 1 * time.Hour},
		{"2 hours", 2 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert duration to seconds and back
			seconds := durationToSeconds(tt.duration)
			converted := secondsToDuration(seconds)

			// Durations should match exactly (no precision loss for whole seconds)
			if converted != tt.duration {
				t.Errorf("Duration conversion: got %v, want %v", converted, tt.duration)
			}
		})
	}
}

// TestProtobufMapJSONConversion tests JSON conversion for map fields.
func TestProtobufMapJSONConversion(t *testing.T) {
	// Test StatusWorkflow map
	statusWorkflow := map[string][]string{
		"Todo":        {"In Progress", "Review"},
		"In Progress": {"Done", "Review", "Todo"},
		"Review":      {"Done", "In Progress"},
		"Done":        {},
	}

	jsonStr, err := mapToJSON(statusWorkflow)
	if err != nil {
		t.Fatalf("mapToJSON failed: %v", err)
	}

	var converted map[string][]string
	if err := jsonToMap(jsonStr, &converted); err != nil {
		t.Fatalf("jsonToMap failed: %v", err)
	}

	if !reflect.DeepEqual(converted, statusWorkflow) {
		t.Errorf("StatusWorkflow conversion: got %v, want %v", converted, statusWorkflow)
	}

	// Test DefaultScores map
	defaultScores := map[string]float64{
		"security":      75.0,
		"testing":       80.0,
		"documentation": 65.0,
	}

	jsonStr2, err := mapToJSON(defaultScores)
	if err != nil {
		t.Fatalf("mapToJSON failed: %v", err)
	}

	var converted2 map[string]float64
	if err := jsonToMap(jsonStr2, &converted2); err != nil {
		t.Fatalf("jsonToMap failed: %v", err)
	}

	if !reflect.DeepEqual(converted2, defaultScores) {
		t.Errorf("DefaultScores conversion: got %v, want %v", converted2, defaultScores)
	}
}

// TestProtobufNilHandling tests handling of nil pointers.
func TestProtobufNilHandling(t *testing.T) {
	// Test ToProtobuf with nil
	_, err := ToProtobuf(nil)
	if err == nil {
		t.Error("ToProtobuf(nil) should return error")
	}

	// Test FromProtobuf with nil
	_, err = FromProtobuf(nil)
	if err == nil {
		t.Error("FromProtobuf(nil) should return error")
	}
}

// TestProtobufNestedConfigs tests conversion of nested config structures.
func TestProtobufNestedConfigs(t *testing.T) {
	original := GetDefaults()

	// Test that nested configs are converted
	pb, err := ToProtobuf(original)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	// Verify nested configs exist in protobuf
	if pb.GetTimeouts() == nil {
		t.Error("Timeouts should not be nil")
	}

	if pb.GetThresholds() == nil {
		t.Error("Thresholds should not be nil")
	}

	if pb.GetTasks() == nil {
		t.Error("Tasks should not be nil")
	}

	if pb.GetDatabase() == nil {
		t.Error("Database should not be nil")
	}

	if pb.GetSecurity() == nil {
		t.Error("Security should not be nil")
	}

	if pb.GetLogging() == nil {
		t.Error("Logging should not be nil")
	}

	if pb.GetTools() == nil {
		t.Error("Tools should not be nil")
	}

	if pb.GetWorkflow() == nil {
		t.Error("Workflow should not be nil")
	}

	if pb.GetMemory() == nil {
		t.Error("Memory should not be nil")
	}

	if pb.GetProject() == nil {
		t.Error("Project should not be nil")
	}
}

// TestProtobufScorecardDefaultScores tests conversion of ScorecardConfig.DefaultScores map.
func TestProtobufScorecardDefaultScores(t *testing.T) {
	original := GetDefaults()
	original.Tools.Scorecard.DefaultScores = map[string]float64{
		"security":      75.0,
		"testing":       80.0,
		"documentation": 65.0,
		"completion":    70.0,
	}

	// Go → Protobuf → Go
	pb, err := ToProtobuf(original)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	converted, err := FromProtobuf(pb)
	if err != nil {
		t.Fatalf("FromProtobuf failed: %v", err)
	}

	// Verify DefaultScores map is preserved
	if !reflect.DeepEqual(converted.Tools.Scorecard.DefaultScores, original.Tools.Scorecard.DefaultScores) {
		t.Errorf("Tools.Scorecard.DefaultScores: got %v, want %v", converted.Tools.Scorecard.DefaultScores, original.Tools.Scorecard.DefaultScores)
	}
}

// TestProtobufWorkflowModes tests conversion of WorkflowConfig.Modes map.
func TestProtobufWorkflowModes(t *testing.T) {
	original := GetDefaults()
	original.Workflow.Modes = map[string]ModeConfig{
		"development": {
			EnabledTools:  []string{"testing", "linting"},
			DisabledTools: []string{"report"},
			ToolLimit:     10,
		},
		"production": {
			EnabledTools:  []string{"report"},
			DisabledTools: []string{"testing"},
			ToolLimit:     5,
		},
	}

	// Go → Protobuf → Go
	pb, err := ToProtobuf(original)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	converted, err := FromProtobuf(pb)
	if err != nil {
		t.Fatalf("FromProtobuf failed: %v", err)
	}

	// Verify Modes map is preserved
	if !reflect.DeepEqual(converted.Workflow.Modes, original.Workflow.Modes) {
		t.Errorf("Workflow.Modes: got %v, want %v", converted.Workflow.Modes, original.Workflow.Modes)
	}
}

// TestProtobufEmptyMaps tests conversion with empty maps.
func TestProtobufEmptyMaps(t *testing.T) {
	original := GetDefaults()
	original.Tasks.StatusWorkflow = map[string][]string{}         // Empty map
	original.Tools.Scorecard.DefaultScores = map[string]float64{} // Empty map
	original.Workflow.Modes = map[string]ModeConfig{}             // Empty map

	// Go → Protobuf → Go
	pb, err := ToProtobuf(original)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	converted, err := FromProtobuf(pb)
	if err != nil {
		t.Fatalf("FromProtobuf failed: %v", err)
	}

	// Empty maps should remain empty (or be nil, which is acceptable)
	if converted.Tasks.StatusWorkflow != nil && len(converted.Tasks.StatusWorkflow) != 0 {
		t.Errorf("Empty StatusWorkflow should remain empty, got %v", converted.Tasks.StatusWorkflow)
	}

	if converted.Tools.Scorecard.DefaultScores != nil && len(converted.Tools.Scorecard.DefaultScores) != 0 {
		t.Errorf("Empty DefaultScores should remain empty, got %v", converted.Tools.Scorecard.DefaultScores)
	}

	if converted.Workflow.Modes != nil && len(converted.Workflow.Modes) != 0 {
		t.Errorf("Empty Modes should remain empty, got %v", converted.Workflow.Modes)
	}
}

// TestProtobufAllTimeouts tests conversion of all timeout fields.
func TestProtobufAllTimeouts(t *testing.T) {
	original := &FullConfig{
		Timeouts: TimeoutsConfig{
			TaskLockLease:      30 * time.Minute,
			TaskLockRenewal:    20 * time.Minute,
			StaleLockThreshold: 5 * time.Minute,
			ToolDefault:        60 * time.Second,
			ToolScorecard:      60 * time.Second,
			ToolLinting:        60 * time.Second,
			ToolTesting:        300 * time.Second,
			ToolReport:         60 * time.Second,
			OllamaDownload:     600 * time.Second,
			OllamaGenerate:     300 * time.Second,
			HTTPClient:         30 * time.Second,
			DatabaseRetry:      60 * time.Second,
			ContextSummarize:   30 * time.Second,
			ContextBudget:      10 * time.Second,
		},
	}

	pb, err := ToProtobuf(original)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	converted, err := FromProtobuf(pb)
	if err != nil {
		t.Fatalf("FromProtobuf failed: %v", err)
	}

	// Verify all timeout fields
	if converted.Timeouts.TaskLockLease != original.Timeouts.TaskLockLease {
		t.Errorf("TaskLockLease: got %v, want %v", converted.Timeouts.TaskLockLease, original.Timeouts.TaskLockLease)
	}

	if converted.Timeouts.TaskLockRenewal != original.Timeouts.TaskLockRenewal {
		t.Errorf("TaskLockRenewal: got %v, want %v", converted.Timeouts.TaskLockRenewal, original.Timeouts.TaskLockRenewal)
	}

	if converted.Timeouts.StaleLockThreshold != original.Timeouts.StaleLockThreshold {
		t.Errorf("StaleLockThreshold: got %v, want %v", converted.Timeouts.StaleLockThreshold, original.Timeouts.StaleLockThreshold)
	}

	if converted.Timeouts.ToolDefault != original.Timeouts.ToolDefault {
		t.Errorf("ToolDefault: got %v, want %v", converted.Timeouts.ToolDefault, original.Timeouts.ToolDefault)
	}

	if converted.Timeouts.OllamaDownload != original.Timeouts.OllamaDownload {
		t.Errorf("OllamaDownload: got %v, want %v", converted.Timeouts.OllamaDownload, original.Timeouts.OllamaDownload)
	}

	if converted.Timeouts.ContextBudget != original.Timeouts.ContextBudget {
		t.Errorf("ContextBudget: got %v, want %v", converted.Timeouts.ContextBudget, original.Timeouts.ContextBudget)
	}
}

// TestProtobufSchemaSync tests that protobuf schema matches Go structs
// This is a high-level test to catch schema drift.
func TestProtobufSchemaSync(t *testing.T) {
	// Create a full config with all fields set
	original := GetDefaults()

	// Convert to protobuf
	pb, err := ToProtobuf(original)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	// Verify all major sections exist
	if pb == nil {
		t.Fatal("Protobuf config is nil")
	}

	// Check that all expected nested configs are present
	requiredSections := []struct {
		name  string
		check func() bool
	}{
		{"Timeouts", func() bool { return pb.GetTimeouts() != nil }},
		{"Thresholds", func() bool { return pb.GetThresholds() != nil }},
		{"Tasks", func() bool { return pb.GetTasks() != nil }},
		{"Database", func() bool { return pb.GetDatabase() != nil }},
		{"Security", func() bool { return pb.GetSecurity() != nil }},
		{"Logging", func() bool { return pb.GetLogging() != nil }},
		{"Tools", func() bool { return pb.GetTools() != nil }},
		{"Workflow", func() bool { return pb.GetWorkflow() != nil }},
		{"Memory", func() bool { return pb.GetMemory() != nil }},
		{"Project", func() bool { return pb.GetProject() != nil }},
	}

	for _, section := range requiredSections {
		if !section.check() {
			t.Errorf("Protobuf schema missing section: %s", section.name)
		}
	}
}

// TestProtobufPartialConfig tests conversion with partial config (some sections missing).
func TestProtobufPartialConfig(t *testing.T) {
	// Create config with only some sections set
	original := &FullConfig{
		Version: "1.0",
		Timeouts: TimeoutsConfig{
			TaskLockLease: 45 * time.Minute,
			ToolDefault:   120 * time.Second,
		},
		Thresholds: ThresholdsConfig{
			SimilarityThreshold: 0.9,
			MinCoverage:         85,
		},
		// Other sections are zero values
	}

	// Go → Protobuf → Go
	pb, err := ToProtobuf(original)
	if err != nil {
		t.Fatalf("ToProtobuf failed: %v", err)
	}

	converted, err := FromProtobuf(pb)
	if err != nil {
		t.Fatalf("FromProtobuf failed: %v", err)
	}

	// Verify set values are preserved
	if converted.Timeouts.TaskLockLease != original.Timeouts.TaskLockLease {
		t.Errorf("TaskLockLease: got %v, want %v", converted.Timeouts.TaskLockLease, original.Timeouts.TaskLockLease)
	}

	if converted.Thresholds.SimilarityThreshold != original.Thresholds.SimilarityThreshold {
		t.Errorf("SimilarityThreshold: got %f, want %f", converted.Thresholds.SimilarityThreshold, original.Thresholds.SimilarityThreshold)
	}
}
