// task_analysis_tags_llm.go — Tag enrichment: LLM batch prompts, parsing, scoring, Ollama/FM bridges.
// See also: task_analysis_tags.go, task_analysis_tags_discover.go
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/davidl71/exarp-go/internal/database"
	"sort"
	"strings"
	"time"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   buildBatchTagPrompt
//   extractOllamaGenerateResponseText — extractOllamaGenerateResponseText extracts the "response" field from Ollama generate JSON.
//   parseBatchTagResponse — parseBatchTagResponse parses LLM response into map[file path][]suggested tags.
//   parseTaskTagResponse — parseTaskTagResponse parses LLM response into map[task_id][]suggested tags (same JSON shape).
//   parseTagResponseJSON — parseTagResponseJSON parses a JSON object of the form {"id": ["tag1","tag2"], ...} into map[string][]string.
//   taskTagLLMBatchItem — taskTagLLMBatchItem holds one task's data for LLM tag suggestion (action=tags semantic pass).
//   maxTaskSnippetLen
//   buildTaskBatchTagPrompt — buildTaskBatchTagPrompt builds one prompt for a batch of tasks (title + content snippet + existing tags).
//   buildTaskBatchTagPromptMatchOnly — buildTaskBatchTagPromptMatchOnly builds a short prompt for quick FM inference: pick 0-3 tags only from allowed list (no free-form).
//   filterSuggestionsToAllowed — filterSuggestionsToAllowed keeps only tags that are in the allowed set (for match_existing_only).
//   llmTagProfile — llmTagProfile holds timing profile for the LLM tag inference phase.
//   enrichTaskTagSuggestionsWithLLM — enrichTaskTagSuggestionsWithLLM runs Apple FM, Ollama, or MLX on batches of tasks and merges suggested tags into analysis.TagSuggestions.
//   modelTypeToMethodString — modelTypeToMethodString maps ModelType to the profiling method string used in tag enrichment.
//   enrichTagsWithOllama — enrichTagsWithOllama uses Ollama to infer additional tags, in batches (same smart prompt as Apple FM).
//   matchDiscoveredTagsToTasks — matchDiscoveredTagsToTasks matches discovered tags from files to related tasks.
// ────────────────────────────────────────────────────────────────────────────

// ─── buildBatchTagPrompt ────────────────────────────────────────────────────
func buildBatchTagPrompt(batch []map[string]interface{}, canonicalTags, projectTags []string) string {
	var b strings.Builder

	b.WriteString("Suggest 0-3 tags per file from path/context. Prefer: ")
	b.WriteString(strings.Join(canonicalTags, ", "))
	b.WriteString(". Project tags: ")
	b.WriteString(strings.Join(projectTags, ", "))
	b.WriteString(".\n\nFiles (path → existing tags):\n")

	for _, d := range batch {
		file, _ := d["file"].(string)
		existing, _ := d["tags"].([]string)

		b.WriteString(file)
		b.WriteString(": ")

		enc, _ := json.Marshal(existing)
		b.WriteString(string(enc))
		b.WriteString("\n")
	}

	b.WriteString("\nReturn ONLY a JSON object: {\"path\": [\"tag1\",\"tag2\"]}. No other text. Use file path as key. Empty array [] if none.")

	return b.String()
}

// ─── extractOllamaGenerateResponseText ──────────────────────────────────────
// extractOllamaGenerateResponseText extracts the "response" field from Ollama generate JSON.
// handleOllamaGenerate returns Text as a JSON object with "response" containing the actual LLM output.
// Returns the inner response string, or the original text if not in that format.
func extractOllamaGenerateResponseText(text string) string {
	text = strings.TrimSpace(text)

	var raw map[string]interface{}

	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return text
	}

	if r, ok := raw["response"].(string); ok && r != "" {
		return r
	}

	return text
}

// ─── parseBatchTagResponse ──────────────────────────────────────────────────
// parseBatchTagResponse parses LLM response into map[file path][]suggested tags.
func parseBatchTagResponse(response string) map[string][]string {
	return parseTagResponseJSON(response)
}

// ─── parseTaskTagResponse ───────────────────────────────────────────────────
// parseTaskTagResponse parses LLM response into map[task_id][]suggested tags (same JSON shape).
func parseTaskTagResponse(response string) map[string][]string {
	return parseTagResponseJSON(response)
}

