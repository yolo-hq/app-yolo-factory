package entities

import "github.com/yolo-hq/yolo/core/entity"

type LintResult struct {
	entity.BaseEntity
	RunID         string `json:"runId" bun:"run_id,notnull" fake:"rel:Run"`
	TaskID        string `json:"taskId" bun:"task_id,notnull" fake:"rel:Task"`
	Passed        bool   `json:"passed" bun:"passed,notnull,default:false" fake:"bool"`
	ChecksRun     int    `json:"checksRun" bun:"checks_run,notnull,default:0" fake:"int:1,20"`
	ChecksPassed  int    `json:"checksPassed" bun:"checks_passed,notnull,default:0" fake:"int:0,20"`
	ChecksFailed  int    `json:"checksFailed" bun:"checks_failed,notnull,default:0" fake:"int:0,10"`
	Findings      string `json:"findings" bun:"findings,notnull,default:'[]'" fake:"-"`

	// Relations
	Run  *Run  `json:"run,omitempty" bun:"-" yolo:"rel:belongs_to,fk:run_id"`
	Task *Task `json:"task,omitempty" bun:"-" yolo:"rel:belongs_to,fk:task_id"`
}

func (LintResult) TableName() string  { return "factory_lint_results" }
func (LintResult) EntityName() string { return "LintResult" }
