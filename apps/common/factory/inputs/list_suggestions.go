package inputs

// ListSuggestionsInput is the input for listing suggestions with optional filters.
type ListSuggestionsInput struct {
	ProjectID string `json:"projectId"`
	Category  string `json:"category"`
	Priority  string `json:"priority"`
	Status    string `json:"status"`
}
