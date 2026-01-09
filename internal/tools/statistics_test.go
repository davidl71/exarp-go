package tools

import (
	"math"
	"testing"
)

func TestMean(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		expected float64
	}{
		{
			name:     "empty slice",
			data:     []float64{},
			expected: 0.0,
		},
		{
			name:     "single value",
			data:     []float64{5.0},
			expected: 5.0,
		},
		{
			name:     "multiple values",
			data:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 3.0,
		},
		{
			name:     "negative values",
			data:     []float64{-1.0, -2.0, -3.0},
			expected: -2.0,
		},
		{
			name:     "mixed positive and negative",
			data:     []float64{-2.0, 0.0, 2.0},
			expected: 0.0,
		},
		{
			name:     "decimal values",
			data:     []float64{2.5, 3.0, 4.5, 2.0, 3.5},
			expected: 3.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Mean(tt.data)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("Mean(%v) = %v, want %v", tt.data, result, tt.expected)
			}
		})
	}
}

func TestMedian(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		expected float64
	}{
		{
			name:     "empty slice",
			data:     []float64{},
			expected: 0.0,
		},
		{
			name:     "single value",
			data:     []float64{5.0},
			expected: 5.0,
		},
		{
			name:     "odd number of values",
			data:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 3.0,
		},
		{
			name:     "even number of values",
			data:     []float64{1.0, 2.0, 3.0, 4.0},
			expected: 2.5, // Average of 2 and 3
		},
		{
			name:     "unsorted data",
			data:     []float64{5.0, 1.0, 3.0, 2.0, 4.0},
			expected: 3.0,
		},
		{
			name:     "duplicate values",
			data:     []float64{2.0, 2.0, 2.0, 2.0},
			expected: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Median(tt.data)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("Median(%v) = %v, want %v", tt.data, result, tt.expected)
			}
		})
	}
}

func TestStdDev(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		expected float64
	}{
		{
			name:     "empty slice",
			data:     []float64{},
			expected: 0.0,
		},
		{
			name:     "single value",
			data:     []float64{5.0},
			expected: 0.0,
		},
		{
			name:     "two values",
			data:     []float64{1.0, 3.0},
			expected: 1.4142135623730951, // sqrt(2)
		},
		{
			name:     "multiple values",
			data:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 1.5811388300841898, // sqrt(2.5)
		},
		{
			name:     "constant values",
			data:     []float64{5.0, 5.0, 5.0},
			expected: 0.0,
		},
		{
			name:     "decimal values",
			data:     []float64{2.5, 3.0, 4.5, 2.0, 3.5},
			expected: 0.8944271909999159, // Approximate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StdDev(tt.data)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("StdDev(%v) = %v, want %v", tt.data, result, tt.expected)
			}
		})
	}
}

func TestQuantile(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		p        float64
		expected float64
	}{
		{
			name:     "empty slice",
			data:     []float64{},
			p:        0.5,
			expected: 0.0,
		},
		{
			name:     "single value, p=0.5",
			data:     []float64{5.0},
			p:        0.5,
			expected: 5.0,
		},
		{
			name:     "p25 (25th percentile)",
			data:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			p:        0.25,
			expected: 2.0,
		},
		{
			name:     "p50 (median)",
			data:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			p:        0.5,
			expected: 3.0,
		},
		{
			name:     "p75 (75th percentile)",
			data:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			p:        0.75,
			expected: 4.0,
		},
		{
			name:     "p90 (90th percentile)",
			data:     []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0},
			p:        0.90,
			expected: 9.0,
		},
		{
			name:     "p=0.0 (minimum)",
			data:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			p:        0.0,
			expected: 1.0,
		},
		{
			name:     "p=1.0 (maximum)",
			data:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			p:        1.0,
			expected: 5.0,
		},
		{
			name:     "invalid p < 0",
			data:     []float64{1.0, 2.0, 3.0},
			p:        -0.1,
			expected: 0.0,
		},
		{
			name:     "invalid p > 1",
			data:     []float64{1.0, 2.0, 3.0},
			p:        1.1,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Quantile(tt.data, tt.p)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("Quantile(%v, %v) = %v, want %v", tt.data, tt.p, result, tt.expected)
			}
		})
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		decimals int
		expected float64
	}{
		{
			name:     "round to 2 decimals",
			value:    3.14159,
			decimals: 2,
			expected: 3.14,
		},
		{
			name:     "round to 0 decimals",
			value:    2.5,
			decimals: 0,
			expected: 3.0,
		},
		{
			name:     "round to 1 decimal",
			value:    2.55,
			decimals: 1,
			expected: 2.6,
		},
		{
			name:     "round negative number",
			value:    -3.14159,
			decimals: 2,
			expected: -3.14,
		},
		{
			name:     "round to 3 decimals",
			value:    1.234567,
			decimals: 3,
			expected: 1.235,
		},
		{
			name:     "negative decimals (should treat as 0)",
			value:    3.14159,
			decimals: -1,
			expected: 3.0,
		},
		{
			name:     "zero value",
			value:    0.0,
			decimals: 2,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Round(tt.value, tt.decimals)
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("Round(%v, %v) = %v, want %v", tt.value, tt.decimals, result, tt.expected)
			}
		})
	}
}

// TestMeanStdDevCombined tests that Mean and StdDev can be used together
// (simulating the Python pattern of calculating both)
func TestMeanStdDevCombined(t *testing.T) {
	data := []float64{2.5, 3.0, 4.5, 2.0, 3.5}
	mean := Mean(data)
	stddev := StdDev(data)

	// Verify both are valid numbers
	if math.IsNaN(mean) || math.IsNaN(stddev) {
		t.Errorf("Mean or StdDev returned NaN: mean=%v, stddev=%v", mean, stddev)
	}

	// Verify reasonable values
	if mean < 0 || mean > 10 {
		t.Errorf("Mean out of expected range: %v", mean)
	}
	if stddev < 0 || stddev > 10 {
		t.Errorf("StdDev out of expected range: %v", stddev)
	}
}

