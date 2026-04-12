package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type PRDSubmit struct {
	command.Base
}

type PRDSubmitInput struct {
	Project  string `flag:"project" validate:"required" usage:"Project ID or name"`
	Title    string `flag:"title" validate:"required" usage:"PRD title"`
	Body     string `flag:"body" usage:"PRD body text"`
	File     string `flag:"file" usage:"Read body from file"`
	Criteria string `flag:"criteria" usage:"Acceptance criteria (comma-separated)"`
}

func (c *PRDSubmit) Name() string        { return "prd:submit" }
func (c *PRDSubmit) Description() string { return "Submit a new PRD" }
func (c *PRDSubmit) Input() any          { return &PRDSubmitInput{} }

func (c *PRDSubmit) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*PRDSubmitInput)

	projectRepo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get project repo: %w", err)
	}
	pr := projectRepo.(entity.ReadRepository[entities.Project])

	project, err := findProjectByIDOrName(ctx, pr, input.Project)
	if err != nil {
		return err
	}

	body := input.Body
	if input.File != "" {
		data, err := os.ReadFile(input.File)
		if err != nil {
			return fmt.Errorf("read file %s: %w", input.File, err)
		}
		body = string(data)
	}
	if body == "" {
		return fmt.Errorf("either --body or --file is required")
	}

	prdRepo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get prd repo: %w", err)
	}
	w := prdRepo.(entity.WriteRepository[entities.PRD])

	prd := &entities.PRD{
		ProjectID:          project.ID,
		Title:              input.Title,
		Body:               body,
		AcceptanceCriteria: input.Criteria,
		Status:             string(enums.PRDStatusDraft),
		Source:             string(enums.PRDSourceManual),
		CreatedBy:          constants.ActorHuman,
	}

	created, err := w.Insert(ctx, prd)
	if err != nil {
		return fmt.Errorf("insert prd: %w", err)
	}

	cctx.Print("Submitted PRD %s: %s", created.ID, created.Title)
	return nil
}
