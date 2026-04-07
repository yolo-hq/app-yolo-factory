package lint

import (
	"path/filepath"
	"runtime"
	"testing"
)

func factoryRoot() string {
	_, file, _, _ := runtime.Caller(0)
	// lint/ -> factory/ -> server/ -> repo root
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
}

func TestLint_FactoryCodebase(t *testing.T) {
	result, err := RunAll(Options{Path: factoryRoot()})
	if err != nil {
		t.Fatal(err)
	}

	// Error-level findings fail the test.
	for _, f := range result.Findings {
		if f.Severity == SeverityError {
			t.Errorf("[%s] %s:%d — %s", f.Check, f.File, f.Line, f.Message)
		}
	}
	// Warning-level findings are logged but don't fail.
	for _, f := range result.Findings {
		if f.Severity == SeverityWarning {
			t.Logf("[WARN] [%s] %s:%d — %s", f.Check, f.File, f.Line, f.Message)
		}
	}
}
