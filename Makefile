GOEXPERIMENT ?= simd
GO ?= go
ARGS ?=
ARTIFACT_DIR := packages/ops/artifacts
BENCHTIME ?= 100ms

.PHONY: help run run-addition run-computation report clean

help:
	@printf '%s\n' \
		'Usage:' \
		'  make run                         Run all CLI benchmarks' \
		'  make run ARGS="-function add"   Run a legacy benchmark group' \
		'  make run ARGS="-group addition" Run only the addition benchmarks' \
		'  make run ARGS="-group computation" Run only the computation benchmarks' \
		'  make run-addition               Run only addition benchmarks' \
		'  make run-computation            Run only computation benchmarks' \
		'  make report                     Save detailed benchmark tables' \
		'  make clean                      Remove generated benchmark artifacts'

run:
	@GOEXPERIMENT=$(GOEXPERIMENT) $(GO) run ./cmd/simdbench $(ARGS)

run-addition:
	@GOEXPERIMENT=$(GOEXPERIMENT) $(GO) run ./cmd/simdbench -addition

run-computation:
	@GOEXPERIMENT=$(GOEXPERIMENT) $(GO) run ./cmd/simdbench -computation

report:
	@printf '%s\n' 'Quick comparison (1,048,576 elements):'
	@GOEXPERIMENT=$(GOEXPERIMENT) $(GO) run ./cmd/simdbench
	@mkdir -p $(ARTIFACT_DIR)
	@report="$(ARTIFACT_DIR)/benchmark-$$(date +%Y%m%d-%H%M%S).txt"; \
	raw="$$report.raw"; \
	trap 'rm -f "$$raw"' EXIT; \
	if GOEXPERIMENT=$(GOEXPERIMENT) $(GO) test ./packages/addition ./packages/computation \
		-run='^$$' -bench='^(BenchmarkAddition|BenchmarkComputation)$$' \
		-benchmem -benchtime=$(BENCHTIME) -count=1 > "$$raw" 2>&1; then \
		awk '\
			/^(Addition|Computation) benchmark comparison/ { \
				if (found) print ""; \
				found = 1; \
				printing = 1; \
				print; \
				next; \
			} \
			printing && /^ELEMENTS[[:space:]]/ { print; next } \
			printing && /^[[:space:]]*[0-9]+[[:space:]]/ { print; next } \
			printing { printing = 0 } \
		' "$$raw" > "$$report"; \
		if [ ! -s "$$report" ]; then \
			cat "$$raw"; \
			printf 'No comparison tables found in benchmark output\n' >&2; \
			exit 1; \
		fi; \
		printf 'Comparison tables saved to %s\n' "$$report"; \
	else \
		cat "$$raw"; \
		exit 1; \
	fi

clean:
	rm -rf packages/ops/artifacts
