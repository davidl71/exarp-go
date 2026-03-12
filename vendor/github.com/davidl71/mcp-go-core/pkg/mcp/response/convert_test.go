package response

import (
	"reflect"
	"testing"
)

func TestConvertToMap(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:  "already a map",
			input: map[string]interface{}{"key": "value", "num": 42},
			want:  map[string]interface{}{"key": "value", "num": 42},
		},
		{
			name: "struct with json tags",
			input: struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
				Count   int    `json:"count"`
			}{
				Success: true,
				Message: "test",
				Count:   10,
			},
			want: map[string]interface{}{
				"success": true,
				"message": "test",
				"count":   float64(10), // JSON numbers become float64
			},
		},
		{
			name: "struct without json tags",
			input: struct {
				Field1 string
				Field2 int
			}{
				Field1: "value1",
				Field2: 20,
			},
			want: map[string]interface{}{
				"Field1": "value1",
				"Field2": float64(20),
			},
		},
		{
			name: "nested struct",
			input: struct {
				Data struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				} `json:"data"`
			}{
				Data: struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}{
					ID:   1,
					Name: "test",
				},
			},
			want: map[string]interface{}{
				"data": map[string]interface{}{
					"id":   float64(1),
					"name": "test",
				},
			},
		},
		{
			name:    "nil input",
			input:   nil,
			want:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "empty map",
			input:   map[string]interface{}{},
			want:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "map with various types",
			input: map[string]interface{}{
				"string": "value",
				"int":    42,
				"bool":   true,
				"float":  3.14,
				"slice":  []string{"a", "b"},
			},
			want: map[string]interface{}{
				"string": "value",
				"int":    42,
				"bool":   true,
				"float":  3.14,
				"slice":  []string{"a", "b"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Compare maps (JSON unmarshaling may convert ints to float64)
			if len(got) != len(tt.want) {
				t.Errorf("ConvertToMap() = %v, want %v", got, tt.want)
				return
			}

			for key, wantValue := range tt.want {
				gotValue, ok := got[key]
				if !ok {
					t.Errorf("ConvertToMap() missing key %q", key)
					continue
				}

				// For nested maps, recursively compare
				if wantMap, ok := wantValue.(map[string]interface{}); ok {
					gotMap, ok := gotValue.(map[string]interface{})
					if !ok {
						t.Errorf("ConvertToMap() key %q: got %T, want map[string]interface{}", key, gotValue)
						continue
					}
					if len(gotMap) != len(wantMap) {
						t.Errorf("ConvertToMap() key %q: got map length %d, want %d", key, len(gotMap), len(wantMap))
						continue
					}
					for k, v := range wantMap {
						if gotMap[k] != v {
							t.Errorf("ConvertToMap() key %q[%q]: got %v, want %v", key, k, gotMap[k], v)
						}
					}
				} else {
					// For primitive types and slices, use reflect.DeepEqual
					if !reflect.DeepEqual(gotValue, wantValue) {
						// Allow float64 comparison for ints (JSON unmarshaling behavior)
						if wantInt, ok := wantValue.(int); ok {
							if gotFloat, ok := gotValue.(float64); ok && float64(wantInt) == gotFloat {
								continue
							}
						}
						t.Errorf("ConvertToMap() key %q: got %v (%T), want %v (%T)", key, gotValue, gotValue, wantValue, wantValue)
					}
				}
			}
		})
	}
}

func TestConvertToMap_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "channel (not JSON serializable)",
			input:   make(chan int),
			wantErr: true,
		},
		{
			name:    "function (not JSON serializable)",
			input:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ConvertToMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
