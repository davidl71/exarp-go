package tools

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// generateTestData creates test data arrays of various sizes
func generateTestData(size int) []float64 {
	data := make([]float64, size)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < size; i++ {
		// Generate random float between 0 and 1000
		data[i] = rand.Float64() * 1000.0
	}

	return data
}

// generateSortedTestData creates sorted test data
func generateSortedTestData(size int) []float64 {
	data := generateTestData(size)
	// Note: We'll sort in the benchmark to measure sorting overhead
	return data
}

// BenchmarkMean benchmarks Mean function with various data sizes
func BenchmarkMean(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Small_100", 100},
		{"Medium_1000", 1000},
		{"Large_10000", 10000},
		{"VeryLarge_100000", 100000},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			data := generateTestData(size.size)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = Mean(data)
			}
		})
	}
}

// BenchmarkMedian benchmarks Median function with various data sizes
func BenchmarkMedian(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Small_100", 100},
		{"Medium_1000", 1000},
		{"Large_10000", 10000},
		{"VeryLarge_100000", 100000},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			data := generateTestData(size.size)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = Median(data)
			}
		})
	}
}

// BenchmarkStdDev benchmarks StdDev function with various data sizes
func BenchmarkStdDev(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Small_100", 100},
		{"Medium_1000", 1000},
		{"Large_10000", 10000},
		{"VeryLarge_100000", 100000},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			data := generateTestData(size.size)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = StdDev(data)
			}
		})
	}
}

// BenchmarkQuantile benchmarks Quantile function with various data sizes and percentiles
func BenchmarkQuantile(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Small_100", 100},
		{"Medium_1000", 1000},
		{"Large_10000", 10000},
		{"VeryLarge_100000", 100000},
	}

	percentiles := []float64{0.25, 0.5, 0.75, 0.90, 0.95, 0.99}

	for _, size := range sizes {
		for _, p := range percentiles {
			b.Run(fmt.Sprintf("%s_p%.2f", size.name, p), func(b *testing.B) {
				data := generateTestData(size.size)

				b.ResetTimer()
				b.ReportAllocs()

				for i := 0; i < b.N; i++ {
					_ = Quantile(data, p)
				}
			})
		}
	}
}

// BenchmarkMedianVsMean compares Median (O(n log n)) vs Mean (O(n))
func BenchmarkMedianVsMean(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Small_100", 100},
		{"Medium_1000", 1000},
		{"Large_10000", 10000},
	}

	for _, size := range sizes {
		data := generateTestData(size.size)

		b.Run(fmt.Sprintf("%s_Mean", size.name), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = Mean(data)
			}
		})

		b.Run(fmt.Sprintf("%s_Median", size.name), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = Median(data)
			}
		})
	}
}
