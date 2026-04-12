package inputs

// DiffPRDInput is the input for the PRD diff action.
type DiffPRDInput struct {
	PRDID string `json:"prdId" validate:"required"`
}