// ─── parseTagResponseJSON ───────────────────────────────────────────────────
// parseTagResponseJSON parses a JSON object of the form {"id": ["tag1","tag2"], ...} into map[string][]string.
func parseTagResponseJSON(response string) map[string][]string {
	response = strings.TrimSpace(response)
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start < 0 || end <= start {
		return nil
	}

	response = response[start : end+1]

	var raw map[string]interface{}

	if err := json.Unmarshal([]byte(response), &raw); err != nil {
		return nil
	}

	out := make(map[string][]string)

	for id, v := range raw {
		var tags []string

		switch arr := v.(type) {
		case []interface{}:
			for _, x := range arr {
				if s, ok := x.(string); ok && s != "" {
					tags = append(tags, s)
				}
			}

			out[id] = tags
		case []string:
			out[id] = arr
		}
	}

	return out
}

// ─── taskTagLLMBatchItem ────────────────────────────────────────────────────
// taskTagLLMBatchItem holds one task's data for LLM tag suggestion (action=tags semantic pass).
type taskTagLLMBatchItem struct {
	TaskID       string
	Title        string
	Snippet      string
	ExistingTags []string
}

// ─── maxTaskSnippetLen ──────────────────────────────────────────────────────
const maxTaskSnippetLen = 400

// ─── buildTaskBatchTagPrompt ────────────────────────────────────────────────
// buildTaskBatchTagPrompt builds one prompt for a batch of tasks (title + content snippet + existing tags).
func buildTaskBatchTagPrompt(batch []taskTagLLMBatchItem, canonicalTags, projectTags []string) string {
	var b strings.Builder

	b.WriteString("Suggest 0-3 tags per task from title and context. Prefer: ")
	b.WriteString(strings.Join(canonicalTags, ", "))
	b.WriteString(". Project tags: ")
	b.WriteString(strings.Join(projectTags, ", "))
	b.WriteString(".\n\nTasks (task_id → title, snippet, existing tags):\n")

	for _, item := range batch {
		b.WriteString(item.TaskID)
		b.WriteString(": title=")
		b.WriteString(item.Title)
		b.WriteString(" snippet=")

		if len(item.Snippet) > maxTaskSnippetLen {
			b.WriteString(item.Snippet[:maxTaskSnippetLen])
			b.WriteString("...")
		} else {
			b.WriteString(item.Snippet)
		}

		b.WriteString(" existing=")

		enc, _ := json.Marshal(item.ExistingTags)
		b.WriteString(string(enc))
		b.WriteString("\n")
	}

	b.WriteString("\nReturn ONLY a JSON object: {\"T-123\": [\"tag1\",\"tag2\"]}. Use task_id as key. No other text. Empty array [] if none.")

	return b.String()
}

// ─── buildTaskBatchTagPromptMatchOnly ───────────────────────────────────────
// buildTaskBatchTagPromptMatchOnly builds a short prompt for quick FM inference: pick 0-3 tags only from allowed list (no free-form).
func buildTaskBatchTagPromptMatchOnly(batch []taskTagLLMBatchItem, allowedTags []string) string {
	var b strings.Builder

	b.WriteString("Pick 0-3 tags per task from this list only: ")
	b.WriteString(strings.Join(allowedTags, ", "))
	b.WriteString(".\n\nTasks (task_id, title, snippet):\n")

	for _, item := range batch {
		b.WriteString(item.TaskID)
		b.WriteString(": ")
		b.WriteString(item.Title)
		b.WriteString(" | ")

		snippet := item.Snippet
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}

		b.WriteString(snippet)
		b.WriteString("\n")
	}

	b.WriteString("\nReturn ONLY JSON: {\"T-123\": [\"tag1\",\"tag2\"]}. Use task_id as key. Only tags from the list. Empty [] if none.")

	return b.String()
}

// ─── filterSuggestionsToAllowed ─────────────────────────────────────────────
// filterSuggestionsToAllowed keeps only tags that are in the allowed set (for match_existing_only).
func filterSuggestionsToAllowed(suggestionsByTask map[string][]string, allowed map[string]bool) map[string][]string {
	out := make(map[string][]string)

	for taskID, tags := range suggestionsByTask {
		filtered := make([]string, 0, len(tags))

		seen := make(map[string]bool)
		for _, tag := range tags {
			if allowed[tag] && !seen[tag] {
				seen[tag] = true

				filtered = append(filtered, tag)
			}
		}

		if len(filtered) > 0 {
			out[taskID] = filtered
		}
	}

	return out
}

// ─── llmTagProfile ──────────────────────────────────────────────────────────
// llmTagProfile holds timing profile for the LLM tag inference phase.
type llmTagProfile struct {
	TotalMs    int64
	Batches    int
	PerBatchMs []int64
}

