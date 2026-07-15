package handlers

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"maps"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"sync"

	echo "github.com/theopenlane/echox"
	"golang.org/x/tools/go/packages"
)

// analyzedPackages are the packages parsed to resolve route handlers; handlers hold the Handler
// methods, route holds the package-level handler funcs for static routes
var analyzedPackages = []string{
	"github.com/theopenlane/core/internal/httpserve/handlers",
	"github.com/theopenlane/core/internal/httpserve/route",
}

// ResponseShape describes one response a handler can produce, derived from its source
type ResponseShape struct {
	// Type is the fully qualified payload type name, empty when the response carries no schema
	Type string
	// ContentType is the response media type, empty means application/json
	ContentType string
}

// HandlerAnalysis is everything the spec needs to know about a handler, derived from its source
type HandlerAnalysis struct {
	// Request is the fully qualified request model type name bound by the handler, empty when none
	Request string
	// Responses maps each status code the handler can write to its payload shape
	Responses map[int]ResponseShape
}

// typeEnv carries call-site type information into callees: generic type arguments and the concrete
// types of arguments bound to the callee's parameters
type typeEnv struct {
	typeParams map[*types.TypeParam]types.Type
	params     map[types.Object]types.Type
}

// sourceIndex holds the type-checked declarations of the analyzed packages
type sourceIndex struct {
	declsByName   map[string]*ast.FuncDecl
	declsByObject map[types.Object]*ast.FuncDecl
	info          *types.Info
}

var (
	sourceIndexOnce   sync.Once
	sharedSourceIndex *sourceIndex
	sourceIndexErr    error
)

// loadSourceIndex type-checks the analyzed packages once; the source is only present where the
// module is on disk (spec generation), which is the only place analysis runs
func loadSourceIndex() (*sourceIndex, error) {
	sourceIndexOnce.Do(func() {
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
				packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
		}

		pkgs, err := packages.Load(cfg, analyzedPackages...)
		if err != nil {
			sourceIndexErr = fmt.Errorf("%w: %w", ErrHandlerSourceUnavailable, err)
			return
		}

		idx := &sourceIndex{
			declsByName:   make(map[string]*ast.FuncDecl),
			declsByObject: make(map[types.Object]*ast.FuncDecl),
			info: &types.Info{
				Types:     make(map[ast.Expr]types.TypeAndValue),
				Defs:      make(map[*ast.Ident]types.Object),
				Uses:      make(map[*ast.Ident]types.Object),
				Instances: make(map[*ast.Ident]types.Instance),
			},
		}

		for _, pkg := range pkgs {
			if pkg.TypesInfo == nil {
				continue
			}

			maps.Copy(idx.info.Types, pkg.TypesInfo.Types)
			maps.Copy(idx.info.Defs, pkg.TypesInfo.Defs)
			maps.Copy(idx.info.Uses, pkg.TypesInfo.Uses)
			maps.Copy(idx.info.Instances, pkg.TypesInfo.Instances)

			for _, file := range pkg.Syntax {
				for _, decl := range file.Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
						idx.declsByName[fn.Name.Name] = fn

						if obj := pkg.TypesInfo.Defs[fn.Name]; obj != nil {
							idx.declsByObject[obj] = fn
						}
					}
				}
			}
		}

		if len(idx.declsByObject) == 0 {
			sourceIndexErr = ErrHandlerSourceUnavailable
			return
		}

		sharedSourceIndex = idx
	})

	return sharedSourceIndex, sourceIndexErr
}

// AnalyzeHandler derives the request model and response set of a route handler from its source
func AnalyzeHandler(handlerFn func(echo.Context) error) (*HandlerAnalysis, error) {
	idx, err := loadSourceIndex()
	if err != nil {
		return nil, err
	}

	name, err := handlerFuncName(handlerFn)
	if err != nil {
		return nil, err
	}

	decl, ok := idx.declsByName[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrHandlerDeclarationNotFound, name)
	}

	analysis := &HandlerAnalysis{Responses: make(map[int]ResponseShape)}

	idx.analyze(decl, typeEnv{}, analysis, map[*ast.FuncDecl]bool{})

	return analysis, nil
}

