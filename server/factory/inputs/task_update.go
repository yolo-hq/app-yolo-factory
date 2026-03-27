package inputs

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
