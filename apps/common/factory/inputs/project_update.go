package inputs

type UpdateProjectInput struct {
	Name             *string  `json:"name"`
	Status           *string  `json:"status"`
	LocalPath        *string  `json:"localPath"`
	DefaultBranch    *string  `json:"defaultBranch"`
	DefaultModel     *string  `json:"defaultModel"`
	EscalationModel  *string  `json:"escalationModel"`
	BudgetPerTaskUSD *float64 `json:"budgetPerTaskUsd"`
	BudgetPerPrdUSD  *float64 `json:"budgetPerPrdUsd"`
	BudgetMonthlyUSD *float64 `json:"budgetMonthlyUsd"`
	MaxRetries       *int     `json:"maxRetries"`
	TimeoutSecs      *int     `json:"timeoutSecs"`
	AutoMerge        *bool    `json:"autoMerge"`
	AutoStart        *bool    `json:"autoStart"`
	SetupCommands    *string  `json:"setupCommands"`
	TestCommands     *string  `json:"testCommands"`
}
