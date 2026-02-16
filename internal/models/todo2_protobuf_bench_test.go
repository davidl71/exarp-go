package models

import (
	"encoding/json"
	"testing"
)

// BenchmarkProtobufSerialization benchmarks protobuf serialization.
func BenchmarkProtobufSerialization(b *testing.B) {
	task := &Todo2Task{
		ID:      "T-1",
		Content: "Benchmark task",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{"nested": "value"},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := SerializeTaskToProtobuf(task)
		if err != nil {
			b.Fatalf("SerializeTaskToProtobuf() error = %v", err)
		}
	}
}

// BenchmarkJSONSerialization benchmarks JSON serialization.
func BenchmarkJSONSerialization(b *testing.B) {
	task := &Todo2Task{
		ID:      "T-1",
		Content: "Benchmark task",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{"nested": "value"},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(task)
		if err != nil {
			b.Fatalf("json.Marshal() error = %v", err)
		}
	}
}

// BenchmarkProtobufDeserialization benchmarks protobuf deserialization.
func BenchmarkProtobufDeserialization(b *testing.B) {
	task := &Todo2Task{
		ID:      "T-1",
		Content: "Benchmark task",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{"nested": "value"},
		},
	}

	data, err := SerializeTaskToProtobuf(task)
	if err != nil {
		b.Fatalf("SerializeTaskToProtobuf() error = %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := DeserializeTaskFromProtobuf(data)
		if err != nil {
			b.Fatalf("DeserializeTaskFromProtobuf() error = %v", err)
		}
	}
}

// BenchmarkJSONDeserialization benchmarks JSON deserialization.
func BenchmarkJSONDeserialization(b *testing.B) {
	task := &Todo2Task{
		ID:      "T-1",
		Content: "Benchmark task",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{"nested": "value"},
		},
	}

	data, err := json.Marshal(task)
	if err != nil {
		b.Fatalf("json.Marshal() error = %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var task Todo2Task

		err := json.Unmarshal(data, &task)
		if err != nil {
			b.Fatalf("json.Unmarshal() error = %v", err)
		}
	}
}

// BenchmarkProtobufRoundTrip benchmarks full round-trip (serialize + deserialize).
func BenchmarkProtobufRoundTrip(b *testing.B) {
	task := &Todo2Task{
		ID:      "T-1",
		Content: "Benchmark task",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{"nested": "value"},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := SerializeTaskToProtobuf(task)
		if err != nil {
			b.Fatalf("SerializeTaskToProtobuf() error = %v", err)
		}

		_, err = DeserializeTaskFromProtobuf(data)
		if err != nil {
			b.Fatalf("DeserializeTaskFromProtobuf() error = %v", err)
		}
	}
}

// BenchmarkJSONRoundTrip benchmarks full round-trip (serialize + deserialize).
func BenchmarkJSONRoundTrip(b *testing.B) {
	task := &Todo2Task{
		ID:      "T-1",
		Content: "Benchmark task",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{"nested": "value"},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := json.Marshal(task)
		if err != nil {
			b.Fatalf("json.Marshal() error = %v", err)
		}

		var task Todo2Task

		err = json.Unmarshal(data, &task)
		if err != nil {
			b.Fatalf("json.Unmarshal() error = %v", err)
		}
	}
}

// BenchmarkProtobufSizeComparison compares payload sizes.
func BenchmarkProtobufSizeComparison(b *testing.B) {
	task := &Todo2Task{
		ID:      "T-1",
		Content: "Benchmark task with extensive metadata",
		Status:  "Todo",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": map[string]interface{}{"nested": "value"},
			"key4": []string{"item1", "item2", "item3"},
			"key5": 42,
			"key6": true,
		},
	}

	var protobufSize, jsonSize int

	for i := 0; i < b.N; i++ {
		protobufData, err := SerializeTaskToProtobuf(task)
		if err != nil {
			b.Fatalf("SerializeTaskToProtobuf() error = %v", err)
		}

		protobufSize = len(protobufData)

		jsonData, err := json.Marshal(task)
		if err != nil {
			b.Fatalf("json.Marshal() error = %v", err)
		}

		jsonSize = len(jsonData)
	}

	b.Logf("Protobuf size: %d bytes", protobufSize)
	b.Logf("JSON size: %d bytes", jsonSize)
	b.Logf("Size reduction: %.1f%%", float64(jsonSize-protobufSize)/float64(jsonSize)*100)
}
