package entities

// --- Project statuses ---

const (
	ProjectActive   = "active"
	ProjectPaused   = "paused"
	ProjectArchived = "archived"
)

// --- PRD statuses ---

const (
	PRDDraft      = "draft"
	PRDApproved   = "approved"
	PRDPlanning   = "planning"
	PRDInProgress = "in_progress"
	PRDCompleted  = "completed"
	PRDFailed     = "failed"
)

// --- Task statuses ---

const (
	TaskQueued    = "queued"
	TaskBlocked   = "blocked"
	TaskRunning   = "running"
	TaskReviewing = "reviewing"
	TaskDone      = "done"
	TaskFailed    = "failed"
	TaskCancelled = "cancelled"
)

// --- Run statuses ---

const (
	RunRunning   = "running"
	RunCompleted = "completed"
	RunFailed    = "failed"
	RunCancelled = "cancelled"
)

// --- Step statuses ---

const (
	StepRunning   = "running"
	StepCompleted = "completed"
	StepFailed    = "failed"
	StepSkipped   = "skipped"
)

// --- Review verdicts ---

const (
	ReviewPass = "pass"
	ReviewFail = "fail"
)

// --- Question statuses ---

const (
	QuestionOpen         = "open"
	QuestionAnswered     = "answered"
	QuestionAutoResolved = "auto_resolved"
)

// --- Question confidence levels ---

const (
	ConfidenceLow    = "low"
	ConfidenceMedium = "medium"
)

// --- Suggestion statuses ---

const (
	SuggestionPending   = "pending"
	SuggestionApproved  = "approved"
	SuggestionRejected  = "rejected"
	SuggestionConverted = "converted"
)

// --- Priority levels ---

const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// --- Agent types ---

const (
	AgentPlanner     = "planner"
	AgentImplementer = "implementer"
	AgentReviewer    = "reviewer"
	AgentAuditor     = "auditor"
	AgentSentinel    = "sentinel"
	AgentAdvisor     = "advisor"
)

// --- Step phases ---

const (
	PhasePlan      = "plan"
	PhaseImplement = "implement"
	PhaseTest      = "test"
	PhaseLint      = "lint"
	PhaseAudit     = "audit"
	PhaseReview    = "review"
)

// --- Suggestion categories ---

const (
	CategoryOptimization      = "optimization"
	CategoryRefactoring       = "refactoring"
	CategoryTechDebt          = "tech_debt"
	CategorySecurity          = "security"
	CategoryNewFeature        = "new_feature"
	CategoryPatternExtraction = "pattern_extraction"
	CategoryBugFix            = "bug_fix"
)

// --- Insight statuses ---

const (
	InsightPending      = "pending"
	InsightAcknowledged = "acknowledged"
	InsightApplied      = "applied"
	InsightDismissed    = "dismissed"
)

// --- Insight categories ---

const (
	InsightRetryRate            = "retry_rate"
	InsightCostOptimization     = "cost_optimization"
	InsightModelSelection       = "model_selection"
	InsightSpecQuality          = "spec_quality"
	InsightGateEffectiveness    = "gate_effectiveness"
	InsightWorkflowOptimization = "workflow_optimization"
)

// --- PRD sources ---

const (
	SourceManual           = "manual"
	SourceGrillMe          = "grill_me"
	SourceFactoryGenerated = "factory_generated"
	SourceImported         = "imported"
)
