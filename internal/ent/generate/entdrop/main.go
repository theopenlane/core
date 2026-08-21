// Command entdrop plans and executes the removal of an ent schema.
//
// Removing a schema is painful for one reason: orphaned generated files. ent's
// cleanOldNodes only deletes the five standard per-type templates plus the
// package dir under internal/ent/generated. Every other artifact codegen emits
// per entity survives - graphql fragments, resolvers, csv samples, history
// schemas, cli commands. Those orphans keep declaring the dead type, so the next
// gqlgen run faithfully rebuilds it into ent.generated.go and that half-million
// line file looks like something to fix by hand. It is not. Delete the orphans
// and regenerate; ent.generated.go rewrites itself clean.
//
// The other half is BLOCKING: entc type-checks the entire import closure of
// internal/ent/schema before generating, so one stale reference in that closure
// stops the regeneration that would have cleaned everything else up.
//
//	go run ./internal/ent/generate/entdrop ScheduledJob JobRunner
//	go run ./internal/ent/generate/entdrop -prune-edges -delete-orphans ScheduledJob
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"time"

	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"
	"github.com/urfave/cli/v3"
)

// the schemas use these same two packages for naming, so name derivation here
// matches what codegen actually emits
var plural = pluralize.NewClient()

// timeNow is a variable so the migration stamp can be pinned in tests
var timeNow = time.Now

// trees codegen owns; anything here is rewritten or orphaned, never hand-edited
var generatedTrees = []string{
	"internal/ent/generated/",
	"internal/ent/historygenerated/",
	"internal/ent/historyschema/",
	"internal/ent/authzgenerated/",
	"internal/ent/csvgenerated/",
	"internal/ent/oscalgenerated/",
	"internal/ent/workflowgenerated/",
	"internal/ent/exportablegenerated/",
	"internal/ent/entityops/",
	"internal/graphapi/generated/",
	"internal/graphapi/historygenerated/",
	"internal/graphapi/testclient/",
	"internal/graphapi/schema/",
	"internal/graphapi/schemahistory/",
	"internal/graphapi/query/",
	"internal/graphapi/clientschema/",
	"internal/graphapi/historyschema/",
	"internal/httpserve/handlers/csv/",
	"cli/cmd/",
}

// the only files ent removes on its own, relative to internal/ent/generated
var entAutoClean = []string{"%s.go", "%s_create.go", "%s_update.go", "%s_delete.go", "%s_query.go"}

var migrationDirs = []string{"db/migrations/", "db/migrations-goose-postgres/"}

const schemaDir = "internal/ent/schema"

type buckets struct {
	blocking, orphan, autoclean, handwritten, migration []string
	// dedicated files exist only to serve the target and can be deleted wholesale
	dedicated []string
}

// preTrims must be cleaned before regenerating. the first two are generated files
// that sit inside entc's closure; the rest reference symbols defined in dedicated
// files that the pre phase deletes, so they stop compiling the moment it runs
var preTrims = []string{
	"internal/ent/entityops/entity_registry.go",
	"internal/graphapi/model/gen_models.go",
	"internal/ent/privacy/token/keys.go",
	"internal/ent/privacy/rule/allow_if_token_valid.go",
	"internal/ent/privacy/rule/modules.go",
	"internal/ent/privacy/rule/modules_test.go",
	"internal/ent/hooks/edges.go",
}

// postTrims still reference generated types, so they only become removable once
// codegen has dropped them
var postTrims = []string{
	"internal/graphapi/search_context.go",
	"internal/graphapi/models_test.go",
	"internal/graphapi/bulk.go",
	"common/enums/allenums_test.go",
	"cli/main.go",
}

// gqlgenDeadMarker begins the block gqlgen parks removed resolvers in
const gqlgenDeadMarker = "The code below was going to be deleted when updating resolvers"

var gqlgenResolvers = []string{
	"internal/graphapi/ent.resolvers.go",
	"internal/graphapi/history/ent.resolvers.go",
}

// runField removes a field from its schema and cleans up around it
func runField(root string, targets []string, phase string) error {
	names := make([]string, 0, len(targets))

	for _, t := range targets {
		entity, fld, ok := strings.Cut(t, ".")
		if !ok || fld == "" {
			return cli.Exit("--field expects Entity.field, e.g. Control.ref_code", 2)
		}

		names = append(names, fld)

		if phase != "pre" {
			continue
		}

		n, where, err := pruneSchemaField(root, entity, fld)
		if err != nil {
			return err
		}

		if n == 0 {
			return fmt.Errorf("%s: field %q not found in %s", t, fld, where)
		}

		fmt.Printf("  removed %s from %s\n", fld, where)
	}

	trims, next := preTrims, "now run: task regenerate"
	if phase == "post" {
		trims, next = postTrims, "field removal done; `task db:create` writes the DROP COLUMN migration"
	}

	for _, f := range trims {
		if _, err := trimIfPresent(root, f, names); err != nil {
			return err
		}
	}

	fmt.Println("\n" + next)

	return nil
}

