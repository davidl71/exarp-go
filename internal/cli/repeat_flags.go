// repeat_flags.go — Collect repeated long-form flags from raw argv (ParseArgs keeps only the last value per key).
package cli

import (
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/taskworkflowspec"
)

// CollectRepeatedStringFlag returns every value passed for --flag or --flag=value across argv (order preserved).
func CollectRepeatedStringFlag(argv []string, flag string) []string {
	if flag == "" || len(argv) == 0 {
		return nil
	}
	long := "--" + flag
	longEq := long + "="
	var out []string
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		if strings.HasPrefix(a, longEq) {
			if v := strings.TrimSpace(strings.TrimPrefix(a, longEq)); v != "" {
				out = append(out, v)
			}
			continue
		}
		if a != long {
			continue
		}
		if i+1 < len(argv) {
			next := argv[i+1]
			if next != "" && next[0] != '-' {
				if v := strings.TrimSpace(next); v != "" {
					out = append(out, v)
				}
				i++
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// MergeTaskTagsFromCSVAndRepeated merges --tags CSV with repeated --tag values, deduping by first occurrence.
func MergeTaskTagsFromCSVAndRepeated(tagsCSV string, argv []string) []string {
	repeated := CollectRepeatedStringFlag(argv, "tag")
	var merged []string
	if strings.TrimSpace(tagsCSV) != "" {
		merged = append(merged, taskworkflowspec.CSVToList(tagsCSV)...)
	}
	merged = append(merged, repeated...)
	return models.NormalizeTags(merged)
}
