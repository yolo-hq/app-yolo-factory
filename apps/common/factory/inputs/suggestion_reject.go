package inputs

// RejectSuggestionInput is the input for rejecting a suggestion.
type RejectSuggestionInput struct {
	Reason string `json:"reason" validate:"required"`
}
