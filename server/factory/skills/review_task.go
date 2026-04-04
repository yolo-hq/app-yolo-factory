package skills

// ReviewTaskTemplate is the prompt template for reviewing a task implementation.
const ReviewTaskTemplate = `You are reviewing a code implementation against its acceptance criteria.

## Task
Title: {{.TaskTitle}}
Spec: {{.TaskSpec}}

## Acceptance Criteria
{{.AcceptanceCriteria}}

## Changes Made
{{.GitDiff}}

## Anti-Pattern Checklist
Check for:
- Hardcoded values that should be configurable
- Missing error handling at system boundaries
- Tests that mock internal code instead of using real implementations
- Code that violates YOLO entity/action patterns
- Scope creep — changes beyond what the task spec asked for
- Missing or incorrect type annotations
- Untested edge cases mentioned in the spec

## Instructions
Review the changes against each acceptance criterion. For each criterion, state whether it passes and why.

If you find anti-patterns or issues, list them.

Output your verdict as structured JSON: pass or fail with detailed reasons.`
