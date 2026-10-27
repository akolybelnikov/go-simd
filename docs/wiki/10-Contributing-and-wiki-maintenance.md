# 10 — Contributing and maintaining the guide

The Wiki is an incremental knowledge guide. Update user knowledge when behavior
ships, while keeping future ideas clearly separate from working commands.

## Where documentation lives

The editable source of these pages is
[`docs/wiki/`](https://github.com/akolybelnikov/go-simd/tree/main/docs/wiki)
in the main repository. GitHub publishes a copy through the repository's
separate Wiki Git repository. `_Sidebar.md` and `_Footer.md` supply navigation
and baseline context. `Home.md` supplies the learning path and guide history.

Prefer editing the tracked sources in a code change so reviewers can assess
behavior, examples, and documentation together. If you edit the live Wiki first,
bring those edits back into the canonical sources before a later synchronization.

## Add knowledge with each implemented feature

1. Describe one user question the feature answers, prerequisites, and scope.
2. Update the command registry from actual flag/help definitions, including
   defaults, valid values, failure cases, output behavior, and shell examples.
3. Add a small reproducible exercise and explain the expected output shape.
4. Specify arithmetic, inputs, timing boundaries, concurrency, and missing-data
   rules that affect valid conclusions.
5. Update report examples and troubleshooting where behavior changes.
6. Mark a capability implemented only after its acceptance checks pass; leave
   conditional ideas in the capability ledger until they exist.
7. Update Home's history, review date/baseline, navigation, and the changelog.
8. Verify example commands, links, and code, then publish the Wiki copy.

Measured output must include environment/context and be labeled as an
observation. Invented numbers must be labeled illustrative. Never turn one
machine's crossover point into a universal claim or a timing assertion in tests.

## Kernel contribution expectations

Start with an experiment or learning goal. A contributed kernel should include:

- A scalar reference and a clear arithmetic/correctness contract.
- Supported input/output types, lengths, aliases, and capability assumptions.
- Tests for boundaries, invalid lengths, and relevant numerical behavior.
- A declared measurement boundary and work/byte accounting.
- Strategy/operation identity and documentation explaining what the comparison
  can and cannot establish.

The current runtime registry is still duplicated between commands and developer
benchmarks; shared catalog work is the next architecture checkpoint, M2. Avoid
assuming that an importer alias table is already a shared execution catalog.

## Local validation commands

```powershell
$env:GOEXPERIMENT = 'simd'
go test ./... -count=1
go vet ./...
gofmt -l cmd internal examples
go run ./examples/basic
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1$/^(Scalar|SIMDPartial)$' -count 1 -benchtime 1x -format md
go run ./cmd/simdreport -h
```

Use the appropriate host/CI race check for concurrency changes. Core
`internal/input`, `internal/result`, and `internal/report` packages also build
without enabling the SIMD experiment on the module's selected Go toolchain.

For a documentation change, check the exact command examples and their links;
do not run destructive cleanup recipes just to prove they delete reports.

## Publish with GitHub CLI and Git

Inspect the repository and clone its Wiki:

```powershell
gh repo view akolybelnikov/go-simd --json hasWikiEnabled,viewerPermission
gh repo clone akolybelnikov/go-simd.wiki ../go-simd-wiki
git -C ../go-simd-wiki remote -v
git -C ../go-simd-wiki status --short --branch
```

Check the destination is the Wiki clone and preserve any uncommitted Wiki edits
before replacing pages. From the main repository root, copy the known page
sources, review, sign, and publish:

```powershell
Copy-Item ./docs/wiki/*.md -Destination ../go-simd-wiki/
git -C ../go-simd-wiki diff --check
git -C ../go-simd-wiki diff --stat
git -C ../go-simd-wiki add -- '*.md'
git -C ../go-simd-wiki commit -S -m 'docs(wiki): update the user knowledge guide'
git -C ../go-simd-wiki push origin HEAD
```

This copies known sources without deleting additional Wiki pages. Review renames
carefully because page filenames determine URLs. Push the Wiki's default branch
as checked out by the clone; do not assume its name matches the main source
repository's branch name.

`gh` has no built-in Wiki-page editor. It can inspect the parent repository and
clone `.wiki`; Git commits/pushes publish page files. GitHub requires an initial
page created through the website before cloning an empty Wiki. See
[GitHub's Wiki editing documentation](https://docs.github.com/en/communities/documenting-your-project-with-wikis/adding-or-editing-wiki-pages)
and [the `gh repo clone` reference](https://cli.github.com/manual/gh_repo_clone).

After pushing, open Home, follow sidebar links, inspect a table/code example,
and verify the Wiki remote points at the published commit. Wiki pushes are
separate from source-repository CI; preserve validation evidence with the source
change and check the source commit's Actions run independently.

## Documentation review checklist

- Does every project flag and Make target still match the source?
- Is shell syntax labeled, with regex quoting and dollar expansion explained?
- Do examples use registered sizes, current variant names, and the right group?
- Are timings, bytes, percentages, repetitions, and missing values explained?
- Does kernel usage respect the current `internal/` import and alias contracts?
- Are implemented capabilities distinguished from planned work?
- Can another engineer repeat the exercise and find its raw evidence?
