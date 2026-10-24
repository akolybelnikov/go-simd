GOEXPERIMENT ?= simd
GO ?= go
ARGS ?=
ARTIFACT_DIR := packages/ops/artifacts
BENCHTIME ?= 100ms

.PHONY: help run report clean

help:
	@printf '%s\n' \
		'Usage:' \
		'  make run                         Run all CLI benchmarks' \
		'  make run ARGS="-function add"   Run one CLI benchmark group' \
		'  make report                     Save detailed multi-size tables' \
		'  make clean                      Remove generated benchmark artifacts'

run:
	@GOEXPERIMENT=$(GOEXPERIMENT) $(GO) run ./cmd/simdbench $(ARGS)

report:
	@printf '%s\n' 'Quick comparison (1,048,576 elements):'
	@GOEXPERIMENT=$(GOEXPERIMENT) $(GO) run ./cmd/simdbench
	@mkdir -p $(ARTIFACT_DIR)
	@report="$(ARTIFACT_DIR)/benchmark-$$(date +%Y%m%d-%H%M%S).txt"; \
	raw="$$report.raw"; \
	trap 'rm -f "$$raw"' EXIT; \
	if GOEXPERIMENT=$(GOEXPERIMENT) $(GO) test ./packages/ops \
		-run='^$$' -bench='^Benchmark(Add|Compute)$$' \
		-benchmem -benchtime=$(BENCHTIME) -count=1 > "$$raw" 2>&1; then \
		awk '\
			/^(Addition|Compute-heavy expression) benchmark comparison/ { \
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
