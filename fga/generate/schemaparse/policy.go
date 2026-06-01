package schemaparse

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"entgo.io/ent/entc/load"
)

const (
	checkServiceCreateAccess = "CheckServiceCreateAccess"
	alwaysDenyQuery          = "AlwaysDenyRule"
)

var (
	internalOnlyPolicies = map[string]struct{}{
		"AllowMutationIfSystemAdmin": {},
		"AllowIfContextAllowRule":    {},
		"AllowQueryIfSystemAdmin":    {},
	}

	ignoreNames = map[string]struct{}{
		"NewPolicy":   {},
		"AllowIfSelf": {}, // these are user owned objects, not token permissions
		"AllowCreate": {}, // this only gives create access, not full crud
	}

	queryRuleNames = map[string]struct{}{
		"WithQueryRules": {},
	}

	mutationRuleNames = map[string]struct{}{
		"WithOnMutationRules": {},
		"WithMutationRules":   {},
	}
)

func (s *SchemaInfo) functionContainsCheckCreateAccess(body *ast.BlockStmt) {
	hasInternalPolicy := false

	inQueryRules := false
	inMutationRules := false

	numQueryRules := 0
	numMutationRules := 0

	policies := []string{}

	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		var sel *ast.SelectorExpr
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			sel = fun
		case *ast.IndexExpr:
			sel, _ = fun.X.(*ast.SelectorExpr)
		case *ast.IndexListExpr:
			sel, _ = fun.X.(*ast.SelectorExpr)
		}

		if sel == nil || sel.Sel == nil {
			return true
		}

		if _, ok := ignoreNames[sel.Sel.Name]; ok {
			return true
		}

		if _, ok := queryRuleNames[sel.Sel.Name]; ok {
			inQueryRules = true
			inMutationRules = false

			return true
		}

		if _, ok := mutationRuleNames[sel.Sel.Name]; ok {
			inQueryRules = false
			inMutationRules = true

			return true
		}

		if _, ok := internalOnlyPolicies[sel.Sel.Name]; ok {
			hasInternalPolicy = true

			return true
		}

		switch sel.Sel.Name {
		case checkServiceCreateAccess:
			s.CanCreateServiceOnly = true
		case alwaysDenyQuery:
			if inQueryRules {
				s.ExcludeFromGeneration = true
			}
		}

		if inQueryRules {
			numQueryRules++
		}

		if inMutationRules {
			numMutationRules++
		}

		policies = append(policies, sel.Sel.Name)

		return true
	})

	// do not generate permissions for schemas that only have internal access since they do not have any organization-level access rules defined
	if len(policies) == 0 && hasInternalPolicy {
		fmt.Printf("%s only has internal policies, skipping\n", s.Name)
		s.ExcludeFromGeneration = true
	}

	if numMutationRules == 0 && numQueryRules == 0 {
		s.ExcludeFromGeneration = true
	}
}

func parsePolicyBody(schema *load.Schema) *ast.BlockStmt {
	srcFile := strings.SplitN(schema.Pos, ":", 2)[0] //nolint:mnd
	if srcFile == "" {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcFile, nil, 0)
	if err != nil {
		return nil
	}

	var body *ast.BlockStmt
	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Policy" || fn.Recv == nil || fn.Body == nil {
			return true
		}

		for _, field := range fn.Recv.List {
			ident, ok := field.Type.(*ast.Ident)
			if !ok {
				if star, ok := field.Type.(*ast.StarExpr); ok {
					ident, _ = star.X.(*ast.Ident)
				}
			}

			if ident != nil && ident.Name == schema.Name {
				body = fn.Body
				return false
			}
		}

		return true
	})

	return body
}
