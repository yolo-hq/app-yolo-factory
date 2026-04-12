package inputs

// GetPRDInput is the input for fetching a single PRD by ID.
type GetPRDInput struct {
	ID string `json:"id" validate:"required"`
}
