package lint

// Severity levels for lint findings.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Finding represents a single lint issue.
type Finding struct {
	Check    string   `json:"check"`
	Severity Severity `json:"severity"`
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Message  string   `json:"message"`
}

// Result aggregates all lint check results.
type Result struct {
	Passed       bool
	ChecksRun    int
	ChecksPassed int
	ChecksFailed int
	Findings     []Finding
}

// Options configures the lint run.
type Options struct {
	Path          string
	ChangedFiles  []string // empty = all files
	TodoThreshold int      // default 20
}

// checkFunc is the signature for each lint check.
type checkFunc func(Options) ([]Finding, error)

// RunAll runs all checks and aggregates findings.
func RunAll(opts Options) (*Result, error) {
	if opts.TodoThreshold <= 0 {
		opts.TodoThreshold = 20
	}

	checks := []checkFunc{
		CheckSwallowedErrors,
		CheckNoShellExec,
		CheckEntityMethods,
		CheckStatusLiterals,
		CheckDuplicateFunctions,
		CheckStubFunctions,
		CheckTodoThreshold,
		CheckResolveBeforeOK,
	}

	result := &Result{}
	var allFindings []Finding

	for _, check := range checks {
		findings, err := check(opts)
		if err != nil {
			return nil, err
		}
		result.ChecksRun++
		hasError := false
		for _, f := range findings {
			if f.Severity == SeverityError {
				hasError = true
			}
		}
		if hasError || len(findings) > 0 {
			result.ChecksFailed++
		} else {
			result.ChecksPassed++
		}
		allFindings = append(allFindings, findings...)
	}

	result.Findings = allFindings

	// Passed = no error-severity findings.
	result.Passed = true
	for _, f := range allFindings {
		if f.Severity == SeverityError {
			result.Passed = false
			break
		}
	}

	return result, nil
}
