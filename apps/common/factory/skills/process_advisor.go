package skills

// ProcessAdvisorTemplate is the prompt template for process improvement analysis.
const ProcessAdvisorTemplate = `You are analyzing Factory execution history to suggest process improvements.

## Execution Metrics
{{.MetricsSummary}}

## Questions to Answer
1. Which task types fail most? Why? What would reduce retry rates?
2. Is the model selection optimal? Should some tasks use Opus instead of Sonnet?
3. Which quality gates (lint, audit, review) catch real issues vs waste time?
4. Are specs with more acceptance criteria more successful? What does this imply?
5. What is the biggest cost driver? How could it be reduced?
6. Are there workflow optimizations (skip redundant steps, reorder steps)?

## Instructions
For each insight:
- Title: short, actionable
- Body: what you found and why it matters
- Recommendation: specific action to take
- Category: retry_rate|cost_optimization|model_selection|spec_quality|gate_effectiveness|workflow_optimization
- Priority: low|medium|high

Only report insights backed by data. Do not speculate.
Output as structured JSON.`
