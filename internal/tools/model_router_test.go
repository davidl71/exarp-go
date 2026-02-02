package tools

import (
	"context"
	"errors"
	"testing"
)

func TestDefaultModelRouter_SelectModel(t *testing.T) {
	r := &defaultModelRouter{}

	tests := []struct {
		name         string
		taskType     string
		requirements ModelRequirements
		// We only assert that we get a valid ModelType; actual choice depends on FMAvailable() at runtime.
		wantOneOf []ModelType
	}{
		{"general empty", "", ModelRequirements{}, []ModelType{ModelFM, ModelOllamaLlama}},
		{"general explicit", "general", ModelRequirements{}, []ModelType{ModelFM, ModelOllamaLlama, ModelMLX}},
		{"code task", "code", ModelRequirements{}, []ModelType{ModelFM, ModelOllamaCode, ModelMLX}},
		{"code_analysis", "code_analysis", ModelRequirements{}, []ModelType{ModelFM, ModelOllamaCode, ModelMLX}},
		{"code_generation", "code_generation", ModelRequirements{}, []ModelType{ModelFM, ModelOllamaCode, ModelMLX}},
		{"prefer speed", "general", ModelRequirements{PreferSpeed: true}, []ModelType{ModelFM, ModelOllamaLlama}},
		{"prefer cost general", "general", ModelRequirements{PreferCost: true}, []ModelType{ModelFM, ModelOllamaLlama, ModelMLX}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.SelectModel(tt.taskType, tt.requirements)
			found := false
			for _, w := range tt.wantOneOf {
				if got == w {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("SelectModel() = %v, want one of %v", got, tt.wantOneOf)
			}
		})
	}
}

func TestDefaultModelRouter_Generate_UnknownType(t *testing.T) {
	r := &defaultModelRouter{}
	ctx := context.Background()
	// Unknown model type falls back to FM chain; may succeed if FM/Ollama available or return ErrFMNotSupported.
	_, err := r.Generate(ctx, ModelType("invalid-backend"), "hello", 10, 0.2)
	if err != nil && !errors.Is(err, ErrFMNotSupported) {
		t.Errorf("Generate with unknown type should return nil or ErrFMNotSupported, got %v", err)
	}
}

func TestModelTypeConstants(t *testing.T) {
	if ModelFM == "" || ModelOllamaLlama == "" || ModelOllamaCode == "" || ModelMLX == "" {
		t.Error("model type constants should be non-empty")
	}
}
