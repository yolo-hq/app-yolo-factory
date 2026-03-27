package inputs

type UpdateRepoInput struct {
	Name          *string `json:"name"`
	URL           *string `json:"url"`
	LocalPath     *string `json:"localPath"`
	TargetBranch  *string `json:"targetBranch"`
	DefaultModel  *string `json:"defaultModel"`
	FeedbackLoops *string `json:"feedbackLoops"`
	Active        *bool   `json:"active"`
}
