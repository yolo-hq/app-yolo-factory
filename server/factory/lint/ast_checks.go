package lint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// walkGoFiles walks .go files under opts.Path, skipping _test.go, vendor/, testdata/.
// If opts.ChangedFiles is set, only those files are visited.
func walkGoFiles(opts Options, fn func(fset *token.FileSet, file *ast.File, path string) error) error {
	if len(opts.ChangedFiles) > 0 {
		for _, f := range opts.ChangedFiles {
			if err := processGoFile(f, fn); err != nil {
				return err
			}
		}
		return nil
	}

	return filepath.Walk(opts.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible
		}
		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		return processGoFile(path, fn)
	})
}

func processGoFile(path string, fn func(fset *token.FileSet, file *ast.File, path string) error) error {
	if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
		return nil
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil // skip unparseable files
	}
	return fn(fset, file, path)
}

// CheckSwallowedErrors finds assignments where all LHS identifiers are blank (_)
// and there are 2+ values on the LHS.
func CheckSwallowedErrors(opts Options) ([]Finding, error) {
	var findings []Finding

	err := walkGoFiles(opts, func(fset *token.FileSet, file *ast.File, path string) error {
		ast.Inspect(file, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok || len(assign.Lhs) < 2 {
				return true
			}
			allBlank := true
			for _, lhs := range assign.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok || ident.Name != "_" {
					allBlank = false
					break
				}
			}
			if allBlank {
				pos := fset.Position(assign.Pos())
				findings = append(findings, Finding{
					Check:    "swallowed-errors",
					Severity: SeverityError,
					File:     path,
					Line:     pos.Line,
					Message:  "all return values discarded — error likely swallowed",
				})
			}
			return true
		})
		return nil
	})

	return findings, err
}

// CheckNoShellExec finds exec.Command("sh", "-c", ...) or exec.CommandContext(ctx, "sh", "-c", ...) calls.
func CheckNoShellExec(opts Options) ([]Finding, error) {
	var findings []Finding

	err := walkGoFiles(opts, func(fset *token.FileSet, file *ast.File, path string) error {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			funcName := callFuncName(call)
			if funcName != "exec.Command" && funcName != "exec.CommandContext" {
				return true
			}

			// For exec.Command: args[0]="sh", args[1]="-c"
			// For exec.CommandContext: args[0]=ctx, args[1]="sh", args[2]="-c"
			argOffset := 0
			if funcName == "exec.CommandContext" {
				argOffset = 1
			}

			if len(call.Args) > argOffset+1 {
				if stringLitValue(call.Args[argOffset]) == "sh" && stringLitValue(call.Args[argOffset+1]) == "-c" {
					pos := fset.Position(call.Pos())
					findings = append(findings, Finding{
						Check:    "no-shell-exec",
						Severity: SeverityError,
						File:     path,
						Line:     pos.Line,
						Message:  "exec.Command with sh -c detected — use direct command execution",
					})
				}
			}

			return true
		})
		return nil
	})

	return findings, err
}

// CheckEntityMethods checks that files in entities/ dir with BaseEntity have TableName() and EntityName().
func CheckEntityMethods(opts Options) ([]Finding, error) {
	var findings []Finding

	err := walkGoFiles(opts, func(fset *token.FileSet, file *ast.File, path string) error {
		// Only scan files in entities/ dir.
		if !strings.Contains(path, "entities/") && !strings.Contains(path, "entities\\") {
			return nil
		}

		// Find struct types embedding BaseEntity.
		type entityStruct struct {
			name string
			pos  token.Pos
		}
		var entities []entityStruct

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range st.Fields.List {
					fieldType := typeString(field.Type)
					if strings.Contains(fieldType, "BaseEntity") {
						entities = append(entities, entityStruct{name: ts.Name.Name, pos: ts.Pos()})
					}
				}
			}
		}

		if len(entities) == 0 {
			return nil
		}

		// Collect method names per receiver type.
		methods := map[string]map[string]bool{}
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
				continue
			}
			recvType := receiverTypeName(funcDecl.Recv.List[0].Type)
			if methods[recvType] == nil {
				methods[recvType] = map[string]bool{}
			}
			methods[recvType][funcDecl.Name.Name] = true
		}

		for _, ent := range entities {
			m := methods[ent.name]
			for _, required := range []string{"TableName", "EntityName"} {
				if m == nil || !m[required] {
					pos := fset.Position(ent.pos)
					findings = append(findings, Finding{
						Check:    "entity-methods",
						Severity: SeverityError,
						File:     path,
						Line:     pos.Line,
						Message:  ent.name + " embeds BaseEntity but missing " + required + "()",
					})
				}
			}
		}

		return nil
	})

	return findings, err
}

// callFuncName extracts "pkg.Func" from a call expression.
func callFuncName(call *ast.CallExpr) string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name + "." + sel.Sel.Name
}

// stringLitValue returns the unquoted value of a basic string literal, or "".
func stringLitValue(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	// Remove quotes.
	v := lit.Value
	if len(v) >= 2 {
		return v[1 : len(v)-1]
	}
	return ""
}

// typeString returns a simple string for a type expression.
func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	}
	return ""
}

// receiverTypeName extracts the type name from a receiver expression.
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	}
	return ""
}
