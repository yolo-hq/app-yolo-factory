package inputs

type UpdateProjectInput struct {
	Name             *string  `json:"name"`
	Status           *string  `json:"status"`
	LocalPath        *string  `json:"local_path"`
	DefaultBranch    *string  `json:"default_branch"`
	DefaultModel     *string  `json:"default_model"`
	EscalationModel  *string  `json:"escalation_model"`
	BudgetPerTaskUSD *float64 `json:"budget_per_task_usd"`
	BudgetPerPrdUSD  *float64 `json:"budget_per_prd_usd"`
	BudgetMonthlyUSD *float64 `json:"budget_monthly_usd"`
	MaxRetries       *int     `json:"max_retries"`
	TimeoutSecs      *int     `json:"timeout_secs"`
	AutoMerge        *bool    `json:"auto_merge"`
	AutoStart        *bool    `json:"auto_start"`
	SetupCommands    *string  `json:"setup_commands"`
	TestCommands     *string  `json:"test_commands"`
}
