package inputs

type RejectSuggestionInput struct {
	Reason string `json:"reason" validate:"required"`
}
