GOEXPERIMENT ?= simd
GO ?= go
ARGS ?=
BENCHTIME ?= 100ms
ARTIFACT_ROOT := packages/benches/artifacts

.PHONY: help run run-addition run-computation report report-addition report-computation clean

help:
	@printf '%s\n' \
		'Usage:' \
		'  make run                           Run all CLI benchmarks' \
		'  make run ARGS="-group addition"   Run the addition benchmark' \
		'  make run ARGS="-group computation" Run the computation benchmark' \
		'  make run-addition                 Run only addition benchmarks' \
		'  make run-computation              Run only computation benchmarks' \
		'  make report                       Save detailed tables for all operations' \
		'  make report-addition              Save the addition table' \
		'  make report-computation           Save the computation table' \
		'  make clean                        Remove generated benchmark artifacts'

run:
	@GOEXPERIMENT=$(GOEXPERIMENT) $(GO) run ./cmd/simdbench $(ARGS)

run-addition:
	@GOEXPERIMENT=$(GOEXPERIMENT) $(GO) run ./cmd/simdbench -addition

run-computation:
	@GOEXPERIMENT=$(GOEXPERIMENT) $(GO) run ./cmd/simdbench -computation

report:
	@printf '%s\n' 'Quick comparison (1,048,576 elements):'
	@GOEXPERIMENT=$(GOEXPERIMENT) $(GO) run ./cmd/simdbench
	@$(MAKE) --no-print-directory report-addition report-computation

report-addition:
	@set -eu; \
	directory="$(ARTIFACT_ROOT)/addition"; \
	mkdir -p "$$directory"; \
	timestamp="$$(date +%Y%m%d-%H%M%S)"; \
	report="$$directory/benchmark-$$timestamp.txt"; \
	suffix=1; \
	while [ -e "$$report" ]; do \
		report="$$directory/benchmark-$$timestamp-$$suffix.txt"; \
		suffix=$$((suffix + 1)); \
	done; \
	raw="$$report.raw"; \
	trap 'rm -f "$$raw"' EXIT; \
	if bash scripts/with-elapsed-time.sh "$$raw" 'Working on addition comparison table' \
		env GOEXPERIMENT=$(GOEXPERIMENT) $(GO) test ./packages/benches/addition \
		-run='^$$' -bench='^BenchmarkAddition$$' \
		-benchmem -benchtime=$(BENCHTIME) -count=1; then \
		awk '/^Addition benchmark comparison/ { printing=1; print; next } \
			printing && /^ELEMENTS[[:space:]]/ { print; next } \
			printing && /^[[:space:]]*[0-9]+[[:space:]]/ { print; next } \
			printing { printing=0 }' "$$raw" > "$$report"; \
		if [ ! -s "$$report" ]; then \
			cat "$$raw"; \
			printf 'No addition comparison table found\n' >&2; \
			rm -f "$$report"; \
			exit 1; \
		fi; \
		printf 'Addition comparison saved to %s\n' "$$report"; \
	else \
		cat "$$raw"; \
		exit 1; \
	fi

report-computation:
	@set -eu; \
	directory="$(ARTIFACT_ROOT)/computation"; \
	mkdir -p "$$directory"; \
	timestamp="$$(date +%Y%m%d-%H%M%S)"; \
	report="$$directory/benchmark-$$timestamp.txt"; \
	suffix=1; \
	while [ -e "$$report" ]; do \
		report="$$directory/benchmark-$$timestamp-$$suffix.txt"; \
		suffix=$$((suffix + 1)); \
	done; \
	raw="$$report.raw"; \
	trap 'rm -f "$$raw"' EXIT; \
	if bash scripts/with-elapsed-time.sh "$$raw" 'Working on computation comparison table' \
		env GOEXPERIMENT=$(GOEXPERIMENT) $(GO) test ./packages/benches/computation \
		-run='^$$' -bench='^BenchmarkComputation$$' \
		-benchmem -benchtime=$(BENCHTIME) -count=1; then \
		awk '/^Computation benchmark comparison/ { printing=1; print; next } \
			printing && /^ELEMENTS[[:space:]]/ { print; next } \
			printing && /^[[:space:]]*[0-9]+[[:space:]]/ { print; next } \
			printing { printing=0 }' "$$raw" > "$$report"; \
		if [ ! -s "$$report" ]; then \
			cat "$$raw"; \
			printf 'No computation comparison table found\n' >&2; \
			rm -f "$$report"; \
			exit 1; \
		fi; \
		printf 'Computation comparison saved to %s\n' "$$report"; \
	else \
		cat "$$raw"; \
		exit 1; \
	fi

clean:
	rm -rf $(ARTIFACT_ROOT)
