# Wisdom Integration Requirements

**Date:** 2026-01-13  
**Status:** ✅ Complete  
**Related:** `WISDOM_INTEGRATION_PLAN.md`, `sources.json`

## Overview

The wisdom integration in exarp-go requires a `sources.json` configuration file to provide contextual wisdom quotes in scorecards and reports.

## Requirements

### ⚠️ CRITICAL: sources.json Configuration File

**The wisdom engine REQUIRES a `sources.json` configuration file to work properly.**

Without this file, the wisdom engine falls back to a minimal hard-coded source (BOFH with one quote), which severely limits the wisdom functionality.

### File Location

The wisdom engine searches for `sources.json` in the following locations (in order of priority):

1. **Project root:**
   - `sources.json`
   - `.wisdom/sources.json`
   - `wisdom/sources.json`

2. **Current working directory:**
   - `sources.json`
   - `wisdom/sources.json`
   - `.wisdom/sources.json`

3. **Home directory:**
   - `~/.wisdom/sources.json`
   - `~/.exarp_wisdom/sources.json`

4. **XDG config directory:**
   - `$XDG_CONFIG_HOME/wisdom/sources.json` (if set)
   - `~/.config/wisdom/sources.json` (fallback)

**Recommendation:** Place `sources.json` in the project root for project-specific wisdom sources.

### File Format

The `sources.json` file contains wisdom sources with quotes organized by aeon levels:

```json
{
  "version": "1.0",
  "last_updated": "2025-01-26",
  "author": "devwisdom-go",
  "sources": {
    "stoic": {
      "id": "stoic",
      "name": "Stoic Philosophers",
      "icon": "🏛️",
      "description": "Marcus Aurelius, Epictetus, and Seneca on resilience and focus",
      "quotes": {
        "chaos": [
          {
            "quote": "The impediment to action advances action...",
            "source": "Marcus Aurelius, Meditations",
            "encouragement": "Obstacles are opportunities."
          }
        ],
        "lower_aeons": [...],
        "middle_aeons": [...],
        "upper_aeons": [...],
        "treasury": [...]
      }
    }
  }
}
```

### Aeon Levels (Score-Based)

Quotes are selected based on the project score:

- **chaos** (0-30%): Quotes for struggling projects
- **lower_aeons** (31-50%): Quotes for projects needing improvement
- **middle_aeons** (51-70%): Quotes for projects making progress
- **upper_aeons** (71-85%): Quotes for strong projects
- **treasury** (86-100%): Quotes for excellent projects

### Setup Instructions

**Option 1: Copy from devwisdom-go (Recommended)**

```bash
# From exarp-go project root
cp ../devwisdom-go/sources.json ./sources.json
```

**Option 2: Create from scratch**

1. Copy the example from `devwisdom-go/examples/sources.json`
2. Customize sources and quotes as needed
3. Place in project root

**Option 3: Use global configuration**

1. Copy `sources.json` to `~/.wisdom/sources.json`
2. All projects will use the same wisdom sources

### Verification

To verify the wisdom engine is working:

1. **Check file exists:**
   ```bash
   test -f sources.json && echo "✅ sources.json exists" || echo "❌ Missing"
   ```

2. **Generate a scorecard:**
   ```bash
   # Use the report tool
   # The scorecard should include a wisdom section at the end
   ```

3. **Check wisdom section:**
   - Scorecard should include "🧘 Wisdom for Your Journey" section
   - Should contain a quote, source, and encouragement
   - Quote should be appropriate for the project score

### Fallback Behavior

If `sources.json` is missing:

- ✅ Wisdom engine still initializes (no error)
- ✅ Falls back to minimal hard-coded source (BOFH)
- ✅ Wisdom section still appears but with limited quotes
- ✅ Scorecard generation never fails due to missing sources

This graceful degradation ensures scorecards always work, even without proper configuration.

### Source Configuration

The `sources.json` file can include multiple wisdom sources:

- **stoic** - Stoic philosophers (Marcus Aurelius, Epictetus, Seneca)
- **tao** - Tao Te Ching wisdom
- **pistis_sophia** - Gnostic wisdom texts
- **bible** - Biblical wisdom (Proverbs, Ecclesiastes)
- **confucius** - Confucian wisdom
- **art_of_war** - Strategic thinking (Sun Tzu)
- **bofh** - Tech humor and sysadmin wisdom

Each source should have quotes for all aeon levels for best coverage.

### Documentation

For complete documentation on `sources.json` format:

- `devwisdom-go/docs/CONFIGURABLE_SOURCES.md` - Complete format documentation
- `devwisdom-go/examples/sources.json` - Example configuration
- `devwisdom-go/sources.json` - Full working example

### Troubleshooting

**Problem: Wisdom section not appearing**

- ✅ Check `sources.json` exists in project root
- ✅ Verify file is valid JSON (`cat sources.json | jq .`)
- ✅ Check wisdom engine initialization (should not fail silently)
- ✅ Verify scorecard handler is using `FormatGoScorecardWithWisdom`

**Problem: Only one quote appearing (BOFH fallback)**

- ✅ `sources.json` is missing or invalid
- ✅ File exists but contains no valid sources
- ✅ JSON syntax error in file

**Problem: Wrong quotes for score**

- ✅ Check aeon level mapping (chaos=0-30%, etc.)
- ✅ Verify source has quotes for the relevant aeon level
- ✅ Check quote selection logic

### Integration Details

The wisdom engine is accessed via:

```go
// Get wisdom engine (singleton)
engine, err := getWisdomEngine()
if err != nil {
    // Graceful degradation
    return baseScorecard
}

// Get quote based on score
quote, err := engine.GetWisdom(scorecard.Score, "random")
if err != nil {
    // Graceful degradation
    return baseScorecard
}
```

The engine automatically loads `sources.json` from the configured locations during initialization.

### Best Practices

1. **Version Control:**
   - ✅ Commit `sources.json` to repository
   - ✅ Allows team to share wisdom sources
   - ✅ Ensures consistent wisdom across environments

2. **Customization:**
   - ✅ Can customize sources per project
   - ✅ Can add project-specific wisdom
   - ✅ Can override global sources

3. **Updates:**
   - ✅ Can update sources without code changes
   - ✅ Wisdom engine supports hot-reloading (if implemented)
   - ✅ Easy to add new sources or quotes

### Related Files

- `sources.json` - Wisdom sources configuration
- `internal/tools/wisdom.go` - Wisdom engine access
- `internal/tools/scorecard_go.go` - Wisdom integration in scorecards
- `docs/WISDOM_INTEGRATION_PLAN.md` - Implementation plan

### Status

**✅ Requirement Met:**
- `sources.json` exists in project root
- Wisdom integration working in scorecards
- Graceful degradation implemented
- Documentation complete
