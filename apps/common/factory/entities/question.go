package entities

import (
	"time"

	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/entity"
)

type Question struct {
	bun.BaseModel `bun:"table:factory_questions"`
	entity.BaseEntity
	TaskID          string     `json:"taskId" bun:"task_id,notnull" fake:"rel:Task"`
	RunID           string     `json:"runId" bun:"run_id,notnull" fake:"rel:Run"`
	Body            string     `json:"body" bun:"body,notnull" fake:"sentence:20"`
	Context         string     `json:"context" bun:"context" fake:"sentence:15"`
	Confidence      string     `json:"confidence" bun:"confidence,notnull" fake:"oneof:low,medium,high" enum:"low,medium"`
	Status          string     `json:"status" bun:"status,notnull,default:'open'" fake:"oneof:open,answered,dismissed" enum:"open,answered,auto_resolved"`
	Answer          string     `json:"answer" bun:"answer" fake:"-"`
	AnsweredBy      string     `json:"answeredBy" bun:"answered_by" fake:"-"`
	AnswerSessionID string     `json:"answerSessionId" bun:"answer_session_id" fake:"-"`
	AnsweredAt      *time.Time `json:"answeredAt" bun:"answered_at" fake:"-"`

	// Relations
	Task *Task `json:"task,omitempty" bun:"-" yolo:"rel:belongs_to,fk:task_id"`
	Run  *Run  `json:"run,omitempty" bun:"-" yolo:"rel:belongs_to,fk:run_id"`
}

func (Question) TableName() string  { return "factory_questions" }
func (Question) EntityName() string { return "Question" }
