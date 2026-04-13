package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/read"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// --- ProjectScan ---

type ProjectScan struct {
	command.Base
}

type ProjectScanInput struct {
	Dir    string `flag:"dir" validate:"required" usage:"Directory to scan for git repos"`
	DryRun bool   `flag:"dry-run" usage:"Show repos without registering"`
	Branch string `flag:"branch" usage:"Default branch (default: main)"`
	Model  string `flag:"model" usage:"Default model (default: sonnet)"`
}

func (c *ProjectScan) Name() string { return "project:scan" }
func (c *ProjectScan) Description() string {
	return "Auto-discover and register git repos in a directory"
}
func (c *ProjectScan) Input() any { return &ProjectScanInput{} }

type scannedRepo struct {
	Name      string
	Path      string
	RemoteURL string
	Status    string // "new" or "existing"
}

func (c *ProjectScan) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*ProjectScanInput)

	branch := input.Branch
	if branch == "" {
		branch = "main"
	}
	model := input.Model
	if model == "" {
		model = "sonnet"
	}

	// Walk Dir, find directories with both .git/ AND go.mod.
	entries, err := os.ReadDir(input.Dir)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", input.Dir, err)
	}

	var repos []scannedRepo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirPath := filepath.Join(input.Dir, e.Name())

		// Check for .git/ and go.mod.
		gitDir := filepath.Join(dirPath, ".git")
		goMod := filepath.Join(dirPath, "go.mod")

		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			continue
		}
		if _, err := os.Stat(goMod); os.IsNotExist(err) {
			continue
		}

		absPath, err := filepath.Abs(dirPath)
		if err != nil {
			absPath = dirPath
		}

		// Get remote URL (best-effort).
		remoteURL := ""
		out, err := exec.CommandContext(ctx, "git", "-C", absPath, "remote", "get-url", "origin").Output()
		if err == nil {
			remoteURL = strings.TrimSpace(string(out))
		}

		repos = append(repos, scannedRepo{
			Name:      e.Name(),
			Path:      absPath,
			RemoteURL: remoteURL,
		})
	}

	if len(repos) == 0 {
		cctx.Print("No git+go.mod repos found in %s", input.Dir)
		return nil
	}

	// Check existing projects.
	existing, err := read.FindMany[entities.Project](ctx)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	existingNames := make(map[string]bool)
	existingPaths := make(map[string]bool)
	for _, p := range existing {
		existingNames[p.Name] = true
		existingPaths[p.LocalPath] = true
	}

	for i := range repos {
		if existingNames[repos[i].Name] || existingPaths[repos[i].Path] {
			repos[i].Status = "existing"
		} else {
			repos[i].Status = "new"
		}
	}

	// Print table.
	tw := cctx.Table()
	fmt.Fprintf(tw, "NAME\tPATH\tREMOTE\tSTATUS\n")
	for _, r := range repos {
		remote := r.RemoteURL
		if remote == "" {
			remote = "(none)"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.Name, r.Path, remote, r.Status)
	}
	tw.Flush()

	if input.DryRun {
		return nil
	}

	// Create new projects.
	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get project repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])
	var created, skipped int
	for _, sr := range repos {
		if sr.Status == "existing" {
			skipped++
			continue
		}
		p := &entities.Project{
			Name:          sr.Name,
			RepoURL:       sr.RemoteURL,
			LocalPath:     sr.Path,
			DefaultBranch: branch,
			DefaultModel:  model,
			Status:        string(enums.ProjectStatusActive),
		}
		if _, err := w.Insert(ctx, p); err != nil {
			cctx.Print("Warning: failed to register %s: %v", sr.Name, err)
			continue
		}
		created++
	}

	cctx.Print("Registered %d new projects, skipped %d existing", created, skipped)
	return nil
}
