package entities

import (
	"time"

	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/entity"
)

type Step struct {
	bun.BaseModel `bun:"table:factory_steps"`
	entity.BaseEntity
	RunID         string     `json:"run_id" bun:"run_id,notnull" fake:"rel:Run"`
	Phase         string     `json:"phase" bun:"phase,notnull" fake:"oneof:setup,execute,verify,review" enum:"plan,implement,test,lint,audit,review"`
	Skill         string     `json:"skill" bun:"skill,notnull" fake:"oneof:code,test,review,fix"`
	Status        string     `json:"status" bun:"status,notnull,default:'running'" fake:"oneof:running,completed,failed" enum:"running,completed,failed,skipped"`
	ResultStatus  string     `json:"result_status" bun:"result_status" fake:"oneof:done,done_with_concerns,needs_context,blocked,failed" enum:"done,done_with_concerns,needs_context,blocked,failed"`
	Model         string     `json:"model" bun:"model,notnull" fake:"oneof:sonnet,opus,haiku"`
	SessionID     string     `json:"session_id" bun:"session_id" fake:"-"`
	CostUSD       float64    `json:"cost_usd" bun:"cost_usd,default:0" fake:"float:0,10"`
	TokensIn      int        `json:"tokens_in" bun:"tokens_in,default:0" fake:"int:0,50000"`
	TokensOut     int        `json:"tokens_out" bun:"tokens_out,default:0" fake:"int:0,25000"`
	Turns         int        `json:"turns" bun:"turns,default:0" fake:"int:0,80"`
	DurationMs    int        `json:"duration_ms" bun:"duration_ms,default:0" fake:"int:500,60000"`
	InputSummary  string     `json:"input_summary" bun:"input_summary" fake:"sentence:10"`
	OutputSummary string     `json:"output_summary" bun:"output_summary" fake:"sentence:10"`
	StartedAt     time.Time  `json:"started_at" bun:"started_at,notnull,default:current_timestamp"`
	CompletedAt   *time.Time `json:"completed_at" bun:"completed_at" fake:"-"`

	// Relations
	Run *Run `json:"run,omitempty" bun:"-" yolo:"rel:belongs_to,fk:run_id"`
}

func (Step) TableName() string  { return "factory_steps" }
func (Step) EntityName() string { return "Step" }