// ─── enrichTaskTagSuggestionsWithLLM ────────────────────────────────────────
// enrichTaskTagSuggestionsWithLLM runs Apple FM, Ollama, or MLX on batches of tasks and merges suggested tags into analysis.TagSuggestions.
// Uses tag cache everywhere as quick hints. When matchExistingOnly is true, uses a constrained prompt and filters output to allowed tags only.
// When useTinyTagModel is true, tries Ollama with tinyllama first, then MLX with TinyLlama (faster small models) before Apple FM.
// Returns method, processed count, and profile (LLM total ms, batch count, per-batch ms) for profiling.
func enrichTaskTagSuggestionsWithLLM(ctx context.Context, tasks []Todo2Task, analysis *TagAnalysis, batchSizeParam int, matchExistingOnly bool, useTinyTagModel bool) (method string, processed int, profile llmTagProfile) {
	batchSize := effectiveLLMBatchSize(batchSizeParam)
	canonicalTags := getCanonicalTagsList()

	projectTags := projectTagsWithCacheHints(tasks, 30)
	if len(projectTags) > 40 {
		projectTags = projectTags[:40]
	}

	allowedSet := make(map[string]bool)
	for _, t := range canonicalTags {
		allowedSet[t] = true
	}

	for _, t := range projectTags {
		allowedSet[t] = true
	}

	const totalTimeout = 120 * time.Second

	totalCtx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()

	// Build flat list of task batch items (title, snippet, existing = Tags + TagSuggestions + cached suggestions)
	items := make([]taskTagLLMBatchItem, 0, len(tasks))

	for _, t := range tasks {
		existing := make(map[string]bool)
		for _, tag := range t.Tags {
			existing[tag] = true
		}

		for _, tag := range analysis.TagSuggestions[t.ID] {
			existing[tag] = true
		}

		cached, _ := database.GetTaskTagSuggestions(t.ID)
		for _, tag := range cached {
			existing[tag] = true
		}

		existingList := make([]string, 0, len(existing))
		for tag := range existing {
			existingList = append(existingList, tag)
		}

		sort.Strings(existingList)

		snippet := strings.TrimSpace(t.LongDescription)
		if snippet == "" {
			snippet = t.Content
		}

		items = append(items, taskTagLLMBatchItem{
			TaskID:       t.ID,
			Title:        t.Content,
			Snippet:      snippet,
			ExistingTags: existingList,
		})
	}

	mergeSuggestions := func(suggestionsByTask map[string][]string) {
		if matchExistingOnly {
			suggestionsByTask = filterSuggestionsToAllowed(suggestionsByTask, allowedSet)
		}

		for taskID, suggested := range suggestionsByTask {
			if len(suggested) == 0 {
				continue
			}

			processed++

			existingSet := make(map[string]bool)
			for _, tag := range analysis.TagSuggestions[taskID] {
				existingSet[tag] = true
			}

			for _, tag := range suggested {
				if !existingSet[tag] {
					existingSet[tag] = true

					analysis.TagSuggestions[taskID] = append(analysis.TagSuggestions[taskID], tag)
				}
			}
		}
	}

	llmStart := time.Now()

	// When useTinyTagModel: try Ollama(tinyllama) then MLX(TinyLlama) for faster inference
	if useTinyTagModel {
		ollama := DefaultOllama()
		if ollama != nil {
			tinyProcessed := 0

			for start := 0; start < len(items); start += batchSize {
				select {
				case <-totalCtx.Done():
					profile.TotalMs = time.Since(llmStart).Milliseconds()
					profile.Batches = len(profile.PerBatchMs)

					return "ollama", processed, profile
				default:
				}

				end := start + batchSize
				if end > len(items) {
					end = len(items)
				}

				batch := items[start:end]
				batchStart := time.Now()

				var prompt string

				if matchExistingOnly {
					allowedList := make([]string, 0, len(allowedSet))
					for tag := range allowedSet {
						allowedList = append(allowedList, tag)
					}

					sort.Strings(allowedList)
					prompt = buildTaskBatchTagPromptMatchOnly(batch, allowedList)
				} else {
					prompt = buildTaskBatchTagPrompt(batch, canonicalTags, projectTags)
				}

				batchCtx, batchCancel := context.WithTimeout(totalCtx, batchTimeout(len(batch)))
				result, err := ollama.Invoke(batchCtx, map[string]interface{}{
					"action":      "generate",
					"model":       ollamaTinyTagModel,
					"prompt":      prompt,
					"max_tokens":  llmTagMaxTokens,
					"temperature": 0.2,
				})

				batchCancel()

				profile.PerBatchMs = append(profile.PerBatchMs, time.Since(batchStart).Milliseconds())

				if err != nil || len(result) == 0 {
					continue
				}

				response := extractOllamaGenerateResponseText(result[0].Text)
				mergeSuggestions(parseTaskTagResponse(response))

				tinyProcessed++
			}

			if tinyProcessed > 0 {
				profile.TotalMs = time.Since(llmStart).Milliseconds()
				profile.Batches = len(profile.PerBatchMs)

				return "ollama", processed, profile
			}
		}
		// Try MLX with TinyLlama (reset profile so we only report MLX times)
		profile = llmTagProfile{}
		mlxProcessed := 0

		for start := 0; start < len(items); start += batchSize {
			select {
			case <-totalCtx.Done():
				profile.TotalMs = time.Since(llmStart).Milliseconds()
				profile.Batches = len(profile.PerBatchMs)

				if mlxProcessed > 0 {
					return "mlx", processed, profile
				}

				goto fallthroughTiny
			default:
			}

			end := start + batchSize
			if end > len(items) {
				end = len(items)
			}

			batch := items[start:end]
			batchStart := time.Now()

			var prompt string

			if matchExistingOnly {
				allowedList := make([]string, 0, len(allowedSet))
				for tag := range allowedSet {
					allowedList = append(allowedList, tag)
				}

				sort.Strings(allowedList)
				prompt = buildTaskBatchTagPromptMatchOnly(batch, allowedList)
			} else {
				prompt = buildTaskBatchTagPrompt(batch, canonicalTags, projectTags)
			}

			batchCtx, batchCancel := context.WithTimeout(totalCtx, batchTimeout(len(batch)))
			raw, err := InvokeMLXTool(batchCtx, map[string]interface{}{
				"action":      "generate",
				"model":       mlxTinyTagModel,
				"prompt":      prompt,
				"max_tokens":  llmTagMaxTokens,
				"temperature": 0.2,
			})

			batchCancel()

			profile.PerBatchMs = append(profile.PerBatchMs, time.Since(batchStart).Milliseconds())

			if err != nil {
				continue
			}

			response, err := parseGeneratedTextFromMLXResponse(raw)
			if err != nil || response == "" {
				continue
			}

			mergeSuggestions(parseTaskTagResponse(response))

			mlxProcessed++
		}

		if mlxProcessed > 0 {
			profile.TotalMs = time.Since(llmStart).Milliseconds()
			profile.Batches = len(profile.PerBatchMs)

			return "mlx", processed, profile
		}

	fallthroughTiny:
		// Tiny path failed; fall through to ModelRouter (FM → Ollama → MLX)
	}

	// Use DefaultModelRouter for tag enrichment: FM chain, then Ollama, then MLX.
	requirements := ModelRequirements{PreferSpeed: useTinyTagModel}
	model := DefaultModelRouter.SelectModel("general", requirements)
	method = modelTypeToMethodString(model)

	for start := 0; start < len(items); start += batchSize {
		select {
		case <-totalCtx.Done():
			profile.TotalMs = time.Since(llmStart).Milliseconds()
			profile.Batches = len(profile.PerBatchMs)

			return method, processed, profile
		default:
		}

		end := start + batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[start:end]
		batchStart := time.Now()

		var prompt string

		if matchExistingOnly {
			allowedList := make([]string, 0, len(allowedSet))
			for tag := range allowedSet {
				allowedList = append(allowedList, tag)
			}

			sort.Strings(allowedList)
			prompt = buildTaskBatchTagPromptMatchOnly(batch, allowedList)
		} else {
			prompt = buildTaskBatchTagPrompt(batch, canonicalTags, projectTags)
		}

		batchCtx, batchCancel := context.WithTimeout(totalCtx, batchTimeout(len(batch)))
		response, err := DefaultModelRouter.Generate(batchCtx, model, prompt, llmTagMaxTokens, 0.2)

		batchCancel()

		profile.PerBatchMs = append(profile.PerBatchMs, time.Since(batchStart).Milliseconds())

		if err != nil {
			continue
		}

		mergeSuggestions(parseTaskTagResponse(response))
	}

	profile.TotalMs = time.Since(llmStart).Milliseconds()
	profile.Batches = len(profile.PerBatchMs)

	return method, processed, profile
}

