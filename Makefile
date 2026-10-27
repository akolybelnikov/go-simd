GOEXPERIMENT ?= simd
export GOEXPERIMENT
GO ?= go
ARGS ?=
BENCHTIME ?= 100ms
COUNT ?= 5
ARTIFACT_ROOT ?= artifacts
BIN_DIR ?= bin

.PHONY: help list run run-addition run-computation report report-addition report-computation build install test vet fmt fmt-check check clean
help:
	@printf '%s\n' 'Go SIMD laboratory' '' 'Explore:' \
	  '  make list [ARGS="-group addition"]  List operations and runtime context' \
	  '  make run [ARGS="-group addition"]   Quick comparison (fixed size)' \
	  '  make report [ARGS="-format md"]     Repeated sweep and saved evidence' \
	  '  make run-addition | run-computation [ARGS="..."]' \
	  '  make report-addition | report-computation [ARGS="..."]' '' \
	  'Report controls: COUNT=5 BENCHTIME=100ms ARTIFACT_ROOT=artifacts' \
	  'All run/list/report targets forward ARGS unchanged.' '' 'Build and install:' \
	  '  make build [BIN_DIR=bin]            Build both standalone commands' \
	  '  make install                       Install both commands with go install' '' \
	  'Develop:' '  make test | vet | fmt | fmt-check | check' \
	  '  make clean                         Remove generated reports and Go build artifacts' '' \
	  'CLI flags: make run ARGS="-h" or make report ARGS="-h"' \
	  'Direct equivalents: go run ./cmd/simdbench [flags]' \
	  '                    go run ./cmd/simdreport [flags]' \
	  'Installed equivalents: simdbench [flags], simdreport [flags]'
list:
	$(GO) run ./cmd/simdbench -list $(ARGS)
run:
	$(GO) run ./cmd/simdbench $(ARGS)
run-addition:
	$(GO) run ./cmd/simdbench -group addition $(ARGS)
run-computation:
	$(GO) run ./cmd/simdbench -group computation $(ARGS)
report:
	$(GO) run ./cmd/simdreport -count=$(COUNT) -benchtime=$(BENCHTIME) -output="$(ARTIFACT_ROOT)" $(ARGS)
report-addition:
	$(GO) run ./cmd/simdreport -group addition -count=$(COUNT) -benchtime=$(BENCHTIME) -output="$(ARTIFACT_ROOT)" $(ARGS)
report-computation:
	$(GO) run ./cmd/simdreport -group computation -count=$(COUNT) -benchtime=$(BENCHTIME) -output="$(ARTIFACT_ROOT)" $(ARGS)
build:
	$(GO) build -o "$(BIN_DIR)/" ./cmd/simdbench ./cmd/simdreport
install:
	$(GO) install ./cmd/simdbench ./cmd/simdreport
test:
	$(GO) test ./...
vet:
	$(GO) vet ./...
fmt:
	gofmt -w cmd internal
fmt-check:
	@test -z "$$(gofmt -l cmd internal)" || { gofmt -l cmd internal; exit 1; }
check: fmt-check test vet
clean:
	$(GO) clean ./...
	rm -rf -- "$(ARTIFACT_ROOT)"
