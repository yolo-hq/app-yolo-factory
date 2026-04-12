package inputs

// ListPRDsInput is the input for listing PRDs with optional filters.
type ListPRDsInput struct {
	ProjectID string `json:"projectId"`
	Status    string `json:"status"`
}
