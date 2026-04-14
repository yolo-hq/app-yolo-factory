package inputs

type DiffPRDInput struct {
	PRDID string `json:"prd_id" validate:"required"`
}
