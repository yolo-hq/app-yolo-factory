package constants

// ReviewTaskTemplate is the prompt template for reviewing a task implementation.
const ReviewTaskTemplate = `You are reviewing code changes against acceptance criteria.

## Task
Title: {{.TaskTitle}}

{{.TaskSpec}}

## Acceptance Criteria
{{.AcceptanceCriteria}}

## Changes Made
{{.GitDiff}}

## Anti-Pattern Checklist
- Hardcoded values that should be configurable
- Missing error handling at system boundaries
- Tests that mock internals instead of real implementations
- Scope creep — changes beyond what the spec asked for
- Swallowed errors (_, _ = pattern)
- Duplicate functions that should be shared helpers
- String literal status values instead of constants

## Instructions
For EACH acceptance criterion:
1. State the criterion
2. Find the specific code or test that satisfies it
3. Quote the evidence: file path and line number
4. Verdict: PASS or FAIL

If you cannot find concrete evidence in the diff for a criterion, it FAILS.
Do not assume or infer — show proof.

Output as structured JSON.`