// handlerFuncName resolves the declaration name behind a route handler func value, either a bound
// Handler method or a package-level function
func handlerFuncName(handlerFn func(echo.Context) error) (string, error) {
	fn := runtime.FuncForPC(reflect.ValueOf(handlerFn).Pointer())
	if fn == nil {
		return "", ErrHandlerDeclarationNotFound
	}

	name := strings.TrimSuffix(fn.Name(), "-fm")

	return name[strings.LastIndex(name, ".")+1:], nil
}

// analyze walks the declaration body recording response writes and request bindings, following
// calls into other functions in the analyzed packages with call-site type information
func (idx *sourceIndex) analyze(decl *ast.FuncDecl, env typeEnv, analysis *HandlerAnalysis, walking map[*ast.FuncDecl]bool) {
	if walking[decl] {
		return
	}

	walking[decl] = true
	defer delete(walking, decl)

	ast.Inspect(decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if idx.recordResponseWrite(call, env, analysis) {
			return true
		}

		idx.followCall(call, env, analysis, walking)

		return true
	})
}

// recordResponseWrite detects echo response writes and json encoder writes, recording the status
// and payload shape; returns true when the call was a response write
func (idx *sourceIndex) recordResponseWrite(call *ast.CallExpr, env typeEnv, analysis *HandlerAnalysis) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || len(call.Args) == 0 {
		return false
	}

	receiver := idx.info.Types[sel.X].Type
	if receiver == nil {
		return false
	}

	receiverName := receiver.String()

	// a json encoder writing to the response is an application/json 200 payload
	if strings.HasSuffix(receiverName, "encoding/json.Encoder") && sel.Sel.Name == "Encode" {
		if shape, exists := analysis.Responses[http.StatusOK]; !exists || shape.Type == "" {
			encoded := ResponseShape{Type: idx.payloadTypeName(call.Args[0], env)}
			if encoded.Type == "" {
				encoded.ContentType = "application/json"
			}

			analysis.Responses[http.StatusOK] = encoded
		}

		return true
	}

	if !strings.HasSuffix(receiverName, "echox.Context") && !strings.HasSuffix(receiverName, "echox.Response") {
		return false
	}

	status, ok := idx.constantStatus(call.Args[0])
	if !ok {
		return false
	}

	shape := ResponseShape{}

	switch sel.Sel.Name {
	case "JSON", "JSONPretty":
		if len(call.Args) > 1 {
			shape.Type = idx.payloadTypeName(call.Args[1], env)
		}

		// payloads without a registrable model still document as a JSON object
		if shape.Type == "" {
			shape.ContentType = "application/json"
		}
	case "Blob":
		if len(call.Args) > 1 {
			if value := idx.info.Types[call.Args[1]].Value; value != nil && value.Kind() == constant.String {
				shape.ContentType = constant.StringVal(value)
			}
		}
	case "String", "HTML":
		shape.ContentType = "text/plain"
	case "File", "Attachment", "Stream":
		shape.ContentType = "application/octet-stream"
	}

	// never let a later write without payload information clobber a recorded shape
	if existing, exists := analysis.Responses[status]; exists && (existing.Type != "" || existing.ContentType != "") {
		return true
	}

	analysis.Responses[status] = shape

	return true
}

// constantStatus resolves an expression to a constant HTTP status code
func (idx *sourceIndex) constantStatus(expr ast.Expr) (int, bool) {
	value := idx.info.Types[expr].Value
	if value == nil || value.Kind() != constant.Int {
		return 0, false
	}

	status, ok := constant.Int64Val(value)
	if !ok || status < http.StatusContinue || status > http.StatusNetworkAuthenticationRequired {
		return 0, false
	}

	return int(status), true
}

