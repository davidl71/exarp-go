# Wisdom Integration Plan for Scorecards and Reports

**Date:** 2026-01-13  
**Status:** Planning  
**Related:** `PROJECT_SCORECARD_WITH_WISDOM.md` (example implementation)

## Executive Summary

This plan outlines where and how to integrate devwisdom-go wisdom into exarp-go's report generation system. Wisdom provides contextual encouragement and guidance based on project scores, making reports more actionable and inspiring.

## Current State Analysis

### ✅ Wisdom Integration Points (Already Working)

1. **Briefing Action** (`handleReportBriefing`)
   - ✅ Uses wisdom engine directly
   - ✅ Gets quotes from multiple sources (up to 3)
   - ✅ Uses project score to determine aeon level
   - ✅ Location: `internal/tools/report.go:66-129`

2. **Memory Maintenance - Dream Action** (`handleMemoryMaintDream`)
   - ✅ Uses wisdom engine for dream insights
   - ✅ Location: `internal/tools/memory_maint.go:551-607`

3. **Recommend Tool - Advisor Action** (`handleRecommendAdvisorNative`)
   - ✅ Uses advisors for metric/tool/stage-specific guidance
   - ✅ Location: `internal/tools/recommend.go:476-589`

### ❌ Missing Wisdom Integration

1. **Scorecard Action** (`handleReport` → `scorecard`)
   - ❌ No wisdom integration
   - Current output: metrics, health checks, recommendations only
   - Location: `internal/tools/handlers.go:205-240`, `internal/tools/scorecard_go.go:609-677`

2. **Overview Action** (`handleReportOverview`)
   - ❌ No wisdom integration
   - Current output: project data, health metrics, tasks
   - Location: `internal/tools/report.go:16-64`

3. **PRD Action** (`handleReportPRD`)
   - ⚠️ Less appropriate (formal document, but could have closing wisdom)

## Wisdom Integration Strategy

### 1. Scorecard Integration (HIGH PRIORITY)

**Location:** `FormatGoScorecard` function  
**File:** `internal/tools/scorecard_go.go`

#### Approach Options

**Option A: Simple Quote Section (Recommended for MVP)**
- Add wisdom section after recommendations
- Single quote based on overall score
- Simple, clean, matches PROJECT_SCORECARD_WITH_WISDOM.md example

**Option B: Multi-Advisor Consultation (Enhanced)**
- Get advisor consultations for specific problem areas:
  - Testing advisor if test coverage < 80%
  - Security advisor if security score < 70%
  - Quality advisor if lint/build issues
- General quote based on overall score
- More contextual but more complex

**Option C: Score-Based Guidance (Full Feature)**
- Multiple quotes from different sources (similar to briefing)
- Advisor consultations for each failing area
- Consultation mode guidance based on score
- Most comprehensive but most complex

#### Recommended: Option A (MVP) + Option B (Future Enhancement)

**Implementation Steps:**
1. Create `FormatGoScorecardWithWisdom` function
2. Add wisdom section after recommendations
3. Use overall score to get appropriate quote
4. Select source based on score range (or use "random")
5. Add optional advisor consultations for key problem areas (future)

**Function Signature:**
```go
func FormatGoScorecardWithWisdom(scorecard *GoScorecardResult, includeWisdom bool) string
```

**Format:**
```
======================================================================
  📊 GO PROJECT SCORECARD
======================================================================

  [Existing scorecard content...]

  Recommendations:
    • [recommendations...]

  ──────────────────────────────────────────────────────────────────
  🧘 Wisdom for Your Journey
  ──────────────────────────────────────────────────────────────────

  > "[quote text]"
  > — [source name]

  Encouragement: [encouragement message]
```

**Score-Based Quote Selection:**
- **Low scores (0-40%)**: Motivational quotes (stoic, pistis_sophia)
- **Medium scores (40-70%)**: Balanced guidance (tao, confucius)
- **High scores (70-100%)**: Encouraging mastery quotes (art_of_war, kybalion)

### 2. Overview Integration (MEDIUM PRIORITY)

**Location:** `formatOverviewText`, `formatOverviewMarkdown` functions  
**File:** `internal/tools/report.go:528-659`

#### Approach