// pruneSchemaField cuts the field declaration out of the entity's schema file. a
// field declared in a mixin is reported instead, since removing it there would
// change every schema that mixes it in
func pruneSchemaField(root, entity, fld string) (int, string, error) {
	rel := filepath.Join(schemaDir, strings.ReplaceAll(snake(entity), "_", "")+".go")
	path := filepath.Join(root, rel)

	if _, err := os.Stat(path); err != nil {
		return 0, rel, fmt.Errorf("no schema file at %s", rel)
	}

	declRx := regexp.MustCompile(`field\.\w+\("` + regexp.QuoteMeta(fld) + `"\)`)

	n, _, err := pruneFile(path, declRx, regexp.MustCompile(`.*`))
	if err != nil {
		return 0, rel, err
	}

	if n == 0 {
		for _, dir := range []string{schemaDir, "internal/ent/mixin"} {
			entries, _ := os.ReadDir(filepath.Join(root, dir))
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
					continue
				}

				b, _ := os.ReadFile(filepath.Join(root, dir, e.Name()))
				if declRx.Match(b) {
					return 0, filepath.Join(dir, e.Name()) + " (a mixin - remove it there by hand)", nil
				}
			}
		}
	}

	return n, rel, nil
}

// runPre does everything that must happen while the tables and types still exist
func runPre(ctx context.Context, cmd *cli.Command, root string, targets, orphans, dedicated, blocking []string) error {
	if name := cmd.String("migration"); name != "" {
		if err := writeDropMigration(ctx, root, name, targets); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
	}

	if err := pruneSchemaEdges(root, targets); err != nil {
		return fmt.Errorf("prune-edges: %w", err)
	}

	if err := removeFiles(root, append(dedupe(orphans), dedupe(dedicated)...)); err != nil {
		return err
	}

	// two passes are needed and neither subsumes the other. codegen deletes the
	// generated package for each target, so anything importing one is certain to
	// break even though it still compiles right now
	if err := trimBlocking(ctx, root, willBreak(root, blocking, targets), targets); err != nil {
		return err
	}

	// and then whatever is already broken, with the compiler naming it
	if err := trimUntilCompiles(ctx, root, targets, buildEntcRoots); err != nil {
		return err
	}

	fmt.Println("\npre phase done - now run: task regenerate")

	return nil
}

// willBreak returns the blocking files that import a generated package for one of the
// targets. codegen deletes those packages, so these files are certain to break even
// though they compile now - the compiler cannot warn about it until it is too late,
// which is why trimUntilCompiles alone is not enough
func willBreak(root string, files, targets []string) []string {
	// codegen deletes the per-entity subpackage and the entity's types in the root
	// generated package, so a file is doomed if it imports either and names a target
	const rootPkg = `internal/ent/generated"`

	var subPkgs [][]byte

	names := make([][]byte, 0, len(targets))

	for _, t := range targets {
		subPkgs = append(subPkgs, []byte("internal/ent/generated/"+strings.ReplaceAll(snake(t), "_", "")+`"`))
		names = append(names, []byte(t))
	}

	var out []string

	for _, rel := range dedupe(files) {
		// entc rewrites these trees wholesale and they stay self consistent:
		// client.go importing generated/jobrunner is fine, codegen rewrites both
		if hasPrefixAny(rel, entcTargets) {
			continue
		}

		src, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue
		}

		if containsAny(src, subPkgs) ||
			(bytes.Contains(src, []byte(rootPkg)) && containsAny(src, names)) {
			out = append(out, rel)
		}
	}

	return out
}

func containsAny(src []byte, needles [][]byte) bool {
	for _, n := range needles {
		if bytes.Contains(src, n) {
			return true
		}
	}

	return false
}

// entcRoots are the two packages entc loads; everything they import has to compile
// before any generation happens
var entcRoots = []string{"./internal/ent/schema/", "./internal/ent/historyschema/"}

// buildErrFile pulls the filename out of a go diagnostic. vet prefixes the first
// type error with "vet: ", build does not, so the prefix is optional
var buildErrFile = regexp.MustCompile(`(?m)^(?:vet: )?(\S+\.go):\d+:\d+:`)

// trimUntilCompiles uses the compiler as the oracle instead of guessing which files
// a removal breaks. build the entc roots, trim whatever the errors name, repeat.
// heuristics kept missing cases in both directions - a file referencing a deleted
// symbol from its own package has no import to scan for, and a file importing a
// deleted package may be fine because codegen rewrites it
func trimUntilCompiles(ctx context.Context, root string, targets []string, check checker) error {
	const maxRounds = 15

	var previous string

	for round := 1; round <= maxRounds; round++ {
		out, ok := check(ctx, root)
		if ok {
			return nil
		}

		var named []string
		for _, m := range buildErrFile.FindAllStringSubmatch(out, -1) {
			named = append(named, m[1])
		}

		files := dedupe(named)
		if len(files) == 0 {
			return fmt.Errorf("entc closure does not compile and no file was named:\n%s", out)
		}

		// the same failures twice running means trimming cannot resolve them
		current := strings.Join(files, "\n")
		if current == previous {
			return fmt.Errorf("trimming did not resolve these:\n%s", out)
		}

		previous = current

		var toDelete []string

		for _, rel := range files {
			hollowed, err := trimIfPresent(root, rel, targets)
			if err != nil {
				return err
			}

			if hollowed {
				toDelete = append(toDelete, rel)
			}
		}

		if err := removeFiles(root, toDelete); err != nil {
			return err
		}

		fmt.Printf("  round %d: addressed %d file(s)\n", round, len(files))
	}

	return fmt.Errorf("entc closure still does not compile after %d rounds", maxRounds)
}

