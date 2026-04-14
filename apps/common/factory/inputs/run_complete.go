package inputs

type CompleteRunInput struct {
	Status       string  `json:"status" validate:"required"`
	CostUSD      float64 `json:"cost_usd"`
	TokensIn     int     `json:"tokens_in"`
	TokensOut    int     `json:"tokens_out"`
	DurationMS   int     `json:"duration_ms"`
	NumTurns     int     `json:"num_turns"`
	Error        string  `json:"error"`
	CommitHash   string  `json:"commit_hash"`
	FilesChanged string  `json:"files_changed"`
	Result       string  `json:"result"`
	SessionID    string  `json:"session_id"`
}
