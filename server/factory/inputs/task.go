package inputs

type CreateTaskInput struct {
	RepoID      string `json:"repoId" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Body        string `json:"body"`
	Type        string `json:"type"`
	Priority    int    `json:"priority"`
	Model       string `json:"model"`
	Labels      string `json:"labels"`
	ParentID    string `json:"parentId"`
	DependsOn   string `json:"dependsOn"`
	MaxRetries  int    `json:"maxRetries"`
	TimeoutSecs int    `json:"timeoutSecs"`
}

type UpdateTaskInput struct {
	Title       *string `json:"title"`
	Body        *string `json:"body"`
	Type        *string `json:"type"`
	Status      *string `json:"status"`
	Priority    *int    `json:"priority"`
	Model       *string `json:"model"`
	Labels      *string `json:"labels"`
	DependsOn   *string `json:"dependsOn"`
	MaxRetries  *int    `json:"maxRetries"`
	TimeoutSecs *int    `json:"timeoutSecs"`
}