// checker reports whether the tree passes some compilation gate, with its output
type checker func(ctx context.Context, root string) (string, bool)

// buildEntcRoots reports whether the packages entc loads compile, with the output.
// go build ignores _test.go, which is correct here: entc does not load test files
// either, so they cannot block generation
func buildEntcRoots(ctx context.Context, root string) (string, bool) {
	args := append([]string{"build", "-tags=codegen"}, entcRoots...)

	return runGo(ctx, root, args...)
}

// vetWithTests type-checks the whole module including _test.go. test files break the
// final build rather than generation, so this is the gate for the post phase. the
// test tag is what this repo builds its test helpers under
func vetWithTests(ctx context.Context, root string) (string, bool) {
	return runGo(ctx, root, "vet", "-tags=test", "./...")
}

func runGo(ctx context.Context, root string, args ...string) (string, bool) {
	c := exec.CommandContext(ctx, "go", args...)
	c.Dir = root

	out, err := c.CombinedOutput()

	return string(out), err == nil
}

// runPost cleans up what only becomes removable once codegen has run
func runPost(ctx context.Context, root string, targets, dedicated []string) error {
	if err := trimBlocking(ctx, root, postTrims, targets); err != nil {
		return err
	}

	// gqlgen parks removed resolvers in a comment block instead of deleting them
	for _, rel := range gqlgenResolvers {
		if err := stripDeadResolvers(filepath.Join(root, rel)); err != nil {
			return err
		}
	}

	// enums and other dedicated files are only unreferenced now that codegen has run
	if err := removeFiles(root, dedupe(dedicated)); err != nil {
		return err
	}

	// whatever is left, including _test.go, which no earlier gate can see: pre builds
	// without the test tag because entc does not load test files either
	if err := trimUntilCompiles(ctx, root, targets, vetWithTests); err != nil {
		return err
	}

	fmt.Println("\npost phase done - remaining by hand: fga/ model and roles")

	return nil
}

// trimIfPresent trims a file, reporting whether it was left hollow
func trimIfPresent(root, rel string, targets []string) (hollowed bool, err error) {
	path := filepath.Join(root, rel)
	if _, err := os.Stat(path); err != nil {
		return false, nil
	}

	n, err := trimGenerated(path, targets)

	switch {
	case errors.Is(err, errHollow):
		return true, nil
	case err != nil:
		return false, fmt.Errorf("trim %s: %w", rel, err)
	}

	if n > 0 {
		fmt.Printf("  trimmed %d declaration(s) from %s\n", n, rel)
	}

	return false, nil
}

// trimBlocking trims every file that still references the targets and deletes the
// ones that existed only for them. driving this off the blocking set rather than a
// fixed list is what catches a file named for a concept, like hooks/job.go, which no
// filename rule would match
func trimBlocking(ctx context.Context, root string, files, targets []string) error {
	var toDelete []string

	for _, rel := range dedupe(files) {
		hollowed, err := trimIfPresent(root, rel, targets)
		if err != nil {
			return err
		}

		if hollowed {
			toDelete = append(toDelete, rel)
		}
	}

	if len(toDelete) > 0 {
		fmt.Printf("  %d file(s) had nothing left after trimming\n", len(toDelete))
	}

	return removeFiles(root, toDelete)
}

// stripDeadResolvers removes the block gqlgen appends when a resolver disappears
func stripDeadResolvers(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(src), "\n")

	for i, l := range lines {
		if !strings.Contains(l, gqlgenDeadMarker) {
			continue
		}

		for i > 0 && strings.TrimSpace(lines[i-1]) == "" {
			i--
		}

		out := strings.TrimRight(strings.Join(lines[:i], "\n"), "\n") + "\n"
		fmt.Printf("  stripped dead resolver block from %s\n", filepath.Base(path))

		return os.WriteFile(path, []byte(out), 0o644)
	}

	return nil
}

// removeFiles deletes paths from the working tree only. it deliberately does not use
// git rm: staging is the caller's decision, not this tool's
func removeFiles(root string, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	for _, rel := range paths {
		if err := os.RemoveAll(filepath.Join(root, rel)); err != nil {
			return fmt.Errorf("remove %s: %w", rel, err)
		}
	}

	fmt.Printf("  removed %d files\n", len(paths))

	return nil
}

