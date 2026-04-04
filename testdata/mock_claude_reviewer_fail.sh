#!/bin/bash
# Mock Claude CLI that returns a failing review result.
cat <<'EOF'
{
  "result": "Review complete. Acceptance criteria not fully met.",
  "session_id": "mock-review-fail-001",
  "total_cost_usd": 0.28,
  "is_error": false,
  "stop_reason": "end_turn",
  "num_turns": 3,
  "duration_ms": 9000,
  "duration_api_ms": 7500,
  "usage": {"input_tokens": 8000, "output_tokens": 1200},
  "structured_output": {
    "verdict": "fail",
    "criteria_results": [
      {"id": "AC1", "passed": true, "reason": "Migration runs cleanly"},
      {"id": "AC2", "passed": false, "reason": "Missing role column in users table"}
    ],
    "anti_patterns": ["Hardcoded default role instead of using config"],
    "suggestions": ["Add role column", "Use configurable defaults"]
  }
}
EOF
