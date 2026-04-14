package entities

import (
	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/entity"
)

type Project struct {
	bun.BaseModel `bun:"table:factory_projects"`
	entity.BaseEntity
	Name                   string  `json:"name" bun:"name,notnull,unique" fake:"name"`
	Status                 string  `json:"status" bun:"status,notnull,default:'active'" fake:"oneof:active,archived,paused" enum:"active,paused,archived"`
	RepoURL                string  `json:"repo_url" bun:"repo_url,notnull" fake:"url"`
	LocalPath              string  `json:"local_path" bun:"local_path,notnull,default:''" fake:"-"`
	UseWorktrees           bool    `json:"use_worktrees" bun:"use_worktrees,default:false" fake:"bool"`
	DefaultBranch          string  `json:"default_branch" bun:"default_branch,notnull,default:'main'" fake:"oneof:main,develop,staging"`
	MaintenanceBranches    string  `json:"maintenance_branches" bun:"maintenance_branches,default:'[]'" fake:"-"`
	DefaultModel           string  `json:"default_model" bun:"default_model,notnull,default:'sonnet'" fake:"oneof:sonnet,opus,haiku"`
	EscalationModel        string  `json:"escalation_model" bun:"escalation_model,notnull,default:'opus'" fake:"oneof:opus,sonnet"`
	EscalationAfterRetries int     `json:"escalation_after_retries" bun:"escalation_after_retries,notnull,default:2" fake:"int:1,5"`
	BudgetPerTaskUSD       float64 `json:"budget_per_task_usd" bun:"budget_per_task_usd,notnull,default:2.0" fake:"float:1,10"`
	BudgetPerPrdUSD        float64 `json:"budget_per_prd_usd" bun:"budget_per_prd_usd,notnull,default:20.0" fake:"float:5,50"`
	BudgetMonthlyUSD       float64 `json:"budget_monthly_usd" bun:"budget_monthly_usd,notnull,default:200.0" fake:"float:50,500"`
	BudgetWarningAt        float64 `json:"budget_warning_at" bun:"budget_warning_at,notnull,default:0.8" fake:"float:0.5,0.9"`
	SpentThisMonthUSD      float64 `json:"spent_this_month_usd" bun:"spent_this_month_usd,notnull,default:0" fake:"float:0,100"`
	MaxRetries             int     `json:"max_retries" bun:"max_retries,notnull,default:3" fake:"int:1,5"`
	TimeoutSecs            int     `json:"timeout_secs" bun:"timeout_secs,notnull,default:600" fake:"int:60,1800"`
	AutoMerge              bool    `json:"auto_merge" bun:"auto_merge,notnull,default:true" fake:"bool"`
	AutoStart              bool    `json:"auto_start" bun:"auto_start,notnull,default:false" fake:"bool"`
	PushFailedBranches     bool    `json:"push_failed_branches" bun:"push_failed_branches,notnull,default:false" fake:"bool"`
	SetupCommands          string  `json:"setup_commands" bun:"setup_commands,default:'[]'" fake:"-"`
	TestCommands           string  `json:"test_commands" bun:"test_commands,default:'[]'" fake:"-"`

	// Relations
	PRDs        []PRD        `json:"prds,omitempty" bun:"-" yolo:"rel:has_many,fk:project_id"`
	Tasks       []Task       `json:"tasks,omitempty" bun:"-" yolo:"rel:has_many,fk:project_id"`
	Suggestions []Suggestion `json:"suggestions,omitempty" bun:"-" yolo:"rel:has_many,fk:project_id"`
}

func (Project) TableName() string  { return "factory_projects" }
func (Project) EntityName() string { return "Project" }