func main() {
	if err := entdropApp().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func entdropApp() *cli.Command {
	return &cli.Command{
		Name:      "entdrop",
		Usage:     "plan and execute the removal of an ent schema without hand-editing generated code",
		ArgsUsage: "Entity [Entity...]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "prune-edges",
				Usage: "remove edge definitions referencing the targets from sibling schema files",
			},
			&cli.BoolFlag{
				Name:  "delete-orphans",
				Usage: "git rm generated files codegen will not clean up itself",
			},
			&cli.StringFlag{
				Name:  "phase",
				Usage: "pre (before regeneration) or post (after); bundles the flags each stage needs",
			},
			&cli.BoolFlag{
				Name:  "field",
				Usage: "targets are Entity.field; no orphans or DROP TABLE apply, ent emits DROP COLUMN itself",
			},
			&cli.BoolFlag{
				Name:  "delete-dedicated",
				Usage: "git rm hand-written files that exist only to serve the targets",
			},
			&cli.StringFlag{
				Name:  "migration",
				Usage: "write a DROP TABLE migration with this name; must run BEFORE the schemas are deleted, while the tables are still in migrate/schema.go",
			},
			&cli.StringSliceFlag{
				Name:  "trim",
				Usage: "generated file inside entc's closure to strip of references; it must compile before codegen can rewrite it (repeatable)",
			},
		},
		Action: run,
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	targets := cmd.Args().Slice()
	if len(targets) == 0 {
		return cli.Exit("at least one entity name is required", 2)
	}

	root := gitRoot()
	closure := entcClosure(root)
	tracked := gitFiles(root)

	phase := cmd.String("phase")

	// a field owns no files of its own, so orphan sweeping and DROP TABLE do not
	// apply; ent emits DROP COLUMN by itself on the next `task db:create`
	if cmd.Bool("field") {
		return runField(root, targets, cmd.String("phase"))
	}

	var allOrphans, allDedicated, allBlocking []string

	for _, target := range targets {
		b := classify(root, target, tracked, closure)
		fmt.Printf("\n%s\n%s\n%s\n", strings.Repeat("=", 72), target, strings.Repeat("=", 72))
		report("BLOCKING", "fix FIRST - entc type-checks these, they stop regeneration", b.blocking)
		report("ORPHAN", "DELETE - codegen will not remove these, and they resurrect the type", b.orphan)
		report("AUTOCLEAN", "codegen removes these itself on the next run - leave them", b.autoclean)
		report("HANDWRITTEN", "delete or edit any time before the final build", b.handwritten)
		report("MIGRATION", "do NOT edit - add a new DROP TABLE migration instead", b.migration)
		allOrphans = append(allOrphans, b.orphan...)
		allDedicated = append(allDedicated, b.dedicated...)
		allBlocking = append(allBlocking, b.blocking...)
	}

	// the pre phase runs while the tables and types still exist; post runs once
	// codegen has removed the generated references
	switch phase {
	case "pre":
		if err := runPre(ctx, cmd, root, targets, allOrphans, allDedicated, allBlocking); err != nil {
			return err
		}

		return nil
	case "post":
		if err := runPost(ctx, root, targets, allDedicated); err != nil {
			return err
		}

		return nil
	case "":
	default:
		return cli.Exit("--phase must be pre or post", 2)
	}

	if name := cmd.String("migration"); name != "" {
		if err := writeDropMigration(ctx, root, name, targets); err != nil {
			return fmt.Errorf("migration: %w", err)
		}
	}

	for _, f := range cmd.StringSlice("trim") {
		if f = strings.TrimSpace(f); f == "" {
			continue
		}
		n, err := trimGenerated(filepath.Join(root, f), targets)
		if err != nil {
			return fmt.Errorf("trim %s: %w", f, err)
		}
		fmt.Printf("  trimmed %d declaration(s) from %s\n", n, f)
	}

	if cmd.Bool("prune-edges") {
		if err := pruneSchemaEdges(root, targets); err != nil {
			return fmt.Errorf("prune-edges: %w", err)
		}
	}

	switch {
	case cmd.Bool("delete-orphans") && len(allOrphans) > 0:
		args := append([]string{"rm", "-q", "--"}, dedupe(allOrphans)...)
		c := exec.CommandContext(ctx, "git", args...)
		c.Dir = root
		if out, err := c.CombinedOutput(); err != nil {
			return fmt.Errorf("git rm: %w: %s", err, out)
		}
		fmt.Printf("\nremoved %d orphaned files\n", len(dedupe(allOrphans)))
	case len(allOrphans) > 0:
		fmt.Printf("\n%d orphans - re-run with --delete-orphans to remove them\n", len(dedupe(allOrphans)))
	}

	fmt.Println("\nthen: task generate:ent:smart && task generate:graphql:smart")

	return nil
}

func report(name, blurb string, items []string) {
	if len(items) == 0 {
		return
	}
	sort.Strings(items)
	fmt.Printf("\n%s (%d) - %s\n", name, len(items), blurb)
	for _, f := range items {
		fmt.Println("  " + f)
	}
}