// ─── modelTypeToMethodString ────────────────────────────────────────────────
// modelTypeToMethodString maps ModelType to the profiling method string used in tag enrichment.
func modelTypeToMethodString(m ModelType) string {
	switch m {
	case ModelFM:
		return "apple_fm"
	case ModelOllamaLlama, ModelOllamaCode:
		return "ollama"
	case ModelMLX:
		return "mlx"
	default:
		return "model_router"
	}
}

// ─── enrichTagsWithOllama ───────────────────────────────────────────────────
// enrichTagsWithOllama uses Ollama to infer additional tags, in batches (same smart prompt as Apple FM).
// Uses tag cache everywhere for LLM hints.
func enrichTagsWithOllama(ctx context.Context, ollama OllamaProvider, discoveries []map[string]interface{}, tasks []Todo2Task, batchSizeParam int) []map[string]interface{} {
	batchSize := effectiveLLMBatchSize(batchSizeParam)
	canonicalTags := getCanonicalTagsList()

	projectTags := projectTagsWithCacheHints(tasks, 30)
	if len(projectTags) > 40 {
		projectTags = projectTags[:40]
	}

	totalCtx, totalCancel := context.WithTimeout(ctx, llmTotalTimeout)
	defer totalCancel()

	for start := 0; start < len(discoveries); start += batchSize {
		select {
		case <-totalCtx.Done():
			return discoveries
		default:
		}

		end := start + batchSize
		if end > len(discoveries) {
			end = len(discoveries)
		}

		batch := discoveries[start:end]

		prompt := buildBatchTagPrompt(batch, canonicalTags, projectTags)
		batchCtx, batchCancel := context.WithTimeout(totalCtx, batchTimeout(len(batch)))
		result, err := ollama.Invoke(batchCtx, map[string]interface{}{
			"action":      "generate",
			"prompt":      prompt,
			"max_tokens":  800,
			"temperature": 0.2,
		})

		batchCancel()

		if err != nil || len(result) == 0 {
			if errors.Is(batchCtx.Err(), context.DeadlineExceeded) {
				for i := start; i < end; i++ {
					discoveries[i]["llm_timeout"] = true
				}
			}

			continue
		}

		response := extractOllamaGenerateResponseText(result[0].Text)
		suggestionsByFile := parseBatchTagResponse(response)

		for i, discovery := range batch {
			file, _ := discovery["file"].(string)
			existingTags, _ := discovery["tags"].([]string)

			suggestedTags := suggestionsByFile[file]
			if len(suggestedTags) == 0 {
				continue
			}

			if discoveries[start+i]["llm_suggestions"] == nil {
				discoveries[start+i]["llm_suggestions"] = suggestedTags
			}

			allTags := append(existingTags, suggestedTags...)
			seen := make(map[string]bool)
			uniqueTags := []string{}

			for _, tag := range allTags {
				if !seen[tag] {
					seen[tag] = true

					uniqueTags = append(uniqueTags, tag)
				}
			}

			discoveries[start+i]["tags"] = uniqueTags
		}
	}

	return discoveries
}

