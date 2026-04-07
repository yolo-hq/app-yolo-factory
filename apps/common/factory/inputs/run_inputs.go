package inputs

// CompleteRunInput is the input for completing a run with metrics.
type CompleteRunInput struct {
	Status       string  `json:"status" validate:"required"`
	CostUSD      float64 `json:"costUsd"`
	TokensIn     int     `json:"tokensIn"`
	TokensOut    int     `json:"tokensOut"`
	DurationMS   int     `json:"durationMs"`
	NumTurns     int     `json:"numTurns"`
	Error        string  `json:"error"`
	CommitHash   string  `json:"commitHash"`
	FilesChanged string  `json:"filesChanged"`
	Result       string  `json:"result"`
	SessionID    string  `json:"sessionId"`
}
