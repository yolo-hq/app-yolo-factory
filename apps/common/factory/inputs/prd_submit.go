package inputs

type SubmitPRDInput struct {
	ProjectID          string `json:"project_id" validate:"required" resolves:"Project"`
	Title              string `json:"title" validate:"required"`
	Body               string `json:"body" validate:"required"`
	AcceptanceCriteria string `json:"acceptance_criteria" validate:"required"`
	DesignDecisions    string `json:"design_decisions"`
	Source             string `json:"source"`
}
