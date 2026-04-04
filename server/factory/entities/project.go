package entities

import "github.com/yolo-hq/yolo/core/entity"

type Project struct {
	entity.BaseEntity
	Name                 string  `json:"name" bun:"name,notnull,unique" fake:"name"`
	Status               string  `json:"status" bun:"status,notnull,default:'active'" fake:"oneof:active,archived,paused"`
	RepoURL              string  `json:"repoUrl" bun:"repo_url,notnull" fake:"url"`
	LocalPath            string  `json:"localPath" bun:"local_path,notnull,default:''" fake:"-"`
	UseWorktrees         bool    `json:"useWorktrees" bun:"use_worktrees,default:false" fake:"bool"`
	DefaultBranch        string  `json:"defaultBranch" bun:"default_branch,notnull,default:'main'" fake:"oneof:main,develop,staging"`
	MaintenanceBranches  string  `json:"maintenanceBranches" bun:"maintenance_branches,default:'[]'" fake:"-"`
	DefaultModel         string  `json:"defaultModel" bun:"default_model,notnull,default:'sonnet'" fake:"oneof:sonnet,opus,haiku"`
	EscalationModel      string  `json:"escalationModel" bun:"escalation_model,notnull,default:'opus'" fake:"oneof:opus,sonnet"`
	EscalationAfterRetries int   `json:"escalationAfterRetries" bun:"escalation_after_retries,notnull,default:2" fake:"int:1,5"`
	BudgetPerTaskUSD     float64 `json:"budgetPerTaskUsd" bun:"budget_per_task_usd,notnull,default:2.0" fake:"float:1,10"`
	BudgetPerPrdUSD      float64 `json:"budgetPerPrdUsd" bun:"budget_per_prd_usd,notnull,default:20.0" fake:"float:5,50"`
	BudgetMonthlyUSD     float64 `json:"budgetMonthlyUsd" bun:"budget_monthly_usd,notnull,default:200.0" fake:"float:50,500"`
	BudgetWarningAt      float64 `json:"budgetWarningAt" bun:"budget_warning_at,notnull,default:0.8" fake:"float:0.5,0.9"`
	SpentThisMonthUSD    float64 `json:"spentThisMonthUsd" bun:"spent_this_month_usd,notnull,default:0" fake:"float:0,100"`
	MaxRetries           int     `json:"maxRetries" bun:"max_retries,notnull,default:3" fake:"int:1,5"`
	TimeoutSecs          int     `json:"timeoutSecs" bun:"timeout_secs,notnull,default:600" fake:"int:60,1800"`
	AutoMerge            bool    `json:"autoMerge" bun:"auto_merge,notnull,default:true" fake:"bool"`
	AutoStart            bool    `json:"autoStart" bun:"auto_start,notnull,default:false" fake:"bool"`
	PushFailedBranches   bool    `json:"pushFailedBranches" bun:"push_failed_branches,notnull,default:false" fake:"bool"`
	SetupCommands        string  `json:"setupCommands" bun:"setup_commands,default:'[]'" fake:"-"`
	TestCommands         string  `json:"testCommands" bun:"test_commands,default:'[\"go build ./...\",\"go test ./...\"]'" fake:"-"`

	// Relations
	PRDs        []PRD        `json:"prds,omitempty" bun:"-" yolo:"rel:has_many,fk:project_id"`
	Tasks       []Task       `json:"tasks,omitempty" bun:"-" yolo:"rel:has_many,fk:project_id"`
	Suggestions []Suggestion `json:"suggestions,omitempty" bun:"-" yolo:"rel:has_many,fk:project_id"`
}

func (Project) TableName() string  { return "factory_projects" }
func (Project) EntityName() string { return "Project" }
