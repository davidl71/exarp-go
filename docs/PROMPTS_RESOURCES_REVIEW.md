# Prompts and Resources Review

**Date:** 2026-01-07  
**Reviewer:** AI Assistant  
**Scope:** Review of all prompts and resources in exarp-go MCP server

## Executive Summary

✅ **Current Status:** 17 prompts and 6 resources are registered and functional  
⚠️ **Critical Gap:** Template variable substitution is **NOT implemented** despite MCP protocol support  
🎯 **Improvement Opportunities:** Multiple areas for enhancement identified

---

## 1. Prompts Review

### 1.1 Current Implementation

**Location:** `internal/prompts/registry.go`

**Registered Prompts (17 total):**
- 8 Original prompts: `align`, `discover`, `config`, `scan`, `scorecard`, `overview`, `dashboard`, `remember`
- 7 Workflow prompts: `daily_checkin`, `sprint_start`, `sprint_end`, `pre_sprint`, `post_impl`, `sync`, `dups`
- 2 mcp-generic-tools prompts: `context`, `mode`

**Current Flow:**
```go
// internal/prompts/registry.go:53-63
func createPromptHandler(promptName string) framework.PromptHandler {
    return func(ctx context.Context, args map[string]interface{}) (string, error) {
        // Retrieve prompt template from Python bridge
        promptText, err := bridge.GetPythonPrompt(ctx, promptName)
        if err != nil {
            return "", fmt.Errorf("failed to get prompt %s: %w", promptName, err)
        }
        return promptText, nil  // ⚠️ args are IGNORED!
    }
}
```

### 1.2 Issues Identified

#### ❌ **CRITICAL: Template Variables Not Used**

**Problem:** The MCP protocol provides `args map[string]interface{}` to prompts, but the current implementation:
1. Receives arguments from MCP client
2. **Completely ignores them**
3. Returns static prompt template without substitution

**Impact:**
- Prompts cannot be customized with user-provided values
- No dynamic content injection (e.g., task IDs, project names, dates)
- Limited reusability and flexibility

**Example of What's Missing:**
```go
// Current (broken):
promptText, err := bridge.GetPythonPrompt(ctx, promptName)
return promptText, nil  // Static text only

// Should be:
promptText, err := bridge.GetPythonPrompt(ctx, promptName)
// Substitute variables: {task_id} -> args["task_id"]
substituted := substituteTemplate(promptText, args)
return substituted, nil
```

#### ⚠️ **Error Handling**

**Current Issues:**
- Generic error messages don't indicate which prompt failed
- No validation of required vs optional arguments
- No helpful error messages for missing required variables

**Improvement:**
```go
// Better error handling
if err != nil {
    return "", fmt.Errorf("prompt '%s' retrieval failed: %w", promptName, err)
}

// Validate required args
if required, ok := getRequiredArgs(promptName); ok {
    for _, req := range required {
        if _, exists := args[req]; !exists {
            return "", fmt.Errorf("prompt '%s' requires argument: %s", promptName, req)
        }
    }
}
```

#### ⚠️ **Input Validation**

**Missing:**
- No validation of argument types
- No validation of argument values (e.g., date formats, IDs)
- No sanitization of user input

#### ⚠️ **Documentation**

**Missing:**
- No schema definition for prompt arguments
- No documentation of which prompts accept arguments
- No examples of argument usage

### 1.3 Template Support Analysis

**MCP Protocol Support:** ✅ **YES**
- MCP protocol supports `Arguments` in `GetPromptRequest`
- Arguments are passed as `map[string]any` to handlers
- Framework correctly extracts and passes arguments

**Current Implementation:** ❌ **NO**
- Arguments received but not used
- No template substitution logic
- No variable placeholder format defined

**Recommended Template Format:**
- **Option 1:** `{variable_name}` (simple, Python `.format()` compatible)
- **Option 2:** `{{variable_name}}` (mustache-style, clearer for complex templates)
- **Option 3:** `$variable_name` (shell-style, familiar)

**Recommendation:** Use `{variable_name}` for Python compatibility and simplicity.

### 1.4 Improvement Recommendations

#### 🔧 **High Priority**

1. **Implement Template Substitution**
   ```go
   func substituteTemplate(template string, args map[string]interface{}) string {
       result := template
       for key, value := range args {
           placeholder := "{" + key + "}"
           result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
       }
       return result
   }
   ```

2. **Add Argument Schema Definition**
   ```go
   type PromptSchema struct {
       Name        string
       Description string
       Arguments   []ArgumentDefinition
   }
   
   type ArgumentDefinition struct {
       Name        string
       Type        string  // "string", "number", "boolean"
       Required    bool
       Description string
       Default     interface{}
   }
   ```

