package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// --- ProjectAdd ---

type ProjectAdd struct {
	command.Base
}

type ProjectAddInput struct {
	Name      string `flag:"name" validate:"required" usage:"Project name"`
	RepoURL   string `flag:"repo" validate:"required" usage:"Git repository URL"`
	LocalPath string `flag:"path" validate:"required" usage:"Local clone path"`
	Branch    string `flag:"branch" usage:"Default branch"`
	Model     string `flag:"model" usage:"Default model"`
}

func (c *ProjectAdd) Name() string        { return "project:add" }
func (c *ProjectAdd) Description() string { return "Add a new project" }
func (c *ProjectAdd) Input() any          { return &ProjectAddInput{} }

func (c *ProjectAdd) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*ProjectAddInput)
	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])

	p := &entities.Project{
		Name:          input.Name,
		RepoURL:       input.RepoURL,
		LocalPath:     input.LocalPath,
		DefaultBranch: input.Branch,
		DefaultModel:  input.Model,
		Status:        "active",
	}
	if p.DefaultBranch == "" {
		p.DefaultBranch = "main"
	}
	if p.DefaultModel == "" {
		p.DefaultModel = "sonnet"
	}

	created, err := w.Insert(ctx, p)
	if err != nil {
		return fmt.Errorf("insert project: %w", err)
	}

	cctx.Print("Created project %s (%s)", created.Name, created.ID)
	return nil
}

// --- ProjectList ---

type ProjectList struct {
	command.Base
}

func (c *ProjectList) Name() string        { return "project:list" }
func (c *ProjectList) Description() string { return "List all projects" }

func (c *ProjectList) Execute(ctx context.Context, cctx command.Context) error {
	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Project])

	result, err := r.FindMany(ctx, entity.FindOptions{
		Sort: &entity.SortParams{Field: "name", Order: "asc"},
	})
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	if len(result.Data) == 0 {
		cctx.Print("No projects found.")
		return nil
	}

	tw := cctx.Table()
	fmt.Fprintf(tw, "ID\tNAME\tSTATUS\tMODEL\tBUDGET\tSPENT\n")
	for _, p := range result.Data {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t$%.2f\t$%.2f\n",
			p.ID, p.Name, p.Status, p.DefaultModel, p.BudgetMonthlyUSD, p.SpentThisMonthUSD)
	}
	tw.Flush()
	return nil
}

// --- ProjectGet ---

type ProjectGet struct {
	command.Base
}

func (c *ProjectGet) Name() string        { return "project:get" }
func (c *ProjectGet) Description() string { return "Get a project by ID or name" }

func (c *ProjectGet) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: project:get <id-or-name>")
	}
	idOrName := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Project])

	p, err := findProjectByIDOrName(ctx, r, idOrName)
	if err != nil {
		return err
	}

	cctx.Print("ID:       %s", p.ID)
	cctx.Print("Name:     %s", p.Name)
	cctx.Print("Status:   %s", p.Status)
	cctx.Print("Repo:     %s", p.RepoURL)
	cctx.Print("Path:     %s", p.LocalPath)
	cctx.Print("Branch:   %s", p.DefaultBranch)
	cctx.Print("Model:    %s", p.DefaultModel)
	cctx.Print("Budget:   $%.2f/month", p.BudgetMonthlyUSD)
	cctx.Print("Spent:    $%.2f", p.SpentThisMonthUSD)
	return nil
}

// --- ProjectUpdate ---

type ProjectUpdate struct {
	command.Base
}

type ProjectUpdateInput struct {
	Name    string  `flag:"name" usage:"Project name"`
	Branch  string  `flag:"branch" usage:"Default branch"`
	Model   string  `flag:"model" usage:"Default model"`
	Budget  float64 `flag:"budget" usage:"Monthly budget USD"`
	Retries int     `flag:"retries" usage:"Max retries"`
}

func (c *ProjectUpdate) Name() string        { return "project:update" }
func (c *ProjectUpdate) Description() string { return "Update a project" }
func (c *ProjectUpdate) Input() any          { return &ProjectUpdateInput{} }

func (c *ProjectUpdate) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: project:update <id> [flags]")
	}
	id := cctx.Args[0]
	input, _ := cctx.TypedInput.(*ProjectUpdateInput)

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])

	ub := w.Update(ctx).WhereID(id)
	if input.Name != "" {
		ub = ub.Set("name", input.Name)
	}
	if input.Branch != "" {
		ub = ub.Set("default_branch", input.Branch)
	}
	if input.Model != "" {
		ub = ub.Set("default_model", input.Model)
	}
	if input.Budget > 0 {
		ub = ub.Set("budget_monthly_usd", input.Budget)
	}
	if input.Retries > 0 {
		ub = ub.Set("max_retries", input.Retries)
	}

	if _, err := ub.Exec(ctx); err != nil {
		return fmt.Errorf("update project: %w", err)
	}

	cctx.Print("Updated project %s", id)
	return nil
}

// --- ProjectPause ---

type ProjectPause struct {
	command.Base
}

func (c *ProjectPause) Name() string        { return "project:pause" }
func (c *ProjectPause) Description() string { return "Pause a project" }

func (c *ProjectPause) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: project:pause <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])

	if _, err := w.Update(ctx).WhereID(id).Set("status", "paused").Exec(ctx); err != nil {
		return fmt.Errorf("pause project: %w", err)
	}

	cctx.Print("Paused project %s", id)
	return nil
}

// --- ProjectResume ---

type ProjectResume struct {
	command.Base
}

func (c *ProjectResume) Name() string        { return "project:resume" }
func (c *ProjectResume) Description() string { return "Resume a paused project" }

func (c *ProjectResume) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: project:resume <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])

	if _, err := w.Update(ctx).WhereID(id).Set("status", "active").Exec(ctx); err != nil {
		return fmt.Errorf("resume project: %w", err)
	}

	cctx.Print("Resumed project %s", id)
	return nil
}

// --- ProjectArchive ---

type ProjectArchive struct {
	command.Base
}

func (c *ProjectArchive) Name() string        { return "project:archive" }
func (c *ProjectArchive) Description() string { return "Archive a project" }

func (c *ProjectArchive) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: project:archive <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])

	if _, err := w.Update(ctx).WhereID(id).Set("status", "archived").Exec(ctx); err != nil {
		return fmt.Errorf("archive project: %w", err)
	}

	cctx.Print("Archived project %s", id)
	return nil
}

// findProjectByIDOrName tries ID first, then falls back to name filter.
func findProjectByIDOrName(ctx context.Context, r entity.ReadRepository[entities.Project], idOrName string) (*entities.Project, error) {
	// Try by ID first.
	p, err := r.FindOne(ctx, entity.FindOneOptions{ID: idOrName})
	if err != nil {
		return nil, fmt.Errorf("find project: %w", err)
	}
	if p != nil {
		return p, nil
	}

	// Try by name.
	p, err = r.FindOne(ctx, entity.FindOneOptions{
		Filters: []entity.FilterCondition{
			{Field: "name", Operator: entity.OpEq, Value: idOrName},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("find project by name: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("project %q not found", idOrName)
	}
	return p, nil
}
