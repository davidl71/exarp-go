// tasks_crud_bench_test.go — Benchmarks for task CRUD used by task_workflow (SQLite).
//
// Profile CPU (example):
//
//	CGO_ENABLED=0 go test -run=^$ -bench='Benchmark(Create|Get|Update|Delete|BatchUpdate)Task' -benchmem -count=5 \
//	  -cpuprofile=crud_cpu.prof ./internal/database/
//	go tool pprof -http=:6060 crud_cpu.prof
//
// Profile allocations:
//
//	CGO_ENABLED=0 go test -run=^$ -bench='Benchmark(Create|Get|Update|Delete|BatchUpdate)Task' -benchmem -memprofile=crud_mem.prof ./internal/database/
//	go tool pprof -http=:6061 -sample_index=alloc_space crud_mem.prof
package database

import (
	"context"
	"fmt"
	"strconv"
	"testing"
)

func benchmarkInitTempDB(b *testing.B) {
	b.Helper()
	testDBMu.Lock()
	b.Cleanup(func() {
		Close()
		testDBMu.Unlock()
	})
	if err := Init(b.TempDir()); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkCreateTask(b *testing.B) {
	benchmarkInitTempDB(b)
	ctx := context.Background()
	var serial int64
	const idBase = 9_000_000_000_000_000_000 // T-<digits> only (models.IsValidTaskID).
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		serial++
		task := &Todo2Task{
			ID:              "T-" + strconv.FormatUint(idBase+uint64(serial), 10),
			Content:         "bench task",
			LongDescription: "description",
			Status:          "Todo",
			Priority:        "medium",
			Tags:            []string{"bench", "crud"},
			Dependencies:    nil,
		}
		if err := CreateTask(ctx, task); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetTask(b *testing.B) {
	benchmarkInitTempDB(b)
	ctx := context.Background()
	seed := &Todo2Task{
		ID:              "T-9000000000000000001",
		Content:         "bench",
		LongDescription: "d",
		Status:          "Todo",
		Priority:        "medium",
		Tags:            []string{"a", "b"},
		Dependencies:    []string{},
	}
	if err := CreateTask(ctx, seed); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := GetTask(ctx, "T-9000000000000000001")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUpdateTask(b *testing.B) {
	benchmarkInitTempDB(b)
	ctx := context.Background()
	task := &Todo2Task{
		ID:              "T-9000000000000000002",
		Content:         "v0",
		LongDescription: "d",
		Status:          "Todo",
		Priority:        "medium",
		Tags:            []string{"x", "y"},
		Dependencies:    nil,
	}
	if err := CreateTask(ctx, task); err != nil {
		b.Fatal(err)
	}
	var serial int64
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		serial++
		task.Content = "v" + strconv.FormatInt(serial, 10)
		if err := UpdateTask(ctx, task); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDeleteTask(b *testing.B) {
	benchmarkInitTempDB(b)
	ctx := context.Background()
	var serial int64
	const delBase = 8_000_000_000_000_000_000
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		serial++
		id := "T-" + strconv.FormatUint(delBase+uint64(serial), 10)
		task := &Todo2Task{
			ID: id, Content: "x", LongDescription: "d", Status: "Todo", Priority: "low",
			Tags: []string{}, Dependencies: nil,
		}
		if err := CreateTask(ctx, task); err != nil {
			b.Fatal(err)
		}
		if err := DeleteTask(ctx, id); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBatchUpdateTaskStatus_64(b *testing.B) {
	const n = 64
	benchmarkInitTempDB(b)
	ctx := context.Background()
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		ids[i] = fmt.Sprintf("T-%d", 7_000_000_000_000_000_000+i)
		task := &Todo2Task{
			ID: ids[i], Content: "x", LongDescription: "d", Status: "Todo", Priority: "medium",
			Tags: nil, Dependencies: nil,
		}
		if err := CreateTask(ctx, task); err != nil {
			b.Fatal(err)
		}
	}
	updates := make([]TaskStatusUpdate, n)
	for i := 0; i < n; i++ {
		updates[i] = TaskStatusUpdate{TaskID: ids[i], Status: "In Progress"}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, err := BatchUpdateTaskStatus(ctx, updates)
		if err != nil {
			b.Fatal(err)
		}
	}
}
