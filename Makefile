GOEXPERIMENT ?= simd
export GOEXPERIMENT
GO ?= go
ARGS ?=
BENCHTIME ?= 100ms
COUNT ?= 5
ARTIFACT_ROOT ?= artifacts

.PHONY: help run run-addition run-computation report report-addition report-computation test vet fmt fmt-check check clean
help:
	@printf '%s\n' 'make run [ARGS="-group addition"]' 'make report [COUNT=5 BENCHTIME=100ms ARGS="-format md"]' 'make report-addition | report-computation' 'make check | fmt | clean'
run:
	$(GO) run ./cmd/simdbench $(ARGS)
run-addition:
	$(GO) run ./cmd/simdbench -group addition
run-computation:
	$(GO) run ./cmd/simdbench -group computation
report:
	$(GO) run ./cmd/simdreport -count=$(COUNT) -benchtime=$(BENCHTIME) -output="$(ARTIFACT_ROOT)" $(ARGS)
report-addition:
	$(GO) run ./cmd/simdreport -group addition -count=$(COUNT) -benchtime=$(BENCHTIME) -output="$(ARTIFACT_ROOT)" $(ARGS)
report-computation:
	$(GO) run ./cmd/simdreport -group computation -count=$(COUNT) -benchtime=$(BENCHTIME) -output="$(ARTIFACT_ROOT)" $(ARGS)
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