**Option A: Closing Wisdom Quote (Recommended)**
- Add wisdom section at the end of overview
- Single quote based on overall health score
- Lightweight addition

**Option B: Section-Specific Wisdom (Enhanced)**
- Wisdom quote in health section
- Advisor consultation for risks section
- More contextual but potentially cluttered

#### Recommended: Option A

**Implementation:**
- Get overall health score from aggregated data
- Add wisdom section at end (before closing)
- Similar format to scorecard wisdom section

### 3. PRD Integration (LOW PRIORITY)

**Location:** `generatePRD` function  
**File:** `internal/tools/report.go:447-526`

**Recommendation:** Skip for now (PRDs are formal documents)

**Future Consideration:** Optional closing quote if requested via parameter

## Implementation Plan

### Phase 1: Scorecard Wisdom (MVP)

**Tasks:**
1. ✅ Create plan document (this file)
2. ⏳ Add `FormatGoScorecardWithWisdom` function
   - Signature: `func FormatGoScorecardWithWisdom(scorecard *GoScorecardResult) string`
   - Calls `FormatGoScorecard` for base content
   - Adds wisdom section using `getWisdomEngine()`
   - Gets quote based on `scorecard.Score`
3. ⏳ Update `handleReport` scorecard case
   - Call `FormatGoScorecardWithWisdom` instead of `FormatGoScorecard`
   - Make wisdom optional via parameter (default: true)
4. ⏳ Update `ScorecardOptions` struct
   - Add `IncludeWisdom bool` field (default: true)
5. ⏳ Add tests
   - Test wisdom integration
   - Test score-based quote selection
   - Test optional wisdom parameter

**Files to Modify:**
- `internal/tools/scorecard_go.go` - Add wisdom formatting function
- `internal/tools/handlers.go` - Update scorecard handler
- `internal/tools/scorecard_go.go` - Update ScorecardOptions
- `internal/tools/scorecard_go_test.go` - Add tests

**Dependencies:**
- ✅ Wisdom engine already available (`internal/tools/wisdom.go`)
- ✅ `getWisdomEngine()` function already exists
- ✅ Score available in `GoScorecardResult.Score`

### Phase 2: Scorecard Wisdom (Enhanced - Advisor Consultations)

**Tasks:**
1. ⏳ Identify problem areas from scorecard
   - Testing issues (low coverage, failing tests)
   - Security issues (vulnerabilities, missing checks)
   - Quality issues (lint errors, build failures)
2. ⏳ Get advisor consultations for problem areas
   - Use `GetAdvisorForMetric("testing")`, `GetAdvisorForMetric("security")`, etc.
   - Get quotes with appropriate scores
3. ⏳ Format advisor consultations in wisdom section
   - Show advisor icon, name, rationale
   - Include quote and encouragement
4. ⏳ Update formatting function
   - Add advisor consultations before general quote
   - Format similar to `PROJECT_SCORECARD_WITH_WISDOM.md`

**Files to Modify:**
- `internal/tools/scorecard_go.go` - Enhance wisdom formatting
- `internal/tools/scorecard_go_test.go` - Add advisor tests

### Phase 3: Overview Wisdom (Optional)

**Tasks:**
1. ⏳ Add wisdom parameter to overview handlers
2. ⏳ Get overall health score from aggregated data
3. ⏳ Add wisdom section to text/markdown formats
4. ⏳ Test wisdom integration in overview

**Files to Modify:**
- `internal/tools/report.go` - Add wisdom to overview formatting
- `internal/tools/handlers.go` - Add wisdom parameter
- `internal/tools/report_test.go` - Add tests

## Requirements

### ⚠️ CRITICAL: sources.json Configuration File

**The wisdom engine REQUIRES a `sources.json` configuration file to work properly.**

**File Locations (searched in order):**
1. Project root: `sources.json`, `.wisdom/sources.json`, `wisdom/sources.json`
2. Current directory: `sources.json`, `wisdom/sources.json`, `.wisdom/sources.json`
3. Home directory: `~/.wisdom/sources.json`, `~/.exarp_wisdom/sources.json`
4. XDG config: `$XDG_CONFIG_HOME/wisdom/sources.json` or `~/.config/wisdom/sources.json`

**If `sources.json` is missing:**
- The wisdom engine falls back to a minimal hard-coded source (BOFH with one quote)
- Wisdom section will still appear but with limited quotes
- The engine will still work, but with reduced functionality

