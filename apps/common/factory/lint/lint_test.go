package lint

import (
	"path/filepath"
	"runtime"
	"testing"
)

func testdataPath() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata")
}

func optsForFile(name string) Options {
	return Options{
		Path:         testdataPath(),
		ChangedFiles: []string{filepath.Join(testdataPath(), name)},
	}
}

func TestCheckSwallowedErrors_Finds(t *testing.T) {
	findings, err := CheckSwallowedErrors(optsForFile("bad_swallowed.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("expected findings for swallowed errors")
	}
	if findings[0].Check != "swallowed-errors" {
		t.Fatalf("wrong check: %s", findings[0].Check)
	}
	if findings[0].Severity != SeverityError {
		t.Fatalf("expected error severity, got %s", findings[0].Severity)
	}
}

func TestCheckSwallowedErrors_Clean(t *testing.T) {
	findings, err := CheckSwallowedErrors(optsForFile("good.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %d", len(findings))
	}
}

func TestCheckNoShellExec_Finds(t *testing.T) {
	findings, err := CheckNoShellExec(optsForFile("bad_shell.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("expected findings for shell exec")
	}
	if findings[0].Check != "no-shell-exec" {
		t.Fatalf("wrong check: %s", findings[0].Check)
	}
}

func TestCheckNoShellExec_Clean(t *testing.T) {
	findings, err := CheckNoShellExec(optsForFile("good.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %d", len(findings))
	}
}

func TestCheckStubFunctions_Finds(t *testing.T) {
	findings, err := CheckStubFunctions(optsForFile("bad_stubs.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("expected findings for stub functions")
	}
	if findings[0].Check != "stub-functions" {
		t.Fatalf("wrong check: %s", findings[0].Check)
	}
}

func TestCheckStatusLiterals_Finds(t *testing.T) {
	findings, err := CheckStatusLiterals(optsForFile("bad_status.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("expected findings for status literals")
	}
	if findings[0].Check != "status-literals" {
		t.Fatalf("wrong check: %s", findings[0].Check)
	}
}

func TestCheckTodoThreshold(t *testing.T) {
	// Use threshold of 0 to trigger finding on any TODO.
	opts := Options{
		Path:          testdataPath(),
		ChangedFiles:  []string{filepath.Join(testdataPath(), "good.go")},
		TodoThreshold: 0,
	}
	findings, err := CheckTodoThreshold(opts)
	if err != nil {
		t.Fatal(err)
	}
	// good.go has no TODOs, so no findings with threshold 0.
	// This validates the count logic works without error.
	if len(findings) != 0 {
		t.Fatalf("expected no findings for clean file with threshold 0, got %d", len(findings))
	}
}

func TestRunAll(t *testing.T) {
	opts := Options{
		Path:          testdataPath(),
		ChangedFiles:  []string{filepath.Join(testdataPath(), "good.go")},
		TodoThreshold: 100,
	}
	result, err := RunAll(opts)
	if err != nil {
		t.Fatal(err)
	}
	if result.ChecksRun != 7 {
		t.Fatalf("expected 7 checks, got %d", result.ChecksRun)
	}
	if !result.Passed {
		t.Fatal("expected clean file to pass")
	}
}
