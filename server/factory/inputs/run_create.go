package inputs

type CreateRunInput struct {
	TaskID string `json:"taskId" validate:"required"`
	RepoID string `json:"repoId" validate:"required"`
	Agent  string `json:"agent"`
	Model  string `json:"model" validate:"required"`
}
