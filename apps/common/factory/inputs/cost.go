package inputs

// CostInput is the input for the cost breakdown query.
type CostInput struct {
	Period    string `json:"period"`
	ProjectID string `json:"projectId"`
}