// classify sorts every tracked file that mentions the entity into an action bucket
func classify(root, entity string, tracked []string, closure map[string]bool) buckets {
	var b buckets
	st := stems(entity)
	// substring, not word bounded: InterceptorJobRunnerRegistrationToken contains
	// JobRunner with no boundary around it, so \b would miss the file entirely
	rx := regexp.MustCompile(strings.Join([]string{
		regexp.QuoteMeta(entity),
		regexp.QuoteMeta(snake(entity)),
		regexp.QuoteMeta(strings.ReplaceAll(snake(entity), "_", "")),
		regexp.QuoteMeta(plural.Plural(entity)),
		regexp.QuoteMeta(strcase.LowerCamelCase(plural.Plural(entity))),
	}, "|"))

	for _, rel := range tracked {
		full := filepath.Join(root, rel)
		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			continue
		}

		stem := normStem(strings.SplitN(filepath.Base(rel), ".", 2)[0])
		dedicated := st[stem]
		if !dedicated {
			for _, seg := range strings.Split(filepath.Dir(rel), string(filepath.Separator)) {
				if st[normStem(seg)] {
					dedicated = true
					break
				}
			}
		}

		inTree := hasPrefixAny(rel, generatedTrees)
		pkg := filepath.Dir(rel)

		switch {
		case dedicated && inTree:
			if entHandles(rel, st) {
				b.autoclean = append(b.autoclean, rel)
			} else {
				b.orphan = append(b.orphan, rel)
			}
		case dedicated && closure[pkg]:
			b.blocking = append(b.blocking, rel)
			b.dedicated = append(b.dedicated, rel)
		case dedicated:
			b.handwritten = append(b.handwritten, rel)
			b.dedicated = append(b.dedicated, rel)
		default:
			data, err := os.ReadFile(full)
			if err != nil || !rx.Match(data) {
				continue
			}
			switch {
			case hasPrefixAny(rel, migrationDirs):
				b.migration = append(b.migration, rel)
			case closure[pkg]:
				// generated or not, anything in the closure must compile before
				// codegen can run and rewrite it
				b.blocking = append(b.blocking, rel)
			case inTree:
				// shared generated file outside the closure: codegen rewrites it
			default:
				b.handwritten = append(b.handwritten, rel)
			}
		}
	}
	return b
}

// entcTargets are the directories ent generates into. cleanOldNodes runs per
// target, so per-type files under either one are removed by codegen itself
var entcTargets = []string{
	"internal/ent/generated/",
	"internal/ent/historygenerated/",
}

// entHandles reports whether ent's own cleanOldNodes will delete this file
func entHandles(rel string, st map[string]bool) bool {
	base := ""

	for _, t := range entcTargets {
		if strings.HasPrefix(rel, t) {
			base = t
			break
		}
	}

	if base == "" {
		return false
	}

	tail := strings.TrimPrefix(rel, base)
	if i := strings.Index(tail, "/"); i >= 0 {
		return st[normStem(tail[:i])] // the per-type package dir
	}
	for s := range st {
		for _, pat := range entAutoClean {
			if normStem(strings.TrimSuffix(tail, ".go")) == normStem(strings.TrimSuffix(fmt.Sprintf(pat, s), ".go")) {
				return true
			}
		}
	}
	return false
}

// pruneSchemaEdges removes edge definitions that reference the targets from every
// other schema file. only elements that are calls to an edge helper are touched, so
// an unrelated construct that merely mentions the type (a mixin parent, say) is
// reported instead of silently deleted
func pruneSchemaEdges(root string, targets []string) error {
	alt := make([]string, len(targets))
	for i, t := range targets {
		alt[i] = regexp.QuoteMeta(t)
	}
	typeRx := regexp.MustCompile(`\b(` + strings.Join(alt, "|") + `)\b`)
	edgeFn := regexp.MustCompile(`(?i)edge`)

	dir := filepath.Join(root, schemaDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	doomedFile := map[string]bool{}
	for _, t := range targets {
		for s := range stems(t) {
			doomedFile[s] = true
		}
	}

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || doomedFile[normStem(strings.TrimSuffix(name, ".go"))] {
			continue
		}
		path := filepath.Join(dir, name)
		changed, skipped, err := pruneFile(path, typeRx, edgeFn)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if changed > 0 {
			fmt.Printf("  pruned %d edge(s) from %s/%s\n", changed, schemaDir, name)
		}
		for _, s := range skipped {
			fmt.Printf("  MANUAL  %s/%s:%s\n", schemaDir, name, s)
		}
	}
	return nil
}

