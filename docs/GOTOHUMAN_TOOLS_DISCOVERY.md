# gotoHuman Tools Discovery

**Date:** 2026-01-07  
**Status:** 🔍 Discovery Phase  
**Server:** `gotohuman` MCP Server

## Discovery Steps

### Step 1: List gotoHuman Tools

In Cursor chat, try:
```
@gotohuman list tools
```

Or check Cursor's MCP interface:
- Settings → Features → Model Context Protocol
- Click on `gotohuman` server
- View available tools

### Step 2: Test Basic Tool

Try calling a gotoHuman tool:
```
@gotohuman [tool-name] [arguments]
```

### Step 3: Document Tools

Once tools are discovered, document:
- Tool names
- Tool descriptions
- Required parameters
- Return values
- Example usage

## Expected Tools (Based on Description)

Based on the description "Human-in-the-loop platform - Allow AI agents and automations to send requests for approval to your gotoHuman inbox", we expect:

1. **Request Approval Tool**
   - Send approval request to human inbox
   - Parameters: title, description, context, metadata
   - Returns: request ID

2. **Check Approval Status Tool**
   - Check status of approval request
   - Parameters: request ID
   - Returns: status (pending/approved/rejected), decision, feedback

3. **List Pending Approvals Tool**
   - List all pending approval requests
   - Returns: list of pending requests

4. **Cancel Approval Tool**
   - Cancel a pending approval request
   - Parameters: request ID

## Integration Points

Once tools are discovered, we can:

1. **Enhance task_workflow tool** to use gotoHuman
2. **Auto-send approval requests** when tasks move to Review
3. **Check approval status** and update Todo2 tasks
4. **Handle approval responses** automatically

---

**Next:** Discover tools and document findings below.

