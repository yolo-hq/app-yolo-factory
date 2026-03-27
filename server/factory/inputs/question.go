package inputs

type CreateQuestionInput struct {
	TaskID  string `json:"taskId" validate:"required"`
	RunID   string `json:"runId" validate:"required"`
	RepoID  string `json:"repoId" validate:"required"`
	Context string `json:"context"`
	Tried   string `json:"tried"`
	Body    string `json:"body" validate:"required"`
}

type ResolveQuestionInput struct {
	Status     string `json:"status" validate:"required"`
	Resolution string `json:"resolution" validate:"required"`
}
