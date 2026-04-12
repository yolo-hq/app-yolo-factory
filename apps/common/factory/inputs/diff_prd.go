package inputs

// DiffPRDInput is the input for computing the combined git diff for a PRD.
type DiffPRDInput struct {
	PRDID string `json:"prdId" validate:"required"`
}
