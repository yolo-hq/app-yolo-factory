package skills

// AdvisorTemplate is the prompt template for optimization advisor analysis.
const AdvisorTemplate = `You are an optimization advisor analyzing project {{.ProjectName}}.

## Analysis Type: {{.AnalysisType}}

## Context
{{.AnalysisContext}}

## Run History
{{.RunHistory}}

## Instructions
Analyze the project and suggest improvements.

Categories: optimization, refactoring, tech_debt, new_feature, pattern_extraction

For each suggestion:
- Title (short, actionable)
- Body (what to do and why)
- Priority (low/medium/high)
- Estimated impact

Output as structured JSON.`
