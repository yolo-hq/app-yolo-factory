package inputs

type AnswerQuestionInput struct {
	Answer string `json:"answer" validate:"required"`
}
