package entities

import (
	"time"

	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/entity"
)

type Run struct {
	bun.BaseModel `bun:"table:factory_runs"`
	entity.BaseEntity
	TaskID         string     `json:"task_id" bun:"task_id,notnull" fake:"rel:Task"`
	AgentType      string     `json:"agent_type" bun:"agent_type,notnull" fake:"oneof:claude-cli,claude-code,custom"`
	Status         string     `json:"status" bun:"status,notnull,default:'running'" fake:"oneof:running,completed,failed,cancelled" enum:"running,completed,failed,cancelled,blocked"`
	Model          string     `json:"model" bun:"model,notnull" fake:"oneof:sonnet,opus,haiku"`
	SessionID      string     `json:"session_id" bun:"session_id" fake:"-"`
	SessionName    string     `json:"session_name" bun:"session_name" fake:"-"`
	EscalatedModel string     `json:"escalated_model" bun:"escalated_model" fake:"-"`
	CostUSD        float64    `json:"cost_usd" bun:"cost_usd,default:0" fake:"float:0,25"`
	TokensIn       int        `json:"tokens_in" bun:"tokens_in,default:0" fake:"int:0,100000"`
	TokensOut      int        `json:"tokens_out" bun:"tokens_out,default:0" fake:"int:0,50000"`
	DurationMs     int        `json:"duration_ms" bun:"duration_ms,default:0" fake:"int:1000,300000"`
	NumTurns       int        `json:"num_turns" bun:"num_turns,default:0" fake:"int:1,50"`
	CommitHash     string     `json:"commit_hash" bun:"commit_hash" fake:"-"`
	BranchName     string     `json:"branch_name" bun:"branch_name" fake:"-"`
	FilesChanged   string     `json:"files_changed" bun:"files_changed,default:'[]'" fake:"-"`
	Result         string     `json:"result" bun:"result" fake:"-"`
	Error          string     `json:"error" bun:"error" fake:"-"`
	StartedAt      time.Time  `json:"started_at" bun:"started_at,notnull,default:current_timestamp"`
	CompletedAt    *time.Time `json:"completed_at" bun:"completed_at" fake:"-"`

	// Relations
	Task      *Task      `json:"task,omitempty" bun:"-" yolo:"rel:belongs_to,fk:task_id"`
	Steps     []Step     `json:"steps,omitempty" bun:"-" yolo:"rel:has_many,fk:run_id"`
	Reviews   []Review   `json:"reviews,omitempty" bun:"-" yolo:"rel:has_many,fk:run_id"`
	Questions []Question `json:"questions,omitempty" bun:"-" yolo:"rel:has_many,fk:run_id"`
}

func (Run) TableName() string  { return "factory_runs" }
func (Run) EntityName() string { return "Run" }
