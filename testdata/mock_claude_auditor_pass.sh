#!/bin/bash
# Mock Claude CLI that returns a passing audit result.
cat <<'EOF'
{
  "result": "Audit passed. Code follows all YOLO conventions.",
  "session_id": "mock-audit-pass-001",
  "total_cost_usd": 0.15,
  "is_error": false,
  "stop_reason": "end_turn",
  "num_turns": 2,
  "duration_ms": 5000,
  "duration_api_ms": 4000,
  "usage": {"input_tokens": 6000, "output_tokens": 500},
  "structured_output": {
    "passed": true,
    "violations": [],
    "suggestions": []
  }
}
EOF
