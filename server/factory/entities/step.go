package entities

import (
	"time"

	"github.com/yolo-hq/yolo/core/entity"
)

type Step struct {
	entity.BaseEntity
	RunID         string     `json:"runId" bun:"run_id,notnull" fake:"rel:Run"`
	Phase         string     `json:"phase" bun:"phase,notnull" fake:"oneof:setup,execute,verify,review"`
	Skill         string     `json:"skill" bun:"skill,notnull" fake:"oneof:code,test,review,fix"`
	Status        string     `json:"status" bun:"status,notnull,default:'running'" fake:"oneof:running,completed,failed"`
	Model         string     `json:"model" bun:"model,notnull" fake:"oneof:sonnet,opus,haiku"`
	SessionID     string     `json:"sessionId" bun:"session_id" fake:"-"`
	CostUSD       float64    `json:"costUsd" bun:"cost_usd,default:0" fake:"float:0,10"`
	TokensIn      int        `json:"tokensIn" bun:"tokens_in,default:0" fake:"int:0,50000"`
	TokensOut     int        `json:"tokensOut" bun:"tokens_out,default:0" fake:"int:0,25000"`
	DurationMs    int        `json:"durationMs" bun:"duration_ms,default:0" fake:"int:500,60000"`
	InputSummary  string     `json:"inputSummary" bun:"input_summary" fake:"sentence:10"`
	OutputSummary string     `json:"outputSummary" bun:"output_summary" fake:"sentence:10"`
	StartedAt     time.Time  `json:"startedAt" bun:"started_at,notnull,default:current_timestamp"`
	CompletedAt   *time.Time `json:"completedAt" bun:"completed_at" fake:"-"`

	// Relations
	Run *Run `json:"run,omitempty" bun:"-" yolo:"rel:belongs_to,fk:run_id"`
}

func (Step) TableName() string  { return "factory_steps" }
func (Step) EntityName() string { return "Step" }
