package skills

// PlanTasksSchema is the JSON schema for plan-tasks structured output.
const PlanTasksSchema = `{
  "type": "object",
  "required": ["tasks"],
  "properties": {
    "tasks": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["title", "spec", "acceptance_criteria", "sequence"],
        "properties": {
          "title": {"type": "string"},
          "spec": {"type": "string"},
          "acceptance_criteria": {
            "type": "array",
            "items": {
              "type": "object",
              "required": ["id", "description"],
              "properties": {
                "id": {"type": "string"},
                "description": {"type": "string"}
              }
            }
          },
          "sequence": {"type": "integer"},
          "depends_on": {
            "type": "array",
            "items": {"type": "integer"}
          }
        }
      }
    }
  }
}`

// ReviewTaskSchema is the JSON schema for task review structured output.
const ReviewTaskSchema = `{
  "type": "object",
  "required": ["verdict", "criteria_results"],
  "properties": {
    "verdict": {"type": "string", "enum": ["pass", "fail"]},
    "criteria_results": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["criteria_id", "passed", "comment"],
        "properties": {
          "criteria_id": {"type": "string"},
          "passed": {"type": "boolean"},
          "evidence": {"type": "string"},
          "file": {"type": "string"},
          "line": {"type": "integer"},
          "comment": {"type": "string"}
        }
      }
    },
    "anti_patterns": {
      "type": "array",
      "items": {"type": "string"}
    },
    "reasons": {
      "type": "array",
      "items": {"type": "string"}
    },
    "suggestions": {
      "type": "array",
      "items": {"type": "string"}
    }
  }
}`

// AuditSchema is the JSON schema for audit structured output.
const AuditSchema = `{
  "type": "object",
  "required": ["passed"],
  "properties": {
    "passed": {"type": "boolean"},
    "violations": {
      "type": "array",
      "items": {"type": "string"}
    },
    "warnings": {
      "type": "array",
      "items": {"type": "string"}
    }
  }
}`

// SentinelSchema is the JSON schema for sentinel structured output.
const SentinelSchema = `{
  "type": "object",
  "required": ["findings"],
  "properties": {
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["watch", "status", "description"],
        "properties": {
          "watch": {"type": "string"},
          "status": {"type": "string", "enum": ["ok", "warning", "critical"]},
          "description": {"type": "string"},
          "suggested_action": {"type": "string"}
        }
      }
    }
  }
}`

// IntegrationReviewSchema is the JSON schema for integration review structured output.
const IntegrationReviewSchema = `{
  "type": "object",
  "required": ["findings"],
  "properties": {
    "findings": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["category", "severity", "message"],
        "properties": {
          "category": {"type": "string"},
          "severity": {"type": "string", "enum": ["error", "warning", "info"]},
          "message": {"type": "string"},
          "files": {"type": "array", "items": {"type": "string"}}
        }
      }
    }
  }
}`

// ProcessAdvisorSchema is the JSON schema for process advisor structured output.
const ProcessAdvisorSchema = `{
  "type": "object",
  "properties": {
    "insights": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "title": {"type": "string"},
          "body": {"type": "string"},
          "recommendation": {"type": "string"},
          "category": {"type": "string"},
          "priority": {"type": "string", "enum": ["low", "medium", "high"]}
        },
        "required": ["title", "body", "recommendation", "category", "priority"]
      }
    }
  },
  "required": ["insights"]
}`

// AdvisorSchema is the JSON schema for advisor structured output.
const AdvisorSchema = `{
  "type": "object",
  "required": ["suggestions"],
  "properties": {
    "suggestions": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["title", "body", "category", "priority"],
        "properties": {
          "title": {"type": "string"},
          "body": {"type": "string"},
          "category": {"type": "string", "enum": ["optimization", "refactoring", "tech_debt", "new_feature", "pattern_extraction"]},
          "priority": {"type": "string", "enum": ["low", "medium", "high"]},
          "estimated_impact": {"type": "string"}
        }
      }
    }
  }
}`
