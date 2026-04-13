package lint

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var statusLiteralPattern = regexp.MustCompile(`"(queued|blocked|running|reviewing|done|failed|cancelled|active|paused|archived|draft|approved|planning|in_progress|completed|open|answered|auto_resolved|pending|rejected|converted)"`)

var skipStatusFiles = map[string]bool{
	"status.go":  true,
	"helpers.go": true,
}

var skipStatusLinePatterns = []string{"fake:", "default:", "bun:"}

// CheckStatusLiterals finds hardcoded status strings that should use constants.
func CheckStatusLiterals(opts Options) ([]Finding, error) {
	var findings []Finding

	err := walkTextFiles(opts, func(path string) error {
		base := filepath.Base(path)
		if skipStatusFiles[base] || strings.HasSuffix(base, "_test.go") {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Skip lines with struct tags.
			skip := false
			for _, p := range skipStatusLinePatterns {
				if strings.Contains(line, p) {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			if statusLiteralPattern.MatchString(line) {
				match := statusLiteralPattern.FindString(line)
				findings = append(findings, Finding{
					Check:    "status-literals",
					Severity: SeverityWarning,
					File:     path,
					Line:     lineNum,
					Message:  fmt.Sprintf("status literal %s — use constant from entities/status.go", match),
				})
			}
		}
		return nil
	})

	return findings, err
}

var exportedFuncPattern = regexp.MustCompile(`^func (\([^)]+\) )?([A-Z]\w+)\(`)

// CheckDuplicateFunctions finds exported functions with the same name across files.
func CheckDuplicateFunctions(opts Options) ([]Finding, error) {
	type loc struct {
		file string
		line int
	}
	funcLocs := map[string][]loc{}

	err := walkTextFiles(opts, func(path string) error {
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			matches := exportedFuncPattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				name := matches[2]
				funcLocs[name] = append(funcLocs[name], loc{file: path, line: lineNum})
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var findings []Finding
	for name, locs := range funcLocs {
		if len(locs) <= 1 {
			continue
		}
		// Only report if in different files.
		files := map[string]bool{}
		for _, l := range locs {
			files[l.file] = true
		}
		if len(files) <= 1 {
			continue
		}
		for _, l := range locs {
			findings = append(findings, Finding{
				Check:    "duplicate-functions",
				Severity: SeverityWarning,
				File:     l.file,
				Line:     l.line,
				Message:  fmt.Sprintf("exported function %s also defined in another file", name),
			})
		}
	}

	return findings, nil
}

// CheckStubFunctions finds functions whose body only contains fmt.Print* calls.
func CheckStubFunctions(opts Options) ([]Finding, error) {
	var findings []Finding

	fmtPrintPattern := regexp.MustCompile(`^\s*fmt\.Print`)
	funcStartPattern := regexp.MustCompile(`^func `)

	err := walkTextFiles(opts, func(path string) error {
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineNum := 0
		inFunc := false
		funcLine := 0
		funcName := ""
		braceDepth := 0
		hasFmtPrint := false
		hasOther := false

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)

			if !inFunc {
				if funcStartPattern.MatchString(line) {
					inFunc = true
					funcLine = lineNum
					// Extract func name.
					parts := strings.Fields(line)
					for i, p := range parts {
						if p == "func" && i+1 < len(parts) {
							name := parts[i+1]
							if idx := strings.Index(name, "("); idx >= 0 {
								name = name[:idx]
							}
							funcName = name
							break
						}
					}
					braceDepth = strings.Count(line, "{") - strings.Count(line, "}")
					hasFmtPrint = false
					hasOther = false
					if fmtPrintPattern.MatchString(line) {
						hasFmtPrint = true
					}
				}
				continue
			}

			braceDepth += strings.Count(line, "{") - strings.Count(line, "}")

			if braceDepth <= 0 {
				// Function ended.
				if hasFmtPrint && !hasOther {
					findings = append(findings, Finding{
						Check:    "stub-functions",
						Severity: SeverityWarning,
						File:     path,
						Line:     funcLine,
						Message:  fmt.Sprintf("function %s appears to be a stub (only fmt.Print calls)", funcName),
					})
				}
				inFunc = false
				continue
			}

			if trimmed == "" || trimmed == "{" || trimmed == "}" {
				continue
			}

			if fmtPrintPattern.MatchString(line) {
				hasFmtPrint = true
			} else {
				hasOther = true
			}
		}
		return nil
	})

	return findings, err
}

// CheckTodoThreshold counts TODO/FIXME/HACK comments across all .go files.
func CheckTodoThreshold(opts Options) ([]Finding, error) {
	todoPattern := regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK)\b`)
	count := 0

	err := walkTextFiles(opts, func(path string) error {
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if todoPattern.MatchString(scanner.Text()) {
				count++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if count > opts.TodoThreshold {
		return []Finding{{
			Check:    "todo-threshold",
			Severity: SeverityInfo,
			File:     "",
			Line:     0,
			Message:  fmt.Sprintf("%d TODO/FIXME/HACK comments found (threshold: %d)", count, opts.TodoThreshold),
		}}, nil
	}

	return nil, nil
}

// walkTextFiles walks .go files (not _test.go, vendor/, testdata/) and calls fn with the path.
func walkTextFiles(opts Options, fn func(path string) error) error {
	if len(opts.ChangedFiles) > 0 {
		for _, f := range opts.ChangedFiles {
			if !strings.HasSuffix(f, ".go") || strings.HasSuffix(f, "_test.go") {
				continue
			}
			if err := fn(f); err != nil {
				return err
			}
		}
		return nil
	}

	return filepath.Walk(opts.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == "testdata" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		return fn(path)
	})
}