// followCall resolves a call to a declaration in the analyzed packages and recurses into it,
// carrying generic type arguments and concrete parameter types from this call site; it also
// captures the request model from BindAndValidate instantiations
func (idx *sourceIndex) followCall(call *ast.CallExpr, env typeEnv, analysis *HandlerAnalysis, walking map[*ast.FuncDecl]bool) {
	ident := calleeIdent(call.Fun)
	if ident == nil {
		return
	}

	obj := idx.info.Uses[ident]
	if obj == nil {
		return
	}

	childEnv := typeEnv{
		typeParams: make(map[*types.TypeParam]types.Type),
		params:     make(map[types.Object]types.Type),
	}

	// map generic type parameters to the concrete type arguments of this instantiation
	if inst, ok := idx.info.Instances[ident]; ok && inst.TypeArgs != nil {
		if sig, ok := obj.Type().(*types.Signature); ok && sig.TypeParams() != nil {
			for i := 0; i < sig.TypeParams().Len() && i < inst.TypeArgs.Len(); i++ {
				childEnv.typeParams[sig.TypeParams().At(i)] = idx.resolveType(inst.TypeArgs.At(i), env)
			}
		}
	}

	// the binding primitive: its type argument is the operation's request model
	if obj.Name() == "BindAndValidate" && obj.Pkg() != nil && strings.HasSuffix(obj.Pkg().Path(), "httpserve/handlers") {
		if inst, ok := idx.info.Instances[ident]; ok && inst.TypeArgs != nil && inst.TypeArgs.Len() > 0 {
			if name := idx.qualifiedName(idx.resolveType(inst.TypeArgs.At(0), env)); name != "" && analysis.Request == "" {
				analysis.Request = name
			}
		}

		return
	}

	callee := idx.declsByObject[obj]
	if callee == nil {
		return
	}

	// bind the callee's parameters to the concrete argument types at this call site
	params := calleeParamObjects(idx.info, callee)
	for i, param := range params {
		if i >= len(call.Args) || param == nil {
			break
		}

		childEnv.params[param] = idx.exprType(call.Args[i], env)
	}

	idx.analyze(callee, childEnv, analysis, walking)
}

// calleeIdent unwraps a call target to its identifier, through selectors and generic instantiation
func calleeIdent(fun ast.Expr) *ast.Ident {
	switch expr := fun.(type) {
	case *ast.Ident:
		return expr
	case *ast.SelectorExpr:
		return expr.Sel
	case *ast.IndexExpr:
		return calleeIdent(expr.X)
	case *ast.IndexListExpr:
		return calleeIdent(expr.X)
	}

	return nil
}

// calleeParamObjects returns the callee's parameter objects in declaration order
func calleeParamObjects(info *types.Info, decl *ast.FuncDecl) []types.Object {
	var params []types.Object

	if decl.Type.Params == nil {
		return params
	}

	for _, field := range decl.Type.Params.List {
		for _, name := range field.Names {
			params = append(params, info.Defs[name])
		}
	}

	return params
}

// exprType resolves the static type of an expression, substituting call-site information for
// parameters and generic type parameters
func (idx *sourceIndex) exprType(expr ast.Expr, env typeEnv) types.Type {
	if ident, ok := expr.(*ast.Ident); ok {
		if obj := idx.info.Uses[ident]; obj != nil {
			if bound, exists := env.params[obj]; exists {
				return bound
			}
		}
	}

	return idx.resolveType(idx.info.Types[expr].Type, env)
}

// resolveType substitutes generic type parameters with their concrete arguments
func (idx *sourceIndex) resolveType(t types.Type, env typeEnv) types.Type {
	if param, ok := t.(*types.TypeParam); ok {
		if bound, exists := env.typeParams[param]; exists {
			return bound
		}
	}

	return t
}

// payloadTypeName resolves a response payload expression to a qualified named-struct type name,
// returning empty when the payload has no registrable schema
func (idx *sourceIndex) payloadTypeName(expr ast.Expr, env typeEnv) string {
	return idx.qualifiedName(idx.exprType(expr, env))
}

// qualifiedName renders a named struct type as its fully qualified name; non-struct and unnamed
// types yield empty, as do ent entities — those are database models whose edge graph would drag
// the entire schema universe into the spec, so handlers returning them get a schema-less JSON
// response instead
func (idx *sourceIndex) qualifiedName(t types.Type) string {
	if t == nil {
		return ""
	}

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return ""
	}

	if _, ok := named.Underlying().(*types.Struct); !ok {
		return ""
	}

	pkgPath := named.Obj().Pkg().Path()
	if strings.Contains(pkgPath, "/internal/ent/") {
		return ""
	}

	return pkgPath + "." + named.Obj().Name()
}
