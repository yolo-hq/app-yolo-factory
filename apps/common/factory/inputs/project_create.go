package inputs

// CreateProjectInput is the input for creating a project.
type CreateProjectInput struct {
	Name             string  `json:"name" validate:"required"`
	RepoURL          string  `json:"repoUrl" validate:"required"`
	LocalPath        string  `json:"localPath" validate:"required"`
	DefaultBranch    string  `json:"defaultBranch"`
	DefaultModel     string  `json:"defaultModel"`
	BudgetPerTaskUSD float64 `json:"budgetPerTaskUsd"`
	BudgetPerPrdUSD  float64 `json:"budgetPerPrdUsd"`
	BudgetMonthlyUSD float64 `json:"budgetMonthlyUsd"`
	MaxRetries       int     `json:"maxRetries"`
	TimeoutSecs      int     `json:"timeoutSecs"`
	AutoMerge        *bool   `json:"autoMerge"`
	SetupCommands    string  `json:"setupCommands"`
	TestCommands     string  `json:"testCommands"`
}
