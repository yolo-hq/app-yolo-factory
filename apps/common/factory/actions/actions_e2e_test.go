//go:build e2e

package actions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// TestE2E_ProjectToPRDApproval tests the complete Project→PRD→Approve flow.
func TestE2E_ProjectToPRDApproval(t *testing.T) {
	tx := dbTx(t)
	ctx := context.Background()

	// Step 1: Create project.
	suffix := newID()
	projectName := "e2e-proj-" + suffix
	projectResult := runAction(t, tx, &CreateProjectAction{},
		yolotest.WithInput(inputs.CreateProjectInput{
			Name:          projectName,
			RepoURL:       "https://github.com/test/e2e-" + suffix,
			LocalPath:     "/tmp/e2e-" + suffix,
			DefaultBranch: "main",
		}),
	)
	require.True(t, projectResult.Success, "CreateProject should succeed: %s", projectResult.Message)

	// Look up the project by name to get its ID.
	var proj entities.Project
	err := tx.NewSelect().Model(&proj).Where("name = ?", projectName).Scan(ctx)
	require.NoError(t, err, "project should exist in DB")
	require.NotEmpty(t, proj.ID, "project ID must be set")
	assert.Equal(t, "active", proj.Status)

	// Step 2: Submit PRD.
	prdResult := runAction(t, tx, &SubmitPRDAction{},
		yolotest.WithInput(inputs.SubmitPRDInput{
			ProjectID:          proj.ID,
			Title:              "E2E Feature: User Auth",
			Body:               "Implement JWT-based user authentication",
			AcceptanceCriteria: "Users can register, login, and refresh tokens",
		}),
	)
	require.True(t, prdResult.Success, "SubmitPRD should succeed: %s", prdResult.Message)

	// Look up PRD by project ID (there's only one in this TX).
	var prd entities.PRD
	err = tx.NewSelect().Model(&prd).Where("project_id = ? AND title = ?", proj.ID, "E2E Feature: User Auth").Scan(ctx)
	require.NoError(t, err, "PRD should exist in DB")
	require.NotEmpty(t, prd.ID, "PRD ID must be set")
	assert.Equal(t, "draft", prd.Status, "PRD should be draft after submit")

	// Step 3: Approve PRD.
	approveResult := runAction(t, tx, &PRDApproveAction{},
		yolotest.WithEntityID(prd.ID),
		yolotest.WithEntityName("PRD"),
	)
	require.True(t, approveResult.Success, "ApprovePRD should succeed: %s", approveResult.Message)

	// Step 4: Verify final state.
	assertPRDStatus(t, tx, prd.ID, "approved")

	var approvedPRD entities.PRD
	err = tx.NewSelect().Model(&approvedPRD).Where("id = ?", prd.ID).Scan(ctx)
	require.NoError(t, err)
	assert.NotNil(t, approvedPRD.ApprovedAt, "approved_at should be set after approval")
}

// TestE2E_CreateAndArchiveProject tests project create→archive lifecycle.
func TestE2E_CreateAndArchiveProject(t *testing.T) {
	tx := dbTx(t)
	ctx := context.Background()

	suffix := newID()
	projectName := "archive-proj-" + suffix
	createResult := runAction(t, tx, &CreateProjectAction{},
		yolotest.WithInput(inputs.CreateProjectInput{
			Name:      projectName,
			RepoURL:   "https://github.com/test/archive-" + suffix,
			LocalPath: "/tmp/archive-" + suffix,
		}),
	)
	require.True(t, createResult.Success, "CreateProject should succeed")

	var proj entities.Project
	err := tx.NewSelect().Model(&proj).Where("name = ?", projectName).Scan(ctx)
	require.NoError(t, err)

	archiveResult := runAction(t, tx, &ProjectArchiveAction{},
		yolotest.WithEntityID(proj.ID),
		yolotest.WithEntityName("Project"),
	)
	require.True(t, archiveResult.Success, "ArchiveProject should succeed: %s", archiveResult.Message)
	assertProjectStatus(t, tx, proj.ID, "archived")
}

// TestE2E_SubmitPRD_PolicyDenied verifies SubmitPRD is denied for paused projects.
func TestE2E_SubmitPRD_PolicyDenied(t *testing.T) {
	tx := dbTx(t)

	proj := seedProject(t, tx, &entities.Project{Status: "paused"})

	result := runAction(t, tx, &SubmitPRDAction{},
		yolotest.WithInput(inputs.SubmitPRDInput{
			ProjectID:          proj.ID,
			Title:              "T",
			Body:               "B",
			AcceptanceCriteria: "AC",
		}),
	)
	assert.False(t, result.Success, "SubmitPRD should be denied for paused project")
}

// TestE2E_PauseAndResumeProject tests project pause → resume lifecycle.
func TestE2E_PauseAndResumeProject(t *testing.T) {
	tx := dbTx(t)

	proj := seedProject(t, tx, nil) // active

	pauseResult := runAction(t, tx, &ProjectPauseAction{},
		yolotest.WithEntityID(proj.ID),
		yolotest.WithEntityName("Project"),
	)
	require.True(t, pauseResult.Success, "PauseProject should succeed: %s", pauseResult.Message)
	assertProjectStatus(t, tx, proj.ID, "paused")

	resumeResult := runAction(t, tx, &ProjectResumeAction{},
		yolotest.WithEntityID(proj.ID),
		yolotest.WithEntityName("Project"),
	)
	require.True(t, resumeResult.Success, "ResumeProject should succeed: %s", resumeResult.Message)
	assertProjectStatus(t, tx, proj.ID, "active")
}

// idFromData extracts "id" from action result Data.
// Returns "" if not found.
func idFromData(result yolotest.ActionResult) string {
	if m, ok := result.Data.(map[string]any); ok {
		for _, v := range m {
			if ent, ok2 := v.(map[string]any); ok2 {
				if id, ok3 := ent["id"].(string); ok3 && id != "" {
					return id
				}
			}
		}
	}
	return ""
}

// selectByID loads an entity by ID from the tx.
func selectByID[T any](t testing.TB, tx bun.Tx, id string) *T {
	t.Helper()
	var result T
	err := tx.NewSelect().Model(&result).Where("id = ?", id).Scan(context.Background())
	if err != nil {
		t.Fatalf("selectByID %T %q: %v", result, id, err)
	}
	return &result
}