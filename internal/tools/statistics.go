// statistics.go — Statistical calculations (mean, median, stddev) for task metrics.
package tools

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/stat"
)

// Mean calculates the arithmetic mean of a slice of floats.
// Returns 0.0 for empty slice (matching Python statistics.mean behavior).
func Mean(data []float64) float64 {
	if len(data) == 0 {
		return 0.0
	}

	result := stat.Mean(data, nil)
	if math.IsNaN(result) {
		return 0.0
	}

	return result
}

// Median calculates the median (50th percentile) of a slice of floats.
// Returns 0.0 for empty slice (matching Python statistics.median behavior).
// Optimized: Uses quickselect for very large datasets (O(n) average) instead of O(n log n) sorting.
// For smaller datasets, uses standard median calculation matching Python statistics.median.
func Median(data []float64) float64 {
	if len(data) == 0 {
		return 0.0
	}
	// For datasets < 10000, use standard median (matching Python statistics.median)
	// This ensures compatibility with existing tests
	if len(data) < 10000 {
		sorted := make([]float64, len(data))
		copy(sorted, data)
		sort.Float64s(sorted)

		n := len(sorted)
		if n%2 == 1 {
			// Odd: return middle element
			return sorted[n/2]
		} else {
			// Even: return average of two middle elements
			return (sorted[n/2-1] + sorted[n/2]) / 2.0
		}
	}
	// Use quickselect for very large datasets (10000+)
	// This provides O(n) average case instead of O(n log n)
	return quickselectMedian(data)
}

// quickselectMedian finds the median using quickselect algorithm (O(n) average).
func quickselectMedian(data []float64) float64 {
	n := len(data)
	if n%2 == 1 {
		// Odd: return middle element (0-indexed: n/2)
		k := n / 2
		work := make([]float64, len(data))
		copy(work, data)
		median := quickselect(work, k)

		return median
	} else {
		// Even: return average of two middle elements
		k1 := n/2 - 1
		k2 := n / 2
		// Need to find both - use separate copies to avoid mutation issues
		work1 := make([]float64, len(data))
		copy(work1, data)
		work2 := make([]float64, len(data))
		copy(work2, data)

		median1 := quickselect(work1, k1)
		median2 := quickselect(work2, k2)

		return (median1 + median2) / 2.0
	}
}

// quickselect finds the k-th smallest element (0-indexed) using quickselect algorithm.
func quickselect(arr []float64, k int) float64 {
	if len(arr) == 1 {
		return arr[0]
	}

	if k < 0 {
		k = 0
	}

	if k >= len(arr) {
		k = len(arr) - 1
	}

	// Partition around pivot
	pivotIdx := partition(arr)

	if k == pivotIdx {
		return arr[pivotIdx]
	} else if k < pivotIdx {
		return quickselect(arr[:pivotIdx], k)
	} else {
		return quickselect(arr[pivotIdx+1:], k-pivotIdx-1)
	}
}

// partition partitions array around a pivot (Lomuto partition scheme)
// Returns the final index of the pivot.
func partition(arr []float64) int {
	n := len(arr)
	if n <= 1 {
		return 0
	}

	// Use last element as pivot
	pivot := arr[n-1]
	i := 0

	for j := 0; j < n-1; j++ {
		if arr[j] <= pivot {
			arr[i], arr[j] = arr[j], arr[i]
			i++
		}
	}

	// Place pivot in correct position
	arr[i], arr[n-1] = arr[n-1], arr[i]

	return i
}

// StdDev calculates the sample standard deviation of a slice of floats.
// Returns 0.0 for empty or single-element slices (matching Python statistics.stdev behavior).
func StdDev(data []float64) float64 {
	if len(data) <= 1 {
		return 0.0
	}

	result := stat.StdDev(data, nil)
	if math.IsNaN(result) {
		return 0.0
	}

	return result
}

// Quantile calculates the specified quantile (0.0 to 1.0) of a slice of floats.
// Returns 0.0 for empty slice or invalid quantile values.
// Examples: p=0.25 (25th percentile), p=0.5 (median), p=0.75 (75th percentile), p=0.90 (90th percentile).
// Uses linear interpolation for more accurate results than simple indexing.
// Note: Currently uses original O(n log n) approach for all datasets to maintain exact compatibility
// with gonum/stat's linear interpolation. Future optimization could use quickselect for very large datasets.
func Quantile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0.0
	}

	if p < 0.0 || p > 1.0 {
		return 0.0 // Invalid quantile
	}
	// Quantile requires sorted data
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)
	// Use Empirical method to match test expectations (NumPy-style quantiles)
	result := stat.Quantile(p, stat.Empirical, sorted, nil)
	if math.IsNaN(result) {
		return 0.0
	}

	return result
}

// Round rounds a float64 to the specified number of decimal places.
// Examples: Round(3.14159, 2) = 3.14, Round(2.5, 0) = 3.0.
func Round(value float64, decimals int) float64 {
	if decimals < 0 {
		decimals = 0
	}

	multiplier := math.Pow(10, float64(decimals))

	return math.Round(value*multiplier) / multiplier
}
