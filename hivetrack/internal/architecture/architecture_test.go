// Package architecture contains tests that enforce architectural constraints.
// These tests verify that layer boundaries are respected and that the CQRS
// structure is not accidentally violated as the codebase grows.
//
// Run with: just test-arch
//
// Layer dependency rules (what each layer is ALLOWED to import):
//
//	setup          → anything (composition root, wires everything together)
//	server         → handlers, middlewares
//	handlers       → commands, queries, models, authentication
//	commands       → repositories (interfaces only), models, events, email
//	queries        → repositories (interfaces only), models
//	behaviors      → models, authentication
//	middlewares    → authentication, models
//	repositories/* → models, database  (implementations stay behind interfaces)
//	authentication → models
//	events         → models
//	email          → models
//	models         → nothing internal  (pure domain types, no dependencies)
//	database       → config
//	config         → nothing internal
package architecture_test

import (
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

const module = "github.com/the127/hivetrack"

// pkg builds a full internal package path from a short name.
func pkg(short string) string {
	return module + "/internal/" + short
}

// rule describes a single architectural constraint.
type rule struct {
	name          string
	from          []string // packages that must NOT import...
	mustNotImport []string // ...these packages (prefix match)
}

var rules = []rule{
	{
		name: "handlers must not directly import repository implementations",
		from: []string{pkg("handlers")},
		mustNotImport: []string{
			pkg("repositories/postgres"),
			pkg("repositories/inmemory"),
		},
	},
	{
		name: "handlers must not import setup (setup wires handlers, not the other way)",
		from: []string{pkg("handlers")},
		mustNotImport: []string{pkg("setup")},
	},
	{
		name: "commands must not import queries (CQRS separation)",
		from: []string{pkg("commands")},
		mustNotImport: []string{pkg("queries")},
	},
	{
		name: "queries must not import commands (CQRS separation)",
		from: []string{pkg("queries")},
		mustNotImport: []string{pkg("commands")},
	},
	{
		name: "commands must not import handlers or middlewares",
		from: []string{pkg("commands")},
		mustNotImport: []string{
			pkg("handlers"),
			pkg("middlewares"),
			pkg("setup"),
		},
	},
	{
		name: "queries must not import handlers or middlewares",
		from: []string{pkg("queries")},
		mustNotImport: []string{
			pkg("handlers"),
			pkg("middlewares"),
			pkg("setup"),
		},
	},
	{
		name: "commands must not import repository implementations directly (use interfaces)",
		from: []string{pkg("commands")},
		mustNotImport: []string{
			pkg("repositories/postgres"),
			pkg("repositories/inmemory"),
		},
	},
	{
		name: "queries must not import repository implementations directly (use interfaces)",
		from: []string{pkg("queries")},
		mustNotImport: []string{
			pkg("repositories/postgres"),
			pkg("repositories/inmemory"),
		},
	},
	{
		name: "behaviors must not import handlers (behaviors are pipeline middleware, not controllers)",
		from: []string{pkg("behaviors")},
		mustNotImport: []string{
			pkg("handlers"),
			pkg("setup"),
		},
	},
	{
		name: "repository implementations must not import business logic layers",
		from: []string{
			pkg("repositories/postgres"),
			pkg("repositories/inmemory"),
		},
		mustNotImport: []string{
			pkg("handlers"),
			pkg("commands"),
			pkg("queries"),
			pkg("behaviors"),
			pkg("middlewares"),
			pkg("setup"),
			pkg("server"),
		},
	},
	{
		name: "models must not import business logic or infrastructure layers (change package is allowed)",
		from: []string{pkg("models")},
		mustNotImport: []string{
			pkg("handlers"),
			pkg("commands"),
			pkg("queries"),
			pkg("behaviors"),
			pkg("repositories"),
			pkg("middlewares"),
			pkg("authentication"),
			pkg("events"),
			pkg("email"),
			pkg("database"),
			pkg("setup"),
			pkg("server"),
		},
	},
	{
		name: "email package must not import business logic layers",
		from: []string{pkg("email")},
		mustNotImport: []string{
			pkg("handlers"),
			pkg("commands"),
			pkg("queries"),
			pkg("behaviors"),
			pkg("repositories"),
			pkg("setup"),
		},
	},
	{
		name: "authentication must not import business logic layers",
		from: []string{pkg("authentication")},
		mustNotImport: []string{
			pkg("handlers"),
			pkg("commands"),
			pkg("queries"),
			pkg("behaviors"),
			pkg("repositories"),
			pkg("setup"),
		},
	},
	{
		name: "config must not import any other internal package",
		from: []string{pkg("config")},
		mustNotImport: []string{
			pkg("handlers"),
			pkg("commands"),
			pkg("queries"),
			pkg("behaviors"),
			pkg("repositories"),
			pkg("middlewares"),
			pkg("authentication"),
			pkg("events"),
			pkg("email"),
			pkg("database"),
			pkg("models"),
			pkg("setup"),
			pkg("server"),
		},
	},
}

func TestArchitecturalConstraints(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedImports | packages.NeedDeps,
		// Load all internal packages recursively
		Tests: false,
	}

	for _, r := range rules {
		r := r
		t.Run(r.name, func(t *testing.T) {
			t.Parallel()

			pkgs, err := packages.Load(cfg, toPatterns(r.from)...)
			if err != nil {
				t.Fatalf("failed to load packages %v: %v", r.from, err)
			}

			for _, p := range pkgs {
				if packages.PrintErrors(pkgs) > 0 {
					t.Fatalf("package load errors in %s", p.PkgPath)
				}

				transitiveImports := collectTransitiveImports(p)

				for _, forbidden := range r.mustNotImport {
					for imported := range transitiveImports {
						if strings.HasPrefix(imported, forbidden) {
							t.Errorf(
								"\nARCHITECTURE VIOLATION:\n  package: %s\n  imports: %s\n  rule:    %s\n",
								p.PkgPath, imported, r.name,
							)
						}
					}
				}
			}
		})
	}
}

