package inputs

type ResolveQuestionInput struct {
	Status     string `json:"status" validate:"required"`
	Resolution string `json:"resolution" validate:"required"`
}
