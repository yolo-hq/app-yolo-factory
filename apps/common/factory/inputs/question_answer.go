package inputs

// AnswerQuestionInput is the input for answering an open question.
type AnswerQuestionInput struct {
	Answer string `json:"answer" validate:"required"`
}