// pruneFile excises whole composite-literal elements by byte range, which leaves
// the rest of the file untouched rather than reprinting it
func pruneFile(path string, typeRx, edgeFn *regexp.Regexp) (int, []string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return 0, nil, err
	}

	type span struct{ start, end int }
	var cuts []span
	var manual []string

	ast.Inspect(file, func(n ast.Node) bool {
		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		for _, elt := range lit.Elts {
			s := fset.Position(elt.Pos()).Offset
			e := fset.Position(elt.End()).Offset
			if s < 0 || e > len(src) || !typeRx.Match(src[s:e]) {
				continue
			}
			call, isCall := elt.(*ast.CallExpr)
			if !isCall || !edgeFn.MatchString(exprName(call.Fun)) {
				manual = append(manual, fmt.Sprintf("%d: %s", fset.Position(elt.Pos()).Line, firstLine(src[s:e])))
				continue
			}
			cs, ce := spanFor(src, s, e)
			cuts = append(cuts, span{cs, ce})
		}
		return true
	})

	// an element nested inside one already being cut is not a manual case
	kept := manual[:0]
	for _, m := range manual {
		var off int
		fmt.Sscanf(m, "%d:", &off)
		inside := false
		for _, c := range cuts {
			if off >= lineOf(src, c.start) && off <= lineOf(src, c.end-1) {
				inside = true
				break
			}
		}
		if !inside {
			kept = append(kept, m)
		}
	}
	manual = kept

	if len(cuts) == 0 {
		return 0, manual, nil
	}
	sort.Slice(cuts, func(i, j int) bool { return cuts[i].start > cuts[j].start })

	out := src
	for _, c := range cuts {
		out = append(out[:c.start:c.start], out[c.end:]...)
	}
	formatted, err := format.Source(out)
	if err != nil {
		return 0, manual, fmt.Errorf("result does not parse: %w", err)
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return 0, manual, err
	}
	return len(cuts), manual, nil
}

// lineOf returns the 1-based line number for a byte offset
func lineOf(src []byte, off int) int {
	if off < 0 {
		off = 0
	}
	if off > len(src) {
		off = len(src)
	}
	return bytes.Count(src[:off], []byte{'\n'}) + 1
}

// migrateSchemas are the generated files that hold the table definitions, read
// before the schemas are deleted
var migrateSchemas = []string{
	"internal/ent/generated/migrate/schema.go",
	"internal/ent/historygenerated/migrate/schema.go",
}

var tableDeclRx = regexp.MustCompile(`(?s)(\w+)Table\s*=\s*&schema\.Table\{(.*?)\n\t\}`)
var tableNameRx = regexp.MustCompile(`Name:\s*"([a-z_0-9]+)"`)

// dropTables finds every table belonging to the targets, including join tables and
// history tables. ent cannot emit DROP TABLE at all: its Atlas differ filters planned
// changes down to AddTable and ModifyTable, so a table removal never reaches the
// migration file no matter which options are set
func dropTables(root string, targets []string) ([]string, error) {
	want := map[string]bool{}
	prefixes := []string{}

	for _, t := range targets {
		sn := snake(t)
		want[snake(plural.Plural(t))] = true
		want[sn+"_history"] = true
		prefixes = append(prefixes, sn+"_")
	}

	var found []string

	for _, rel := range migrateSchemas {
		src, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue // the history package may legitimately not exist
		}

		for _, m := range tableDeclRx.FindAllSubmatch(src, -1) {
			nm := tableNameRx.FindSubmatch(m[2])
			if nm == nil {
				continue
			}

			name := string(nm[1])
			if want[name] {
				found = append(found, name)
				continue
			}
			// join tables, e.g. scheduled_job_controls
			for _, p := range prefixes {
				if strings.HasPrefix(name, p) {
					found = append(found, name)
					break
				}
			}
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("no tables found; run this before deleting the schemas")
	}

	return dedupe(found), nil
}

// writeDropMigration emits the atlas and goose migration files and re-hashes both dirs
func writeDropMigration(ctx context.Context, root, name string, targets []string) error {
	tables, err := dropTables(root, targets)
	if err != nil {
		return err
	}

	var atlasBody, gooseBody strings.Builder

	gooseBody.WriteString("-- +goose Up\n")

	for _, t := range tables {
		fmt.Fprintf(&atlasBody, "-- Drop %q table\nDROP TABLE IF EXISTS %q CASCADE;\n", t, t)
		fmt.Fprintf(&gooseBody, "-- drop %q table\nDROP TABLE IF EXISTS %q CASCADE;\n", t, t)
	}

	// the tables cannot be recreated from here, so down is deliberately a no-op
	gooseBody.WriteString("\n-- +goose Down\n-- irreversible: the dropped tables cannot be restored from this migration\n")

	stamp := timeNow().Format("20060102150405")
	files := map[string]string{
		filepath.Join(root, "db", "migrations", stamp+"_"+name+".sql"):                atlasBody.String(),
		filepath.Join(root, "db", "migrations-goose-postgres", stamp+"_"+name+".sql"): gooseBody.String(),
	}

	for path, body := range files {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return err
		}
		fmt.Printf("  wrote %s\n", strings.TrimPrefix(path, root+"/"))
	}

	// atlas keeps a checksum per directory; a hand-written file has to be hashed in.
	// this needs a valid atlas login, so a failure here is reported rather than fatal:
	// the migration files themselves are already correct
	hashed := true

	for _, dir := range []string{"migrations", "migrations-goose-postgres"} {
		c := exec.CommandContext(ctx, "atlas", "migrate", "hash", "--dir", "file://"+dir)
		c.Dir = filepath.Join(root, "db")

		if out, err := c.CombinedOutput(); err != nil {
			fmt.Printf("  WARNING could not hash %s: %s", dir, out)

			hashed = false
		}
	}

	if !hashed {
		fmt.Println("  atlas.sum is stale - run `task db:resethash` once atlas is authenticated")
	}

	fmt.Printf("  %d tables dropped; review before applying\n", len(tables))
	fmt.Println("  leftover columns on surviving tables are removed by the next `task db:create`")

	return nil
}

