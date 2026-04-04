#!/bin/bash
# Mock Claude CLI that returns a passing review result.
cat <<'EOF'
{
  "result": "Review complete. All acceptance criteria pass.",
  "session_id": "mock-review-pass-001",
  "total_cost_usd": 0.30,
  "is_error": false,
  "stop_reason": "end_turn",
  "num_turns": 3,
  "duration_ms": 10000,
  "duration_api_ms": 8000,
  "usage": {"input_tokens": 8000, "output_tokens": 1500},
  "structured_output": {
    "verdict": "pass",
    "criteria_results": [
      {"id": "AC1", "passed": true, "reason": "Migration runs cleanly"},
      {"id": "AC2", "passed": true, "reason": "All columns present with correct types"}
    ],
    "anti_patterns": [],
    "suggestions": ["Consider adding an index on email column"]
  }
}
EOF
