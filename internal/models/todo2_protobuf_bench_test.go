package models

import (
	"encoding/json"
	"testing"
)

func newMinimalTask() *Todo2Task {
	return &Todo2Task{
		ID:      "T-1",
		Content: "Benchmark task",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{"nested": "value"},
		},
	}
}

func newRealisticTask() *Todo2Task {
	return &Todo2Task{
		ID:              "T-1771428476221782000",
		Content:         "Scorecard: Add 5-minute result cache for improved performance",
		LongDescription: "Implement a time-based cache (5-minute TTL) for scorecard results to avoid recomputing metrics on every request. Use sync.Map or a simple struct with mutex. Invalidate on task status changes.",
		Status:          "In Progress",
		Priority:        "high",
		Tags:            []string{"#cache", "#performance", "#scorecard", "#refactor"},
		Dependencies:    []string{"T-1771426912215029000", "T-1771426910494633000"},
		Completed:       false,
		ProjectID:       "exarp-go",
		AssignedTo:      "backend-agent",
		Host:            "Davids-Mac-mini",
		Agent:           "backend-agent-Davids-Mac-mini-12345",
		Metadata: map[string]interface{}{
			"preferred_backend": "fm",
			"estimated_hours":   2.5,
			"wave":              1,
			"sprint":            "2026-W08",
			"created_by":        "cursor-agent",
			"complexity":        "medium",
			"files_touched":     []string{"internal/tools/report.go", "internal/tools/scorecard.go"},
			"acceptance_criteria": map[string]interface{}{
				"cache_ttl":     "5 minutes",
				"invalidation":  "on task status change",
				"thread_safety": true,
			},
		},
	}
}

// --- Protobuf binary benchmarks ---

func BenchmarkProtobufSerialization(b *testing.B) {
	task := newMinimalTask()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := SerializeTaskToProtobuf(task); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONSerialization(b *testing.B) {
	task := newMinimalTask()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(task); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProtobufDeserialization(b *testing.B) {
	task := newMinimalTask()
	data, err := SerializeTaskToProtobuf(task)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DeserializeTaskFromProtobuf(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONDeserialization(b *testing.B) {
	task := newMinimalTask()
	data, err := json.Marshal(task)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var t Todo2Task
		if err := json.Unmarshal(data, &t); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProtobufRoundTrip(b *testing.B) {
	task := newMinimalTask()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := SerializeTaskToProtobuf(task)
		if err != nil {
			b.Fatal(err)
		}
		if _, err = DeserializeTaskFromProtobuf(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONRoundTrip(b *testing.B) {
	task := newMinimalTask()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(task)
		if err != nil {
			b.Fatal(err)
		}
		var t Todo2Task
		if err = json.Unmarshal(data, &t); err != nil {
			b.Fatal(err)
		}
	}
}

// --- Protobuf JSON (protojson) benchmarks ---

func BenchmarkProtojsonSerialization(b *testing.B) {
	task := newMinimalTask()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := SerializeTaskToProtobufJSON(task); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProtojsonDeserialization(b *testing.B) {
	task := newMinimalTask()
	data, err := SerializeTaskToProtobufJSON(task)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := DeserializeTaskFromProtobufJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProtojsonRoundTrip(b *testing.B) {
	task := newMinimalTask()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := SerializeTaskToProtobufJSON(task)
		if err != nil {
			b.Fatal(err)
		}
		if _, err = DeserializeTaskFromProtobufJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}

// --- Realistic task benchmarks (all three formats) ---

func BenchmarkRealisticProtobufRoundTrip(b *testing.B) {
	task := newRealisticTask()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := SerializeTaskToProtobuf(task)
		if err != nil {
			b.Fatal(err)
		}
		if _, err = DeserializeTaskFromProtobuf(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRealisticJSONRoundTrip(b *testing.B) {
	task := newRealisticTask()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(task)
		if err != nil {
			b.Fatal(err)
		}
		var t Todo2Task
		if err = json.Unmarshal(data, &t); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRealisticProtojsonRoundTrip(b *testing.B) {
	task := newRealisticTask()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := SerializeTaskToProtobufJSON(task)
		if err != nil {
			b.Fatal(err)
		}
		if _, err = DeserializeTaskFromProtobufJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}

// --- Payload size comparison ---

func BenchmarkPayloadSizeComparison(b *testing.B) {
	task := newRealisticTask()
	var protobufSize, jsonSize, protojsonSize int

	for i := 0; i < b.N; i++ {
		pb, err := SerializeTaskToProtobuf(task)
		if err != nil {
			b.Fatal(err)
		}
		protobufSize = len(pb)

		js, err := json.Marshal(task)
		if err != nil {
			b.Fatal(err)
		}
		jsonSize = len(js)

		pj, err := SerializeTaskToProtobufJSON(task)
		if err != nil {
			b.Fatal(err)
		}
		protojsonSize = len(pj)
	}

	b.Logf("Protobuf binary: %d bytes", protobufSize)
	b.Logf("Standard JSON:   %d bytes", jsonSize)
	b.Logf("Protobuf JSON:   %d bytes", protojsonSize)
	b.Logf("Binary vs JSON size reduction: %.1f%%", float64(jsonSize-protobufSize)/float64(jsonSize)*100)
}
