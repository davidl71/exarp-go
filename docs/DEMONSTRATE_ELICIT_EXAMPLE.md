# demonstrate_elicit Tool - Example Output

## Interactive Flow Simulation

This document shows what the `demonstrate_elicit` tool would return after the example interaction flow.

---

## Example Interaction Flow

```
AI: [Calls demonstrate_elicit]
AI: "Would you like to proceed with the demonstration?"
    [Question appears inline in chat]
You: "Yes"
AI: "Which demonstration pattern would you like to see?"
    Options: [Simple confirmation] [Multiple choice] [Text input] [All of the above]
You: "All of the above"
AI: "Enter a short message to demonstrate text input:"
You: "Hello, this is inline!"
AI: "Enter a number between 1 and 10:"
You: 7
AI: [Shows results with all responses]
```

---

## Expected JSON Result

After completing the above flow, the tool would return:

```json
{
  "demonstration": "FastMCP Elicitation API",
  "patterns_demonstrated": [
    "confirmation",
    "choice",
    "text_input",
    "number_input"
  ],
  "results": {
    "confirmation": {
      "question": "Would you like to proceed with the demonstration?",
      "user_response": true,
      "note": "This question appeared inline in chat, not as a pop-up"
    },
    "choice": {
      "question": "Which demonstration pattern would you like to see?",
      "options": [
        "Simple confirmation",
        "Multiple choice selection",
        "Text input with validation",
        "All of the above"
      ],
      "user_selection": "All of the above",
      "note": "User selected from options displayed inline in chat"
    },
    "text_input": {
      "question": "Enter a short message to demonstrate text input:",
      "user_response": "Hello, this is inline!",
      "action": "accept",
      "note": "User entered text inline in chat"
    },
    "number_input": {
      "question": "Enter a number between 1 and 10:",
      "user_response": 7,
      "validated": true,
      "note": "Number input with type validation"
    }
  },
  "summary": {
    "total_patterns": 4,
    "patterns": [
      "confirmation",
      "choice",
      "text_input",
      "number_input"
    ],
    "all_responses_received": true
  },
  "success": true
}
```

---

## Pattern Breakdown

### 1. Confirmation Pattern
- **Question**: "Would you like to proceed with the demonstration?"
- **User Response**: `true` (Yes)
- **API Used**: `elicit_confirmation(ctx, message)`
- **Returns**: `bool`

### 2. Choice Pattern
- **Question**: "Which demonstration pattern would you like to see?"
- **Options**: 4 choices displayed inline
- **User Selection**: "All of the above"
- **API Used**: `elicit_choice(ctx, message, choices)`
- **Returns**: Selected choice string

### 3. Text Input Pattern
- **Question**: "Enter a short message to demonstrate text input:"
- **User Response**: "Hello, this is inline!"
- **API Used**: `elicit(ctx, message, response_type=str)`
- **Returns**: `ElicitResult` with `action` and `data`

### 4. Number Input Pattern
- **Question**: "Enter a number between 1 and 10:"
- **User Response**: `7`
- **Validation**: ✅ Valid (1 ≤ 7 ≤ 10)
- **API Used**: `elicit(ctx, message, response_type=int)`
- **Returns**: `ElicitResult` with validated number

---

## Key Features Demonstrated

✅ **Inline Questions**: All questions appear in chat, not as pop-ups  
✅ **Type Validation**: Number input validates range (1-10)  
✅ **Multiple Patterns**: Shows 4 different interaction patterns  
✅ **Context Preservation**: All Q&A stays in chat history  
✅ **Natural Flow**: Feels like a conversation  

---

## Current Limitation

**Note**: This tool requires FastMCP Context, which is not available in stdio mode. When called in stdio mode, it returns:

```json
{
  "success": false,
  "error": "demonstrate_elicit requires FastMCP Context (not available in stdio mode)",
  "note": "This tool uses FastMCP's elicit() API for inline chat questions. Use FastMCP mode to access it.",
  "alternative": "Use interactive-mcp tools for pop-up questions in stdio mode"
}
```

---

## When to Use

- **Learning**: Understand how FastMCP's elicit() API works
- **Testing**: Verify elicit() functionality in FastMCP mode
- **Demonstration**: Show inline chat question capabilities
- **Development**: Reference implementation for your own tools

---

## Related Tools

- `interactive_task_create` - Practical example using elicit() for task creation
- See `docs/ELICIT_API_EXAMPLE.md` for full documentation