// ─── matchDiscoveredTagsToTasks ─────────────────────────────────────────────
// matchDiscoveredTagsToTasks matches discovered tags from files to related tasks.
func matchDiscoveredTagsToTasks(discoveries []map[string]interface{}, tasks []Todo2Task) map[string][]string {
	matches := make(map[string][]string)

	for _, discovery := range discoveries {
		file, _ := discovery["file"].(string)

		tags, ok := discovery["tags"].([]string)
		if !ok {
			continue
		}

		// Match tasks that reference this file or have related content
		for _, task := range tasks {
			matched := false

			// Check if task references the file
			if strings.Contains(task.Content, file) || strings.Contains(task.LongDescription, file) {
				matched = true
			}

			// Check if task was discovered from this file
			if df, ok := task.Metadata["discovered_from"].(string); ok && df == file {
				matched = true
			}

			// Check planning doc link
			if pd, ok := task.Metadata["planning_doc"].(string); ok && pd == file {
				matched = true
			}

			if matched {
				// Apply canonical rules to discovered tags
				canonicalRules := CanonicalTagRules()

				for _, tag := range tags {
					finalTag := tag
					if newTag, exists := canonicalRules[tag]; exists && newTag != "" {
						finalTag = newTag
					}
					// Check if task already has this tag
					hasTag := false

					for _, existingTag := range task.Tags {
						if existingTag == finalTag {
							hasTag = true
							break
						}
					}

					if !hasTag {
						matches[task.ID] = append(matches[task.ID], finalTag)
					}
				}
			}
		}
	}

	// Dedupe matches
	for taskID, tags := range matches {
		seen := make(map[string]bool)
		unique := []string{}

		for _, tag := range tags {
			if !seen[tag] {
				seen[tag] = true

				unique = append(unique, tag)
			}
		}

		matches[taskID] = unique
	}

	return matches
}

// handleTaskAnalysisDependencies handles dependency analysis.
