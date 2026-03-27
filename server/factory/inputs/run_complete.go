package inputs

type CompleteRunInput struct {
	RunID      string  `json:"runId" validate:"required"`
	Status     string  `json:"status" validate:"required"`
	Cost       float64 `json:"cost"`
	Duration   int     `json:"duration"`
	Error      string  `json:"error"`
	CommitHash string  `json:"commitHash"`
	LogURL     string  `json:"logUrl"`
}
