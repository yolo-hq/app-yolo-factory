package inputs

// ListQuestionsInput is the input for listing questions with optional filters.
type ListQuestionsInput struct {
	Status string `json:"status"`
}
