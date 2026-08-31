package project

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"
)

// Route source/confidence values (GIN-014).
const (
	RouteSourceAnnotation = "annotation"
	RouteSourceAST        = "ast"
	RouteSourceRegex      = "regex"

	RouteConfidenceHigh   = "high"
	RouteConfidenceMedium = "medium"
	RouteConfidenceLow    = "low"
)

// routeGroupPrefixes maps local identifiers to their Group("/prefix") value.
// Interprocedural (1 level): a function whose parameter receives a prefixed
// router at the call site inherits that prefix (GIN-014 — the generated code
// pattern `v1 := r.Group("/api/v1"); registerCoreRoutes(v1)` needs it).
type routeGroupPrefixes struct {
	prefixes map[string]string // identifier → prefix
	byFunc   map[string]map[string]string // funcName → paramName → prefix
}

func newRouteGroupPrefixes() *routeGroupPrefixes {
	return &routeGroupPrefixes{prefixes: map[string]string{}, byFunc: map[string]map[string]string{}}
}

// astRoutes extracts HTTP routes from parsed Go source with correctness the
// line-regex cannot offer: group prefix composition, method filters on real
// call expressions, and annotation precedence.
func astRoutes(content, file string) []Route {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, content, parser.ParseComments)
	if err != nil {
		return nil // fallback regex handles unparsable files
	}

	gp := newRouteGroupPrefixes()
	var routes []Route
	seen := map[string]bool{}

	add := func(method, path string, line int, source, confidence string) {
		key := method + " " + path
		if seen[key] {
			return
		}
		seen[key] = true
		routes = append(routes, Route{Method: method, Path: path, File: file, Line: line, Source: source, Confidence: confidence})
	}

	// ── Pass 1: annotations on function declarations (authoritative).
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Doc == nil {
			continue
		}
		for _, c := range fd.Doc.List {
			text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
			fields := strings.Fields(text)
			if len(fields) < 3 || fields[0] != "ginger:route" {
				continue
			}
			pos := fset.Position(c.Pos())
			add(strings.ToUpper(fields[1]), fields[2], pos.Line, RouteSourceAnnotation, RouteConfidenceHigh)
		}
	}

	// ── Pass 2: collect Group prefixes (assignments + call-site inheritance).
	callPrefixArgs := map[*ast.CallExpr][]string{} // call → arg idents (strings) at call site
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if ok {
			for i, rhs := range assign.Rhs {
				call, ok := rhs.(*ast.CallExpr)
				if !ok {
					continue
				}
				if prefix, receiver := groupCall(call); prefix != "" {
					if ident, ok := assignLhsIdent(assign, i); ok {
						gp.prefixes[ident.Name] = prefix
						_ = receiver
					}
				}
			}
			return true
		}
		expr, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}
		stmtCall, ok := expr.X.(*ast.CallExpr)
		if !ok {
			return true
		}
		// registerCoreRoutes(v1) — remember the string idents passed positionally.
		var args []string
		for _, a := range stmtCall.Args {
			if id, ok := a.(*ast.Ident); ok {
				args = append(args, id.Name)
			} else {
				args = append(args, "")
			}
		}
		callPrefixArgs[stmtCall] = args
		return true
	})

	// Pass 2b: function declarations whose params receive a prefixed router.
	for stmtCall, args := range callPrefixArgs {
		fnName := calleeName(stmtCall)
		if fnName == "" {
			continue
		}
		var fd *ast.FuncDecl
		for _, decl := range f.Decls {
			if d, ok := decl.(*ast.FuncDecl); ok && d.Name.Name == fnName {
				fd = d
				break
			}
		}
		if fd == nil {
			continue
		}
		idx := 0 // flattened param index
		for _, p := range fd.Type.Params.List {
			for _, paramName := range p.Names {
				if idx < len(args) && args[idx] != "" {
					if prefix, ok := gp.prefixes[args[idx]]; ok {
						if gp.byFunc[fnName] == nil {
							gp.byFunc[fnName] = map[string]string{}
						}
						gp.byFunc[fnName][paramName.Name] = prefix
					}
				}
				idx++
			}
		}
	}

	// ── Pass 3: route expressions with prefix composition.
	var currentFunc string
	ast.Inspect(f, func(n ast.Node) bool {
		fd, ok := n.(*ast.FuncDecl)
		if ok {
			currentFunc = fd.Name.Name
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		method := sel.Sel.Name
		recvIdent := ""
		if id, ok := sel.X.(*ast.Ident); ok {
			recvIdent = id.Name
		}

		line := fset.Position(call.Pos()).Line
		prefix := gp.prefixes[recvIdent]
		if byFunc, ok := gp.byFunc[currentFunc]; ok {
			if p, ok := byFunc[recvIdent]; ok {
				prefix = p
			}
		}

		switch method {
		case "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD":
			if len(call.Args) == 0 {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true // dynamic path — regex fallback may catch literals
			}
			path, err := strconv.Unquote(lit.Value)
			if err != nil || !looksLikeRoutePath(path) {
				return true
			}
			add(method, joinPath(prefix, path), line, RouteSourceAST, RouteConfidenceHigh)
		case "HandleFunc":
			if len(call.Args) == 0 {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			raw, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			m := strings.Fields(raw)
			pathIdx := 0
			if len(m) >= 2 {
				pathIdx = 1
			}
			if pathIdx < len(m) && !looksLikeRoutePath(m[pathIdx]) {
				return true
			}
			switch {
			case len(m) >= 2:
				add(strings.ToUpper(m[0]), joinPath(prefix, m[1]), line, RouteSourceAST, RouteConfidenceHigh)
			case len(m) == 1:
				add("ANY", joinPath(prefix, m[0]), line, RouteSourceAST, RouteConfidenceHigh)
			}
		}
		return true
	})

	return routes
}

func groupCall(call *ast.CallExpr) (prefix string, receiver string) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Group" || len(call.Args) == 0 {
		return "", ""
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", ""
	}
	p, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", ""
	}
	if id, ok := sel.X.(*ast.Ident); ok {
		return p, id.Name
	}
	return p, ""
}

