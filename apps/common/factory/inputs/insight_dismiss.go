package inputs

// DismissInsightInput is the input for dismissing an insight.
type DismissInsightInput struct {
	Reason string `json:"reason" validate:"required"`
}