**Recommended Setup:**
- Copy `sources.json` from `devwisdom-go` project root
- Place it in the exarp-go project root
- This provides quotes from multiple sources (stoic, tao, pistis_sophia, bible, etc.)

**Example:**
```bash
cp ../devwisdom-go/sources.json ./sources.json
```

**File Format:**
The `sources.json` file contains wisdom sources with quotes organized by aeon levels:
- `chaos` (0-30% score)
- `lower_aeons` (31-50% score)
- `middle_aeons` (51-70% score)
- `upper_aeons` (71-85% score)
- `treasury` (86-100% score)

See `devwisdom-go/sources.json` for the complete format and example sources.

## Technical Details

### Wisdom Engine Usage

**Getting Wisdom Engine:**
```go
engine, err := getWisdomEngine()
if err != nil {
    // Handle error gracefully (skip wisdom, don't fail scorecard)
    return FormatGoScorecard(scorecard)
}
```

**Getting Quote:**
```go
quote, err := engine.GetWisdom(score, "random") // or specific source
if err != nil {
    // Skip wisdom section if error
    return FormatGoScorecard(scorecard)
}
```

**Note:** If `sources.json` is missing, `getWisdomEngine()` will still succeed, but `GetWisdom()` may return limited quotes or fail if the source is not found.

**Getting Advisor Consultation:**
```go
advisors := engine.GetAdvisors()
advisorInfo, err := advisors.GetAdvisorForMetric("testing")
if err == nil {
    quote, err := engine.GetWisdom(testingScore, advisorInfo.Advisor)
    // Use quote and advisorInfo
}
```

### Score-Based Source Selection

**Strategy:**
- Use "random" for variety (date-seeded for consistency)
- OR select source based on score range:
  - 0-40%: "stoic" (resilience)
  - 40-70%: "tao" (balance)
  - 70-100%: "art_of_war" (mastery)

**Recommended:** Use "random" for simplicity (date-seeded ensures consistency)

### Error Handling

