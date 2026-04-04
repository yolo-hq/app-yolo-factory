package inputs

// ApproveSuggestionInput is the input for approving a suggestion.
type ApproveSuggestionInput struct {
	PRDID string `json:"prdId"`
}

// RejectSuggestionInput is the input for rejecting a suggestion.
type RejectSuggestionInput struct {
	Reason string `json:"reason" validate:"required"`
}
