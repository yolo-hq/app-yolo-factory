package inputs

// GetProjectInput is the input for fetching a single project by ID.
type GetProjectInput struct {
	ID string `json:"id" validate:"required"`
}