**Graceful Degradation:**
- If wisdom engine fails to initialize → Skip wisdom section, return base scorecard
- If quote retrieval fails → Skip wisdom section, return base scorecard
- Never fail scorecard generation due to wisdom errors
- Log errors to stderr (don't expose to user)

**Example:**
```go
func FormatGoScorecardWithWisdom(scorecard *GoScorecardResult) string {
    base := FormatGoScorecard(scorecard)
    
    engine, err := getWisdomEngine()
    if err != nil {
        return base // Skip wisdom, don't fail
    }
    
    quote, err := engine.GetWisdom(scorecard.Score, "random")
    if err != nil {
        return base // Skip wisdom, don't fail
    }
    
    // Append wisdom section
    var sb strings.Builder
    sb.WriteString(base)
    sb.WriteString("\n\n")
    sb.WriteString(formatWisdomSection(quote))
    return sb.String()
}
```

## Configuration

### Scorecard Options

**Add to `ScorecardOptions` struct:**
```go
type ScorecardOptions struct {
    FastMode     bool // Existing
    IncludeWisdom bool // New: default true
}
```

**Usage in handler:**
```go
opts := &ScorecardOptions{
    FastMode: true,
    IncludeWisdom: true, // Default, can be overridden
}
```

### Report Tool Schema

**Optional parameter (if needed):**
```go
"include_wisdom": map[string]interface{}{
    "type":    "boolean",
    "default": true,
}
```

**Note:** For MVP, wisdom can be always included (no parameter needed)

## Testing Strategy

### Unit Tests

1. **Test `FormatGoScorecardWithWisdom`**
   - Test with valid scorecard and wisdom engine
   - Test with wisdom engine error (should return base scorecard)
   - Test with quote error (should return base scorecard)
   - Test wisdom section formatting

2. **Test Score-Based Quote Selection**
   - Test low score quote selection
   - Test medium score quote selection
   - Test high score quote selection

3. **Test Advisor Consultations (Phase 2)**
   - Test advisor selection for problem areas
   - Test multiple advisor consultations
   - Test advisor formatting

### Integration Tests

1. **Test Scorecard Generation with Wisdom**
   - Generate scorecard for test project
   - Verify wisdom section is included
   - Verify quote is appropriate for score

2. **Test Error Handling**
   - Test with missing sources.json (graceful degradation)
   - Test with invalid scorecard data

## Example Output

### Current Scorecard (No Wisdom)
```
======================================================================
  📊 GO PROJECT SCORECARD
======================================================================

  OVERALL SCORE: 40.0%
  Production Ready: NO ❌

  [Metrics, Health Checks, Recommendations...]
```

### Enhanced Scorecard (With Wisdom)
```
======================================================================
  📊 GO PROJECT SCORECARD
======================================================================

  OVERALL SCORE: 40.0%
  Production Ready: NO ❌

  [Metrics, Health Checks, Recommendations...]

  ──────────────────────────────────────────────────────────────────
  🧘 Wisdom for Your Journey
  ──────────────────────────────────────────────────────────────────

  > "The best revenge is not to be like your enemy."
  > — Marcus Aurelius, Meditations

  Encouragement: Rise above. Don't let technical debt define your
  project—use it as motivation to improve. Every challenge is an
  opportunity for growth.

  [Optional: Advisor Consultations for specific problem areas]
```

## Benefits

1. **Actionable Guidance**: Wisdom quotes provide context and encouragement
2. **Score-Appropriate**: Quotes match project state (struggling vs. thriving)
3. **Consistent Format**: Matches existing briefing format
4. **Graceful Degradation**: Never fails due to wisdom errors
5. **Optional Enhancement**: Can be disabled if needed

## Risks & Mitigations

### Risks

1. **Performance**: Wisdom engine initialization adds overhead
   - **Mitigation**: Singleton pattern already in place, minimal overhead
2. **Error Handling**: Wisdom failures could break scorecard
   - **Mitigation**: Graceful degradation, never fail scorecard
3. **Output Length**: Wisdom section makes scorecards longer
   - **Mitigation**: Optional parameter, or keep it short
4. **User Preferences**: Some users may not want wisdom
   - **Mitigation**: Optional parameter (default: true)

### Mitigations Summary

- ✅ Graceful error handling
- ✅ Optional parameter (default: true)
- ✅ Singleton wisdom engine (performance)
- ✅ Short wisdom section (length)

## Success Criteria

### Phase 1 (MVP) - ✅ COMPLETE

- ✅ Scorecard includes wisdom section
- ✅ Quote is appropriate for score
- ✅ Graceful error handling
- ✅ `sources.json` configuration file added
- ✅ Documentation updated

### Phase 2 (Enhanced)

- ✅ Advisor consultations for problem areas
- ✅ Multiple advisor consultations formatted correctly
- ✅ Tests pass for advisors
- ✅ Documentation updated

### Phase 3 (Overview - Optional)

- ✅ Overview includes wisdom (optional)
- ✅ Tests pass
- ✅ Documentation updated

## Timeline

- **Phase 1 (MVP)**: 1-2 hours
  - Implement basic wisdom integration
  - Add tests
  - Documentation

- **Phase 2 (Enhanced)**: 1-2 hours
  - Add advisor consultations
  - Enhanced formatting
  - Tests

- **Phase 3 (Overview)**: 30-60 minutes
  - Add to overview
  - Tests
  - Documentation

**Total Estimated Time**: 3-5 hours

## References

- `PROJECT_SCORECARD_WITH_WISDOM.md` - Example implementation
- `internal/tools/report.go:66-129` - Briefing implementation (reference)
- `internal/tools/recommend.go:476-589` - Advisor consultation (reference)
- `internal/tools/wisdom.go` - Wisdom engine access
- `internal/tools/scorecard_go.go:609-677` - Current scorecard formatting
- `sources.json` - Wisdom sources configuration (required)
- `devwisdom-go/sources.json` - Source of wisdom configuration
- `devwisdom-go/docs/CONFIGURABLE_SOURCES.md` - Complete sources.json documentation

## Implementation Status

**Phase 1 (MVP): ✅ COMPLETE**
- Wisdom integration implemented in scorecards
- `sources.json` configuration file added
- Tested and working

## Next Steps

1. ✅ Phase 1 (MVP) - Complete
2. Consider Phase 2 (Enhanced) - Advisor consultations for problem areas
3. Consider Phase 3 (Overview) - Optional wisdom in overview reports
