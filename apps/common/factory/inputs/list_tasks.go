package inputs

// ListTasksInput is the input for listing tasks with optional filters.
type ListTasksInput struct {
	PRDID     string `json:"prdId"`
	ProjectID string `json:"projectId"`
	Status    string `json:"status"`
}
