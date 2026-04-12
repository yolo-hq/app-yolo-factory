package inputs

// ListInsightsInput is the input for listing insights with optional filters.
type ListInsightsInput struct {
	ProjectID string `json:"projectId"`
	Category  string `json:"category"`
	Priority  string `json:"priority"`
	Status    string `json:"status"`
}
