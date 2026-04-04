#!/bin/bash
# Mock Claude CLI that returns a successful implementation result.
cat <<'EOF'
{
  "result": "Implementation complete. Created user entity, actions, and migrations. All tests pass.",
  "session_id": "mock-impl-session-001",
  "total_cost_usd": 1.20,
  "is_error": false,
  "stop_reason": "end_turn",
  "num_turns": 12,
  "duration_ms": 45000,
  "duration_api_ms": 38000,
  "usage": {"input_tokens": 15000, "output_tokens": 8000}
}
EOF