func assignLhsIdent(assign *ast.AssignStmt, i int) (*ast.Ident, bool) {
	if i >= len(assign.Lhs) {
		return nil, false
	}
	id, ok := assign.Lhs[i].(*ast.Ident)
	return id, ok
}

func calleeName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		return fn.Sel.Name
	}
	return ""
}

// looksLikeRoutePath rejects URLs and non-path literals (GIN-014: no false
// positives like client.GET("https://api.example.com/...")).
func looksLikeRoutePath(path string) bool {
	if path == "" || !strings.HasPrefix(path, "/") {
		return false
	}
	return !strings.Contains(path, "://")
}

// joinPath composes a group prefix with a route path (GIN-014 correctness).
func joinPath(prefix, path string) string {
	if prefix == "" {
		return path
	}
	if path == "" || path == "/" {
		return strings.TrimSuffix(prefix, "/") + "/"
	}
	return strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(path, "/")
}

// mergeRouteResults dedupes AST routes with regex fallback routes; the regex
// only contributes entries the AST missed (low confidence).
func mergeRouteResults(astRoutesList, regexRoutes []Route) []Route {
	seen := map[string]bool{}
	var out []Route
	for _, r := range astRoutesList {
		key := r.Method + " " + r.Path + " " + r.File + ":" + itoa(r.Line)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, r)
	}
	for _, r := range regexRoutes {
		r.Source = RouteSourceRegex
		r.Confidence = RouteConfidenceLow
		// Regex matching by path: skip if AST already found the same path
		// (possibly with composed prefix).
		dup := false
		for _, a := range astRoutesList {
			if a.Path == r.Path && a.Method == r.Method {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		key := r.Method + " " + r.Path + " " + r.File + ":" + itoa(r.Line)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		if out[i].Method != out[j].Method {
			return out[i].Method < out[j].Method
		}
		return out[i].Line < out[j].Line
	})
	return out
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