// trimGenerated strips every top-level declaration, statement, and literal element
// that mentions one of the targets. it exists for generated files that sit inside
// entc's import closure: codegen would rewrite them correctly, but they have to
// compile before codegen can run at all. the file is regenerated afterwards, so an
// over-broad cut here costs nothing
func trimGenerated(path string, targets []string) (int, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var alt []string
	for _, t := range targets {
		// every spelling codegen and imports use: SchemaJobResult,
		// jobRunnerTokenContextKey, generated/jobrunner, job_runner
		alt = append(alt,
			regexp.QuoteMeta(t),
			regexp.QuoteMeta(strcase.LowerCamelCase(t)),
			regexp.QuoteMeta(snake(t)),
			regexp.QuoteMeta(strings.ReplaceAll(snake(t), "_", "")),
		)
	}
	// any identifier containing a target name
	rx := regexp.MustCompile(`\b\w*(` + strings.Join(alt, "|") + `)\w*\b`)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return 0, err
	}

	type span struct{ start, end int }
	var cuts []span
	cut := func(n ast.Node) {
		s := fset.Position(n.Pos()).Offset
		e := fset.Position(n.End()).Offset

		if s < 0 || e > len(src) || !rx.Match(src[s:e]) {
			return
		}

		// take the doc comment with the declaration, otherwise it is orphaned
		if doc := docOf(n); doc != nil {
			if ds := fset.Position(doc.Pos()).Offset; ds >= 0 && ds < s {
				s = ds
			}
		}

		cs, ce := spanFor(src, s, e)
		cuts = append(cuts, span{cs, ce})
	}

	// only cut a declaration when the declared name itself matches. a var whose
	// value merely lists doomed entries (allSchemas) must survive; its elements are
	// removed by the composite-literal pass below
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				// imports go by path; everything else by declared name
				_, isImport := spec.(*ast.ImportSpec)
				if !isImport && !declNameMatches(spec, rx) {
					continue
				}

				// without parens the spec cannot be removed on its own: cutting it
				// leaves a bare `import`, `type`, `var` or `const` keyword and an
				// unparsable file, so the whole declaration goes
				if d.Lparen.IsValid() {
					cut(spec)
				} else {
					cut(d)
				}
			}
		case *ast.FuncDecl:
			// a whole func goes only when its signature mentions a target, e.g. a
			// stale gqlgen resolver returning *generated.JobResultConnection. a body
			// that merely mentions one is left for the statement pass
			sigEnd := d.Type.End()
			ss := fset.Position(d.Type.Pos()).Offset
			se := fset.Position(sigEnd).Offset
			if ss >= 0 && se <= len(src) && rx.Match(src[ss:se]) {
				cut(d)
				continue
			}

			if d.Body == nil {
				continue
			}
			for _, stmt := range d.Body.List {
				if es, ok := stmt.(*ast.ExprStmt); ok {
					// e.g. t.Run("bypass with JobRunnerRegistrationToken", ...)
					cut(es)
					continue
				}

				if as, ok := stmt.(*ast.AssignStmt); ok {
					// SchemaJobResult.Fields = ... : keyed off the assignment target
					s := fset.Position(as.Lhs[0].Pos()).Offset
					e := fset.Position(as.Lhs[len(as.Lhs)-1].End()).Offset
					if s >= 0 && e <= len(src) && rx.Match(src[s:e]) {
						cut(stmt)
					}
				}
			}
		}
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.CompositeLit:
			for _, elt := range v.Elts {
				cut(elt)
			}
		case *ast.StructType:
			// a surviving struct can still carry a field typed by a doomed entity
			if v.Fields != nil {
				for _, f := range v.Fields.List {
					cut(f)
				}
			}
		case *ast.CaseClause:
			// a whole switch arm, keyed off the case expressions only; a surviving
			// arm whose body happens to mention a target must not be cut
			for _, e := range v.List {
				es := fset.Position(e.Pos()).Offset
				ee := fset.Position(e.End()).Offset
				if es >= 0 && ee <= len(src) && rx.Match(src[es:ee]) {
					cut(v)
					break
				}
			}
		case *ast.InterfaceType:
			if v.Methods != nil {
				for _, m := range v.Methods.List {
					cut(m)
				}
			}
		}
		return true
	})

	if len(cuts) == 0 {
		return 0, nil
	}
	// drop spans contained in a larger span so nested nodes are not cut twice
	sort.Slice(cuts, func(i, j int) bool {
		if cuts[i].start != cuts[j].start {
			return cuts[i].start < cuts[j].start
		}
		return cuts[i].end > cuts[j].end
	})
	var merged []span
	for _, c := range cuts {
		if len(merged) > 0 && c.start < merged[len(merged)-1].end {
			continue
		}
		merged = append(merged, c)
	}

	out := src
	for i := len(merged) - 1; i >= 0; i-- {
		c := merged[i]
		out = append(out[:c.start:c.start], out[c.end:]...)
	}
	formatted, err := format.Source(out)
	if err != nil {
		return 0, fmt.Errorf("result does not parse: %w", err)
	}

	if hollow(formatted) {
		// nothing but the package clause and imports survived, so the file existed
		// only to serve the targets
		return len(merged), errHollow
	}

	return len(merged), os.WriteFile(path, formatted, 0o644)
}