3. **Validate Required Arguments**
   - Check for required args before template substitution
   - Return clear error messages for missing args

#### 🔧 **Medium Priority**

4. **Enhanced Error Messages**
   - Include prompt name in all errors
   - List available arguments when validation fails
   - Suggest correct argument names for typos

5. **Argument Type Validation**
   - Validate argument types match expected schema
   - Convert types when safe (e.g., string to number)

6. **Documentation**
   - Add argument schemas to prompt descriptions
   - Create examples showing argument usage
   - Document template variable syntax

#### 🔧 **Low Priority**

7. **Template Validation**
   - Validate template syntax on registration
   - Check for undefined variables in templates
   - Warn about unused arguments

8. **Caching**
   - Cache prompt templates (they're static)
   - Only re-fetch on errors or explicit refresh

---

## 2. Resources Review

### 2.1 Current Implementation

**Location:** `internal/resources/handlers.go`

**Registered Resources (6 total):**
- `stdio://scorecard` - Project scorecard
- `stdio://memories` - All memories
- `stdio://memories/category/{category}` - Memories by category
- `stdio://memories/task/{task_id}` - Memories for task
- `stdio://memories/recent` - Recent memories
- `stdio://memories/session/{date}` - Session memories

**Current Flow:**
```go
// internal/resources/handlers.go:86-108
func handleScorecard(ctx context.Context, uri string) ([]byte, string, error) {
    return bridge.ExecutePythonResource(ctx, uri)
}

func handleMemoriesByCategory(ctx context.Context, uri string) ([]byte, string, error) {
    return bridge.ExecutePythonResource(ctx, uri)
}
// ... all handlers just pass URI to Python bridge
```

### 2.2 Issues Identified

#### ⚠️ **URI Template Parsing**

**Current:** URI templates like `stdio://memories/category/{category}` are registered, but:
- No validation that URI matches template pattern
- No extraction of template variables from actual URI
- Python bridge receives full URI and must parse it

**Example:**
```go
// Registered: "stdio://memories/category/{category}"
// Actual URI: "stdio://memories/category/debug"
// Current: Passes full URI to Python
// Better: Extract {category} = "debug" and pass as structured data
```

#### ⚠️ **Error Handling**

**Issues:**
- Generic error messages don't indicate which resource failed
- No validation of URI format before passing to Python
- No helpful errors for malformed URIs

**Improvement:**
```go
func handleMemoriesByCategory(ctx context.Context, uri string) ([]byte, string, error) {
    // Validate URI format
    if !strings.HasPrefix(uri, "stdio://memories/category/") {
        return nil, "", fmt.Errorf("invalid URI format: expected stdio://memories/category/{category}")
    }
    
    // Extract category
    category := strings.TrimPrefix(uri, "stdio://memories/category/")
    if category == "" {
        return nil, "", fmt.Errorf("category cannot be empty")
    }
    
    return bridge.ExecutePythonResource(ctx, uri)
}
```

#### ⚠️ **Input Validation**

**Missing:**
- No validation of URI parameters (e.g., date format, task ID format)
- No sanitization of extracted values
- No type checking

**Example:**
```go
// stdio://memories/session/{date} - should validate YYYY-MM-DD format
func handleSessionMemories(ctx context.Context, uri string) ([]byte, string, error) {
    date := extractDateFromURI(uri)
    if !isValidDateFormat(date) {
        return nil, "", fmt.Errorf("invalid date format: expected YYYY-MM-DD, got %s", date)
    }
    // ...
}
```

#### ⚠️ **Resource Metadata**

**Missing:**
- No resource schema definition
- No documentation of URI patterns
- No examples of valid URIs

### 2.3 Template Support Analysis

**MCP Protocol Support:** ✅ **YES** (via URI templates)
- Resources can use URI templates with `{variable}` placeholders
- MCP clients can resolve templates with actual values
- Framework supports template registration

**Current Implementation:** ⚠️ **PARTIAL**
- URI templates are registered correctly
- But template variables are not extracted or validated
- Python bridge must parse URI manually

**Recommendation:** Extract template variables in Go before passing to Python.

### 2.4 Improvement Recommendations

#### 🔧 **High Priority**

1. **URI Template Variable Extraction**
   ```go
   func extractTemplateVars(template, uri string) (map[string]string, error) {
       // Parse template: "stdio://memories/category/{category}"
       // Parse URI: "stdio://memories/category/debug"
       // Return: map[string]string{"category": "debug"}
   }
   ```

2. **URI Validation**
   - Validate URI matches registered template pattern
   - Extract and validate template variables
   - Return clear errors for invalid URIs

3. **Parameter Validation**
   - Validate date formats (YYYY-MM-DD)
   - Validate task IDs (format, existence)
   - Validate category names (enum)

#### 🔧 **Medium Priority**

4. **Resource Schema Definition**
   ```go
   type ResourceSchema struct {
       URI         string
       Name        string
       Description string
       Template    string  // URI template
       Parameters  []ParameterDefinition
   }
   ```

5. **Enhanced Error Messages**
   - Include resource URI in errors
   - Show expected URI format
   - Suggest valid parameter values

6. **Documentation**
   - Document URI patterns
   - Provide examples of valid URIs
   - Document parameter requirements

#### 🔧 **Low Priority**

7. **Caching**
   - Cache resource responses when appropriate
   - Implement cache invalidation strategies

8. **Resource Discovery**
   - Add resource listing endpoint
   - Provide resource metadata via MCP protocol

---

## 3. Template Support Implementation Plan

### 3.1 Prompt Template Support

**Phase 1: Basic Substitution**
1. Implement `substituteTemplate()` function
2. Update `createPromptHandler()` to use substitution
3. Test with simple prompts

**Phase 2: Schema & Validation**
1. Define prompt argument schemas
2. Add argument validation
3. Improve error messages

**Phase 3: Documentation**
1. Document template syntax
2. Add argument schemas to descriptions
3. Create usage examples

### 3.2 Resource Template Support

**Phase 1: URI Parsing**
1. Implement `extractTemplateVars()` function
2. Update resource handlers to extract variables
3. Pass structured data to Python bridge

**Phase 2: Validation**
1. Add URI format validation
2. Validate extracted parameters
3. Improve error messages

**Phase 3: Documentation**
1. Document URI patterns
2. Provide URI examples
3. Document parameter requirements

---

## 4. Code Quality Improvements

### 4.1 Error Handling

**Current:**
```go
if err != nil {
    return "", fmt.Errorf("failed to get prompt %s: %w", promptName, err)
}
```

**Improved:**
```go
if err != nil {
    return "", fmt.Errorf("prompt '%s' retrieval failed: %w (hint: check Python bridge)", promptName, err)
}
```

### 4.2 Input Validation

**Add validation functions:**
```go
func validatePromptArgs(promptName string, args map[string]interface{}) error {
    schema := getPromptSchema(promptName)
    for _, param := range schema.Required {
        if _, exists := args[param.Name]; !exists {
            return fmt.Errorf("required argument missing: %s", param.Name)
        }
    }
    return nil
}
```

### 4.3 Logging

**Add structured logging:**
```go
logger.Info("prompt requested",
    "prompt", promptName,
    "args_count", len(args),
    "args", args,
)
```

---

## 5. Testing Recommendations

### 5.1 Unit Tests

- Test template substitution with various variable types
- Test error handling for missing required args
- Test URI template parsing and extraction
- Test parameter validation

### 5.2 Integration Tests

- Test prompt retrieval with arguments
- Test resource URI resolution
- Test error propagation from Python bridge

### 5.3 Edge Cases

- Empty arguments map
- Missing required arguments
- Invalid argument types
- Malformed URIs
- Special characters in variables

---

## 6. Summary

### Current State
- ✅ 17 prompts registered and functional
- ✅ 6 resources registered and functional
- ✅ MCP protocol support for arguments/templates
- ❌ Template substitution **NOT implemented**
- ⚠️ Limited error handling and validation

### Priority Improvements

1. **CRITICAL:** Implement prompt template variable substitution
2. **HIGH:** Add URI template variable extraction for resources
3. **HIGH:** Add argument/parameter validation
4. **MEDIUM:** Improve error messages and documentation
5. **LOW:** Add caching and performance optimizations

### Estimated Effort

- **Prompt Templates:** 4-6 hours (basic) to 8-12 hours (full with validation)
- **Resource Templates:** 2-4 hours (basic) to 6-8 hours (full with validation)
- **Documentation:** 2-4 hours
- **Testing:** 4-6 hours

**Total:** 12-22 hours for complete implementation

---

## 7. Next Steps

1. **Create implementation tasks** for template support
2. **Prioritize** based on user needs
3. **Implement** basic template substitution first
4. **Add validation** and error handling
5. **Document** template syntax and usage
6. **Test** thoroughly with various scenarios

---

**Review Complete** ✅

