package constants

// InsightCategory identifies the category of an insight.
type InsightCategory string

const (
	InsightRetryRate            InsightCategory = "retry_rate"
	InsightCostOptimization     InsightCategory = "cost_optimization"
	InsightModelSelection       InsightCategory = "model_selection"
	InsightSpecQuality          InsightCategory = "spec_quality"
	InsightGateEffectiveness    InsightCategory = "gate_effectiveness"
	InsightWorkflowOptimization InsightCategory = "workflow_optimization"
)
