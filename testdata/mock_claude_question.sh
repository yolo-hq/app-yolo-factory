#!/bin/bash
# Mock Claude CLI that returns a result containing a question for the human.
cat <<'EOF'
{
  "result": "I encountered an ambiguity in the spec. QUESTION: Should the user entity support soft deletes or hard deletes? The PRD mentions 'delete user' but doesn't specify the deletion strategy.",
  "session_id": "mock-question-session-001",
  "total_cost_usd": 0.50,
  "is_error": false,
  "stop_reason": "end_turn",
  "num_turns": 4,
  "duration_ms": 12000,
  "duration_api_ms": 10000,
  "usage": {"input_tokens": 10000, "output_tokens": 3000}
}
EOF
