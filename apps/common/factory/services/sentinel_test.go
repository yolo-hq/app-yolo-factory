package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestSentinel_ParseFindings(t *testing.T) {
	// Verify Finding struct fields.
	f := Finding{
		Watch:    "build_health",
		Severity: "critical",
		Message:  "build failed: missing package",
		Action:   "create_task",
	}
	assert.Equal(t, "build_health", f.Watch)
	assert.Equal(t, "critical", f.Severity)
	assert.Equal(t, "create_task", f.Action)
}

func TestSentinel_CategoryFromWatch(t *testing.T) {
	assert.Equal(t, "bug", categoryFromWatch("build_health"))
	assert.Equal(t, "bug", categoryFromWatch("test_health"))
	assert.Equal(t, "bug", categoryFromWatch("security"))
	assert.Equal(t, "refactor", categoryFromWatch("convention_drift"))
	assert.Equal(t, "refactor", categoryFromWatch("unknown"))
}

func TestSentinel_BuildHealthPass(t *testing.T) {
	svc := &SentinelService{}
	// Use a known-good Go directory. t.TempDir() won't have Go files,
	// but we can test with a trivial Go module.
	dir := t.TempDir()
	writeGoModule(t, dir)

	findings, err := svc.checkBuild(t.Context(), entities.Project{LocalPath: dir})
	assert.NoError(t, err)
	assert.Len(t, findings, 1)
	assert.Equal(t, "info", findings[0].Severity)
	assert.Equal(t, "none", findings[0].Action)
}

func TestSentinel_BuildHealthFail(t *testing.T) {
	svc := &SentinelService{}
	dir := t.TempDir()
	writeBrokenGoModule(t, dir)

	findings, err := svc.checkBuild(t.Context(), entities.Project{LocalPath: dir})
	assert.NoError(t, err)
	assert.Len(t, findings, 1)
	assert.Equal(t, "critical", findings[0].Severity)
	assert.Equal(t, "create_task", findings[0].Action)
	assert.Contains(t, findings[0].Message, "build failed")
}

// writeGoModule creates a minimal valid Go module in dir.
func writeGoModule(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, dir, "go.mod", "module test\n\ngo 1.21\n")
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
}

// writeBrokenGoModule creates a Go module with a syntax error.
func writeBrokenGoModule(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, dir, "go.mod", "module test\n\ngo 1.21\n")
	writeFile(t, dir, "main.go", "package main\n\nfunc main() { undefined_symbol }\n")
}

func TestSentinel_GoVet_Pass(t *testing.T) {
	svc := &SentinelService{}
	dir := t.TempDir()
	writeGoModule(t, dir)

	findings, err := svc.checkSecurity(t.Context(), entities.Project{LocalPath: dir})
	assert.NoError(t, err)
	// go vet should pass on a clean module.
	vetFinding := findingByMessage(findings, "go vet clean")
	assert.NotNil(t, vetFinding, "expected go vet clean finding")
}

func TestSentinel_GoVet_Fail(t *testing.T) {
	svc := &SentinelService{}
	dir := t.TempDir()
	writeGoModule(t, dir)
	// printf with wrong arg count triggers go vet.
	writeFile(t, dir, "bad.go", `package main
import "fmt"
func bad() { fmt.Printf("%d %d", 1) }
`)

	findings, err := svc.checkSecurity(t.Context(), entities.Project{LocalPath: dir})
	assert.NoError(t, err)
	vetFinding := findingByWatch(findings, "security", "warning")
	assert.NotNil(t, vetFinding, "expected a warning from go vet")
	assert.Equal(t, "create_suggestion", vetFinding.Action)
}

func TestSentinel_ScanHardcodedCredentials_Found(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "config.go", `package main
var x = password = "s3cr3t"
`)

	findings := scanHardcodedCredentials(dir)
	assert.NotEmpty(t, findings, "should detect hardcoded credential")
	assert.Equal(t, "warning", findings[0].Severity)
	assert.Equal(t, "create_suggestion", findings[0].Action)
}

func TestSentinel_ScanHardcodedCredentials_SkipsTests(t *testing.T) {
	dir := t.TempDir()
	// credential in a _test.go file — should be ignored.
	writeFile(t, dir, "config_test.go", `package main
var x = password = "s3cr3t"
`)

	findings := scanHardcodedCredentials(dir)
	assert.Empty(t, findings, "should skip test files")
}

func TestSentinel_ScanHardcodedCredentials_None(t *testing.T) {
	dir := t.TempDir()
	writeGoModule(t, dir)

	findings := scanHardcodedCredentials(dir)
	assert.Empty(t, findings)
}

// findingByMessage returns the first finding whose message contains s, or nil.
func findingByMessage(findings []Finding, s string) *Finding {
	for i := range findings {
		if strings.Contains(findings[i].Message, s) {
			return &findings[i]
		}
	}
	return nil
}

// findingByWatch returns the first finding with matching watch and severity, or nil.
func findingByWatch(findings []Finding, watch, severity string) *Finding {
	for i := range findings {
		if findings[i].Watch == watch && findings[i].Severity == severity {
			return &findings[i]
		}
	}
	return nil
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
