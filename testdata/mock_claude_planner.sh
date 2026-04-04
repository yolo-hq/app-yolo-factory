#!/bin/bash
# Mock Claude CLI that returns a planner result with structured task list.
cat <<'EOF'
{
  "result": "I have analyzed the PRD and created 3 implementation tasks.",
  "session_id": "mock-planner-session-001",
  "total_cost_usd": 0.45,
  "is_error": false,
  "stop_reason": "end_turn",
  "num_turns": 5,
  "duration_ms": 15000,
  "duration_api_ms": 12000,
  "usage": {"input_tokens": 5000, "output_tokens": 2000},
  "structured_output": {
    "tasks": [
      {
        "title": "Setup database schema",
        "spec": "Create migration for users table with name, email, role columns",
        "acceptance_criteria": [
          {"id": "AC1", "description": "Migration runs without error"},
          {"id": "AC2", "description": "Table has correct columns"}
        ],
        "sequence": 1,
        "depends_on": [],
        "estimated_complexity": "low"
      },
      {
        "title": "Implement user CRUD",
        "spec": "Build create, read, update, delete actions for user entity",
        "acceptance_criteria": [
          {"id": "AC1", "description": "All CRUD endpoints work"}
        ],
        "sequence": 2,
        "depends_on": [1],
        "estimated_complexity": "medium"
      },
      {
        "title": "Add user tests",
        "spec": "Write integration tests for all user actions",
        "acceptance_criteria": [
          {"id": "AC1", "description": "Tests pass"},
          {"id": "AC2", "description": "Coverage above 80%"}
        ],
        "sequence": 3,
        "depends_on": [2],
        "estimated_complexity": "medium"
      }
    ]
  }
}
EOF
