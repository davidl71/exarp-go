# gotoHuman MCP Tools Documentation

**Date:** 2026-01-07  
**Status:** ⚠️ **API Key Required**  
**Server:** `gotohuman` MCP Server

## ⚠️ Setup Requirement

**gotoHuman requires an API key for authentication.**

### Error Message
```
API key is required. Provide it in params or set the GOTOHUMAN_API_KEY environment variable.
```

### How to Get API Key

1. **Sign up for gotoHuman:**
   - Visit gotoHuman website/platform
   - Create an account
   - Generate an API key from your account settings

2. **Set API Key:**

   **Option A: Environment Variable (Recommended)**
   ```bash
   export GOTOHUMAN_API_KEY="your-api-key-here"
   ```
   
   **Option B: MCP Configuration**
   Update `.cursor/mcp.json` to include API key in environment:
   ```json
   {
     "mcpServers": {
       "gotohuman": {
         "command": "uvx",
         "args": [...],
         "env": {
           "GOTOHUMAN_API_KEY": "your-api-key-here"
         }
       }
     }
   }
   ```

   **Option C: Tool Parameters**
   - Provide API key when calling gotoHuman tools
   - Pass as parameter in tool calls

## Available Tools/Forms

Based on MCP server discovery, gotoHuman provides:

### 1. `list-forms`
- **Purpose:** List all available review forms
- **Requires:** API key
- **Returns:** List of form IDs and names

### 2. `get-form-schema`
- **Purpose:** Get schema for a specific form
- **Parameters:** `formId` (string)
- **Requires:** API key
- **Returns:** Form schema with field definitions

### 3. `request-human-review-with-form`
- **Purpose:** Request human approval with a form
- **Parameters:**
  - `formId` (string) - Form to use
  - `fieldData` (object) - Form field data
  - `metadata` (object, optional) - Additional metadata
  - `assignToUsers` (array, optional) - User emails to assign
- **Requires:** API key
- **Returns:** Review request ID

## Discovery Results

**Status:** ⚠️ **Blocked - API Key Required**

Once API key is configured, we can:
1. ✅ List all available forms
2. ✅ Get schema for each form
3. ✅ Test approval request workflow
4. ✅ Document integration patterns

## Next Steps

1. **Get gotoHuman API Key**
   - Sign up at gotoHuman platform
   - Generate API key from account settings

2. **Configure API Key**
   - Add to `.cursor/mcp.json` env section
   - Or set as environment variable

3. **Restart Cursor**
   - Restart to load API key configuration

4. **Discover Tools**
   - List forms: `@gotoHuman list-forms`
   - Get form schema: `@gotoHuman get-form-schema [formId]`
   - Test approval: `@gotoHuman request-human-review-with-form`

## Integration Plan

Once API key is configured:

### Phase 1: Basic Discovery
- [ ] List all available forms
- [ ] Get schema for each form
- [ ] Test basic approval request
- [ ] Document form fields and usage

### Phase 2: Todo2 Integration
- [ ] Create approval form mapping
- [ ] Enhance `task_workflow` tool
- [ ] Auto-send approval on Review state
- [ ] Handle approval responses

### Phase 3: Workflow Enhancement
- [ ] Approval status checking
- [ ] Approval timeout handling
- [ ] Approval history tracking
- [ ] Approval analytics

---

**Last Updated:** 2026-01-07  
**Status:** ⚠️ Waiting for API key configuration