// TestNoCycles verifies there are no import cycles within the internal packages.
// The Go compiler will catch direct cycles, but this catches indirect ones
// that only appear when looking at the full transitive closure.
func TestNoCycles(t *testing.T) {
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedImports | packages.NeedDeps,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, module+"/internal/...")
	if err != nil {
		t.Fatalf("failed to load packages: %v", err)
	}

	// Build import graph restricted to internal packages only
	graph := make(map[string][]string)
	for _, p := range pkgs {
		if !strings.HasPrefix(p.PkgPath, module+"/internal/") {
			continue
		}
		for imp := range p.Imports {
			if strings.HasPrefix(imp, module+"/internal/") {
				graph[p.PkgPath] = append(graph[p.PkgPath], imp)
			}
		}
	}

	// DFS cycle detection
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var path []string

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		inStack[node] = true
		path = append(path, node)

		for _, neighbor := range graph[node] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if inStack[neighbor] {
				// Found cycle — report the cycle path
				cycleStart := -1
				for i, p := range path {
					if p == neighbor {
						cycleStart = i
						break
					}
				}
				cycle := append(path[cycleStart:], neighbor)
				t.Errorf("import cycle detected:\n  %s", strings.Join(cycle, "\n  → "))
				return true
			}
		}

		path = path[:len(path)-1]
		inStack[node] = false
		return false
	}

	for _, p := range pkgs {
		if !strings.HasPrefix(p.PkgPath, module+"/internal/") {
			continue
		}
		if !visited[p.PkgPath] {
			path = nil
			dfs(p.PkgPath)
		}
	}
}

// TestSetupIsCompositionRoot verifies that setup is the only package
// allowed to import both handlers and repositories/postgres together.
// If any other package does this, it is acting as a composition root,
// which violates the single-responsibility principle for wiring.
func TestSetupIsCompositionRoot(t *testing.T) {
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedImports | packages.NeedDeps,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, module+"/internal/...")
	if err != nil {
		t.Fatalf("failed to load packages: %v", err)
	}

	for _, p := range pkgs {
		if p.PkgPath == pkg("setup") {
			continue // setup is allowed to do this
		}
		if !strings.HasPrefix(p.PkgPath, module+"/internal/") {
			continue
		}

		imports := collectTransitiveImports(p)
		importsHandlers := containsPrefix(imports, pkg("handlers"))
		importsPostgres := containsPrefix(imports, pkg("repositories/postgres"))

		if importsHandlers && importsPostgres {
			t.Errorf(
				"package %s imports both handlers and repositories/postgres — only setup should do this",
				p.PkgPath,
			)
		}
	}
}

// collectTransitiveImports returns all transitive imports of a package,
// restricted to the module's own packages.
func collectTransitiveImports(p *packages.Package) map[string]bool {
	result := make(map[string]bool)
	var collect func(pkg *packages.Package)
	collect = func(pkg *packages.Package) {
		for path, imp := range pkg.Imports {
			if !result[path] {
				result[path] = true
				collect(imp)
			}
		}
	}
	collect(p)
	return result
}

func containsPrefix(imports map[string]bool, prefix string) bool {
	for imp := range imports {
		if strings.HasPrefix(imp, prefix) {
			return true
		}
	}
	return false
}

func toPatterns(paths []string) []string {
	patterns := make([]string, len(paths))
	for i, p := range paths {
		patterns[i] = p + "/..."
	}
	return patterns
}
