package inputs

type CreateProjectInput struct {
	Name             string  `json:"name" validate:"required"`
	RepoURL          string  `json:"repo_url" validate:"required"`
	LocalPath        string  `json:"local_path" validate:"required"`
	DefaultBranch    string  `json:"default_branch"`
	DefaultModel     string  `json:"default_model"`
	BudgetPerTaskUSD float64 `json:"budget_per_task_usd"`
	BudgetPerPrdUSD  float64 `json:"budget_per_prd_usd"`
	BudgetMonthlyUSD float64 `json:"budget_monthly_usd"`
	MaxRetries       int     `json:"max_retries"`
	TimeoutSecs      int     `json:"timeout_secs"`
	AutoMerge        *bool   `json:"auto_merge"`
	SetupCommands    string  `json:"setup_commands"`
	TestCommands     string  `json:"test_commands"`
}
