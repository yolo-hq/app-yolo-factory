package constants

// ReviewPRDTemplate is the prompt template for final PRD alignment review.
const ReviewPRDTemplate = `You are performing a final alignment review of a completed PRD.

## Original PRD
Title: {{.PRDTitle}}

{{.PRDBody}}

## PRD Acceptance Criteria
{{.AcceptanceCriteria}}

## Tasks Completed
{{.TaskSummaries}}

## Instructions
Compare what was requested in the PRD against what was actually delivered.

Check:
1. Are all acceptance criteria met?
2. Is there scope drift (things built that weren't asked for)?
3. Is there scope reduction (things asked for but not built)?
4. Do the tasks integrate correctly with each other?
5. Are there gaps between individual task completions and the overall PRD goal?

Output: alignment score (0.0-1.0), criteria met/missed, and recommendations.`
