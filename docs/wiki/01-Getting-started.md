# 01 — Getting started

Goal: obtain the checkout, enable the experimental toolchain feature, and verify
that your environment can run the lab. No account, cloud service, or GUI is
required for the laboratory commands.

## 1. Requirements

- The module declares Go **1.27.3**. The selected toolchain must also provide the
  experimental `simd` package and recognize `GOEXPERIMENT=simd`.
- Git, or GitHub CLI (`gh`), to obtain the source.
- A shell such as PowerShell or Bash.
- Optional: GNU Make and a POSIX shell for the convenience Makefile.

Windows/amd64 and Linux/amd64 are exercised by the current CI setup. Other
targets need actual execution before assuming support. Cross-compilation alone
does not prove SIMD behavior or performance. Experimental APIs can change with
the toolchain; record the version that actually built the program.

## 2. Obtain the source

With GitHub CLI:

```powershell
gh repo clone akolybelnikov/go-simd
cd go-simd
```

Or with Git:

```powershell
git clone https://github.com/akolybelnikov/go-simd.git
cd go-simd
```

All `go run ./...` and report commands in this guide assume this repository-root
working directory unless a section says otherwise.

## 3. Enable the experiment in your current shell

PowerShell:

```powershell
$env:GOEXPERIMENT = 'simd'
go version
go env -json GOEXPERIMENT GOAMD64 GOARM64 GOTOOLCHAIN
```

Bash or another POSIX shell:

```bash
export GOEXPERIMENT=simd
go version
go env -json GOEXPERIMENT GOAMD64 GOARM64 GOTOOLCHAIN
```

These commands configure the current process environment and its child commands.
Set the variable again in a new terminal. Make exports `GOEXPERIMENT=simd` for
its targets unless you override it.

## 4. Verify the checkout

```powershell
go test ./...
go vet ./...
gofmt -l cmd internal examples
go run ./cmd/simdbench -h
go run ./cmd/simdbench -list
go run ./cmd/simdreport -h
```

`gofmt -l` lists files needing formatting; no output means those files are
formatted. Help is written to stderr and exits successfully without measuring.
The test suite includes short real report exports; it is not only kernel tests.

`-list` shows this process's Go/platform, vector width/lanes, emulation flag and
GOMAXPROCS, followed by the implemented operations and strategies. It allocates
no benchmark fixtures and runs no measurements. Filter with `-group addition`
or `-group computation`; only the existing six kernels are registered.

Use `go test -race ./...` where the host C toolchain supports race builds. Linux
CI runs this check. A local `runtime/cgo` build failure is a toolchain problem
that can occur before the project's tests execute; see troubleshooting.

## 5. Run the small code example

```powershell
go run ./examples/basic
```

This calls all six kernels and demonstrates in-place output, unchanged output
on length mismatch, and empty input. It prints values, not benchmark claims.

## 6. Check report generation

```powershell
go run ./cmd/simdreport -group addition -bench '^BenchmarkAddition$/^n=1$/^(Scalar|SIMDPartial)$' -count 1 -benchtime 1x
```

Expect a readable terminal table and two paths: a `.txt` summary and matching
`.raw` observations under `artifacts/addition/`. Timing values vary by machine.
You have completed setup when the command succeeds and both files are present.

Continue to [your first experiment](https://github.com/akolybelnikov/go-simd/wiki/02-First-experiment).
