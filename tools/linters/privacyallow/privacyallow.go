// Package privacyallow is a golangci-lint module plugin that reports privacy.Allow decision contexts;
// internal operations use auth capabilities such as auth.WithInternalOperationContext instead
package privacyallow

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const linterName = "privacyallow"

// defaultExclude is the set of path fragments skipped when no exclude setting is provided
var defaultExclude = []string{
	"internal/ent/generated/",
	"internal/ent/historygenerated/",
}

func init() {
	register.Plugin(linterName, New)
}

// Settings configures which files the linter checks
type Settings struct {
	// Include lists path fragments; when set, only files whose path contains one of them are checked
	Include []string `json:"include"`
	// Exclude lists path fragments; files whose path contains one of them are skipped
	Exclude []string `json:"exclude"`
}

// Plugin is the privacyallow linter plugin
type Plugin struct {
	settings Settings
}

// New builds the plugin from its golangci settings
func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}

	if len(s.Exclude) == 0 {
		s.Exclude = defaultExclude
	}

	return &Plugin{settings: s}, nil
}

// BuildAnalyzers returns the analyzers for the plugin
func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{p.Analyzer()}, nil
}

// GetLoadMode returns the load mode required by the analyzer
func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

// Analyzer returns the analyzer that reports allow decisions
func (p *Plugin) Analyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: linterName,
		Doc:  "reports privacy.DecisionContext(ctx, privacy.Allow) and privacy.Allowf decisions; use auth capabilities instead",
		Run:  p.run,
	}
}

func (p *Plugin) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if !p.included(filename) {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isAllowDecision(pass, call) {
				return true
			}

			pass.Reportf(call.Pos(), "privacy.Allow decision contexts are not allowed; use auth.WithInternalOperationContext or a scoped caller instead")

			return true
		})
	}

	return nil, nil
}

// included reports whether the file is checked based on the include and exclude settings
func (p *Plugin) included(filename string) bool {
	if strings.HasSuffix(filename, "_test.go") {
		return false
	}

	for _, fragment := range p.settings.Exclude {
		if strings.Contains(filename, fragment) {
			return false
		}
	}

	if len(p.settings.Include) == 0 {
		return true
	}

	for _, fragment := range p.settings.Include {
		if strings.Contains(filename, fragment) {
			return true
		}
	}

	return false
}

// isAllowDecision reports whether the call is privacy.DecisionContext(ctx, privacy.Allow) or privacy.DecisionContext(ctx, privacy.Allowf(...))
func isAllowDecision(pass *analysis.Pass, call *ast.CallExpr) bool {
	fn, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if !ok || fn.Name() != "DecisionContext" || !isPrivacyPkg(fn.Pkg()) || len(call.Args) != 2 {
		return false
	}

	switch decision := call.Args[1].(type) {
	case *ast.SelectorExpr:
		obj, ok := pass.TypesInfo.Uses[decision.Sel].(*types.Var)

		return ok && obj.Name() == "Allow" && isPrivacyPkg(obj.Pkg())
	case *ast.CallExpr:
		allowf, ok := typeutil.Callee(pass.TypesInfo, decision).(*types.Func)

		return ok && allowf.Name() == "Allowf" && isPrivacyPkg(allowf.Pkg())
	default:
		return false
	}
}

// isPrivacyPkg reports whether the package is ent's privacy package or a generated re-export of it
func isPrivacyPkg(pkg *types.Package) bool {
	return pkg != nil && (pkg.Path() == "entgo.io/ent/privacy" || strings.HasSuffix(pkg.Path(), "/generated/privacy"))
}
