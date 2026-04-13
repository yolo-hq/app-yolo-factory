package inputs

type DiffPRDInput struct {
	PRDID string `json:"prdId" validate:"required"`
}
