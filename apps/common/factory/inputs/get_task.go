package inputs

// GetTaskInput is the input for fetching a single task by ID.
type GetTaskInput struct {
	ID string `json:"id" validate:"required"`
}
