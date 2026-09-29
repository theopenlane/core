// Package privacyallow is a golangci-lint module plugin that reports bare privacy.Allow decision
// contexts; an allow decision is only permitted when it is wrapped directly in a scoped caller
package privacyallow

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const (
	linterName = "privacyallow"
	authPkg    = "github.com/theopenlane/iam/auth"
)

// defaultInclude is the set of path fragments checked when no include setting is provided
var defaultInclude = []string{
	"internal/integrations/",
	"internal/ent/hooks/listeners_",
	"internal/ent/notifications/",
	"internal/workflows/",
}

// allowedWrappers are the auth functions that scope a caller around an allow decision
var allowedWrappers = map[string]bool{
	"WithCaller":              true,
	"EnsureIntegrationCaller": true,
}

func init() {
	register.Plugin(linterName, New)
}

// Settings configures which files the linter checks
type Settings struct {
	// Include lists path fragments; a file is checked when its path contains any of them
	Include []string `json:"include"`
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

	if len(s.Include) == 0 {
		s.Include = defaultInclude
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

// Analyzer returns the analyzer that reports bare allow decisions
func (p *Plugin) Analyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: linterName,
		Doc:  "reports privacy.DecisionContext(ctx, privacy.Allow) that is not wrapped directly in auth.WithCaller or auth.EnsureIntegrationCaller",
		Run:  p.run,
	}
}

func (p *Plugin) run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if !p.included(filename) {
			continue
		}

		wrapped := wrappedAllowCalls(pass, file)

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isAllowDecision(pass, call) || wrapped[call] {
				return true
			}

			pass.Reportf(call.Pos(), "privacy.Allow must be wrapped in a scoped caller (auth.WithCaller or auth.EnsureIntegrationCaller); use capabilities instead of a bare allow decision")

			return true
		})
	}

	return nil, nil
}

// included reports whether the file path matches one of the configured include fragments
func (p *Plugin) included(filename string) bool {
	if strings.HasSuffix(filename, "_test.go") {
		return false
	}

	for _, fragment := range p.settings.Include {
		if strings.Contains(filename, fragment) {
			return true
		}
	}

	return false
}

// wrappedAllowCalls returns the allow decision calls passed directly as the context to an allowed wrapper
func wrappedAllowCalls(pass *analysis.Pass, file *ast.File) map[*ast.CallExpr]bool {
	wrapped := map[*ast.CallExpr]bool{}

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 || !isAllowedWrapper(pass, call) {
			return true
		}

		if inner, ok := call.Args[0].(*ast.CallExpr); ok && isAllowDecision(pass, inner) {
			wrapped[inner] = true
		}

		return true
	})

	return wrapped
}

// isAllowedWrapper reports whether the call is one of the auth functions that scope a caller
func isAllowedWrapper(pass *analysis.Pass, call *ast.CallExpr) bool {
	fn, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if !ok || fn.Pkg() == nil {
		return false
	}

	return fn.Pkg().Path() == authPkg && allowedWrappers[fn.Name()]
}

// isAllowDecision reports whether the call is privacy.DecisionContext(ctx, privacy.Allow)
func isAllowDecision(pass *analysis.Pass, call *ast.CallExpr) bool {
	fn, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if !ok || fn.Name() != "DecisionContext" || !isPrivacyPkg(fn.Pkg()) || len(call.Args) != 2 {
		return false
	}

	sel, ok := call.Args[1].(*ast.SelectorExpr)
	if !ok {
		return false
	}

	obj, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Var)

	return ok && obj.Name() == "Allow" && isPrivacyPkg(obj.Pkg())
}

// isPrivacyPkg reports whether the package is ent's privacy package or a generated re-export of it
func isPrivacyPkg(pkg *types.Package) bool {
	return pkg != nil && (pkg.Path() == "entgo.io/ent/privacy" || strings.HasSuffix(pkg.Path(), "/generated/privacy"))
}
