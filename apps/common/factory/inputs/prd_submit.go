package inputs

type SubmitPRDInput struct {
	ProjectID          string `json:"projectId" validate:"required" resolves:"Project"`
	Title              string `json:"title" validate:"required"`
	Body               string `json:"body" validate:"required"`
	AcceptanceCriteria string `json:"acceptanceCriteria" validate:"required"`
	DesignDecisions    string `json:"designDecisions"`
	Source             string `json:"source"`
}