// errHollow signals that trimming emptied a file out entirely
var errHollow = errors.New("file has no declarations left")

// hollow reports whether a file has no top-level declaration other than imports
func hollow(src []byte) bool {
	f, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		return false
	}

	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok || g.Tok != token.IMPORT {
			return false
		}
	}

	return true
}

// docOf returns the doc comment attached to a declaration, if any
func docOf(n ast.Node) *ast.CommentGroup {
	switch v := n.(type) {
	case *ast.FuncDecl:
		return v.Doc
	case *ast.GenDecl:
		return v.Doc
	case *ast.TypeSpec:
		return v.Doc
	case *ast.ValueSpec:
		return v.Doc
	case *ast.Field:
		return v.Doc
	}

	return nil
}

// declNameMatches reports whether a spec declares a name matching the pattern
func declNameMatches(spec ast.Spec, rx *regexp.Regexp) bool {
	switch v := spec.(type) {
	case *ast.TypeSpec:
		return rx.MatchString(v.Name.Name)
	case *ast.ValueSpec:
		for _, n := range v.Names {
			if rx.MatchString(n.Name) {
				return true
			}
		}
	}
	return false
}

func exprName(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return v.Sel.Name
	case *ast.IndexExpr:
		return exprName(v.X)
	}
	return ""
}

// spanFor widens an element's byte range to swallow its indentation, trailing comma
// and newline, but only where those are the element's alone. generated files are
// sometimes packed several elements to a line, as in `}, "procedures": {`, and
// expanding blindly to line boundaries there destroys the neighbours
func spanFor(src []byte, s, e int) (int, int) {
	if ls := lineStart(src, s); strings.TrimSpace(string(src[ls:s])) == "" {
		s = ls
	}

	i := e
	for i < len(src) && (src[i] == ' ' || src[i] == '\t') {
		i++
	}

	if i < len(src) && src[i] == ',' {
		i++
	}

	j := i
	for j < len(src) && src[j] != '\n' {
		j++
	}

	// only take the newline when nothing else follows on the line
	if j < len(src) && strings.TrimSpace(string(src[i:j])) == "" {
		i = j + 1
	}

	return s, i
}

func lineStart(src []byte, off int) int {
	if i := bytes.LastIndexByte(src[:off], '\n'); i >= 0 {
		return i + 1
	}
	return 0
}

func firstLine(b []byte) string {
	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		return strings.TrimSpace(string(b[:i]))
	}
	return strings.TrimSpace(string(b))
}

// stems returns the file stems that belong to this entity, keyed with separators
// stripped so jobrunner_token.go and job_runner_token.go both match JobRunnerToken.
// matching stays exact rather than prefix-based so ScheduledJob never claims
// ScheduledJobRun's files
func stems(entity string) map[string]bool {
	flat := strings.ReplaceAll(snake(entity), "_", "")
	pl := strings.ReplaceAll(snake(plural.Plural(entity)), "_", "")
	out := map[string]bool{}
	for _, s := range []string{flat, pl, flat + "history", pl + "history", "sample" + flat, "sample" + pl} {
		out[s] = true
	}
	return out
}

// normStem strips separators so a stem can be compared against stems()
func normStem(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), "_", "")
}

func snake(s string) string { return strcase.SnakeCase(s) }

func hasPrefixAny(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func gitRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "not a git repo:", err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(out))
}

func gitFiles(root string) []string {
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "git ls-files:", err)
		os.Exit(1)
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n")
}

// entcClosure returns the local packages entc must type-check before generating.
// go list takes a few seconds, so the result is cached in the temp dir rather
// than in the repo
func entcClosure(root string) map[string]bool {
	cache := filepath.Join(os.TempDir(), "entdrop-closure.txt")
	if b, err := os.ReadFile(cache); err == nil && len(b) > 0 {
		out := map[string]bool{}
		for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
			out[l] = true
		}
		return out
	}
	pkgs := map[string]bool{}
	const mod = "github.com/theopenlane/core/"
	for _, target := range []string{"./internal/ent/schema", "./internal/ent/historyschema"} {
		cmd := exec.Command("go", "list", "-deps", "-tags=codegen", target)
		cmd.Dir = root
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, mod) {
				pkgs[strings.TrimPrefix(line, mod)] = true
			}
		}
	}
	keys := make([]string, 0, len(pkgs))
	for k := range pkgs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	_ = os.WriteFile(cache, []byte(strings.Join(keys, "\n")), 0o644)
	return pkgs
}
