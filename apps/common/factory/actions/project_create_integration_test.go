//go:build integration

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

func TestCreateProject_HappyPath(t *testing.T) {
	tx := dbTx(t)

	result := runAction(t, tx, &CreateProjectAction{},
		yolotest.WithInput(inputs.CreateProjectInput{
			Name:      "My Project " + newID(),
			RepoURL:   "https://github.com/test/repo",
			LocalPath: "/tmp/myproject",
		}),
	)
	require.True(t, result.Success, "create should succeed: %s", result.Message)
}

func TestCreateProject_DenyMissingName(t *testing.T) {
	tx := dbTx(t)

	result := runAction(t, tx, &CreateProjectAction{},
		yolotest.WithInput(inputs.CreateProjectInput{
			RepoURL:   "https://github.com/test/repo",
			LocalPath: "/tmp/myproject",
			// Missing Name
		}),
	)
	assert.False(t, result.Success, "create should fail without name")
	assert.Equal(t, 422, result.StatusCode)
}

func TestCreateProject_DenyMissingRepoURL(t *testing.T) {
	tx := dbTx(t)

	result := runAction(t, tx, &CreateProjectAction{},
		yolotest.WithInput(inputs.CreateProjectInput{
			Name:      "My Project " + newID(),
			LocalPath: "/tmp/myproject",
			// Missing RepoURL
		}),
	)
	assert.False(t, result.Success, "create should fail without repo URL")
	assert.Equal(t, 422, result.StatusCode)
}