package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// --- PRDSubmit ---

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
		Status:             entities.PRDDraft,
		Source:             entities.SourceManual,
		CreatedBy:          "human",
	}

	created, err := w.Insert(ctx, prd)
	if err != nil {
		return fmt.Errorf("insert prd: %w", err)
	}

	cctx.Print("Submitted PRD %s: %s", created.ID, created.Title)
	return nil
}

// --- PRDList ---

type PRDList struct {
	command.Base
}

type PRDListInput struct {
	Project string `flag:"project" usage:"Filter by project ID or name"`
	Status  string `flag:"status" usage:"Filter by status"`
}

func (c *PRDList) Name() string        { return "prd:list" }
func (c *PRDList) Description() string { return "List PRDs" }
func (c *PRDList) Input() any          { return &PRDListInput{} }

func (c *PRDList) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*PRDListInput)

	repo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.PRD])

	opts := entity.FindOptions{
		Sort: &entity.SortParams{Field: "created_at", Order: "desc"},
	}
	if input.Status != "" {
		opts.Filters = append(opts.Filters, entity.FilterCondition{
			Field: "status", Operator: entity.OpEq, Value: input.Status,
		})
	}
	if input.Project != "" {
		opts.Filters = append(opts.Filters, entity.FilterCondition{
			Field: "project_id", Operator: entity.OpEq, Value: input.Project,
		})
	}

	result, err := r.FindMany(ctx, opts)
	if err != nil {
		return fmt.Errorf("list prds: %w", err)
	}

	if len(result.Data) == 0 {
		cctx.Print("No PRDs found.")
		return nil
	}

	tw := cctx.Table()
	fmt.Fprintf(tw, "ID\tTITLE\tSTATUS\tTASKS\tCOST\n")
	for _, p := range result.Data {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%d/%d\t$%.2f\n",
			p.ID, p.Title, p.Status, p.CompletedTasks, p.TotalTasks, p.TotalCostUSD)
	}
	tw.Flush()
	return nil
}

// --- PRDGet ---

type PRDGet struct {
	command.Base
}

func (c *PRDGet) Name() string        { return "prd:get" }
func (c *PRDGet) Description() string { return "Get a PRD by ID" }

func (c *PRDGet) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: prd:get <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.PRD])

	p, err := r.FindOne(ctx, entity.FindOneOptions{ID: id})
	if err != nil {
		return fmt.Errorf("find prd: %w", err)
	}
	if p == nil {
		return fmt.Errorf("PRD %s not found", id)
	}

	cctx.Print("ID:       %s", p.ID)
	cctx.Print("Title:    %s", p.Title)
	cctx.Print("Status:   %s", p.Status)
	cctx.Print("Source:   %s", p.Source)
	cctx.Print("Tasks:    %d/%d completed", p.CompletedTasks, p.TotalTasks)
	cctx.Print("Cost:     $%.2f", p.TotalCostUSD)
	cctx.Print("")
	cctx.Print("Body:")
	cctx.Print("%s", p.Body)
	return nil
}

// --- PRDApprove ---

type PRDApprove struct {
	command.Base
}

func (c *PRDApprove) Name() string        { return "prd:approve" }
func (c *PRDApprove) Description() string { return "Approve a PRD" }

func (c *PRDApprove) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: prd:approve <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.PRD])

	if _, err := w.Update(ctx).WhereID(id).Set("status", entities.PRDApproved).Exec(ctx); err != nil {
		return fmt.Errorf("approve prd: %w", err)
	}

	cctx.Print("Approved PRD %s", id)
	return nil
}

// --- PRDExecute ---

type PRDExecute struct {
	command.Base
}

func (c *PRDExecute) Name() string        { return "prd:execute" }
func (c *PRDExecute) Description() string { return "Execute a PRD (triggers planning job)" }

func (c *PRDExecute) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: prd:execute <id>")
	}
	id := cctx.Args[0]

	// Verify PRD exists and is in a valid state.
	repo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.PRD])
	w := repo.(entity.WriteRepository[entities.PRD])

	prd, err := r.FindOne(ctx, entity.FindOneOptions{ID: id})
	if err != nil {
		return fmt.Errorf("find prd: %w", err)
	}
	if prd == nil {
		return fmt.Errorf("PRD %s not found", id)
	}

	if prd.Status != entities.PRDApproved && prd.Status != entities.PRDDraft {
		return fmt.Errorf("PRD must be in draft or approved status, got %s", prd.Status)
	}

	// Mark PRD as planning to trigger the PlanPRDJob via the worker.
	if _, err := w.Update(ctx).WhereID(id).Set("status", entities.PRDPlanning).Exec(ctx); err != nil {
		return fmt.Errorf("update prd: %w", err)
	}

	cctx.Print("PRD %s marked for planning. Worker will pick up the PlanPRDJob.", id)
	return nil
}
