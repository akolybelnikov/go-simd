#!/usr/bin/env bash

set -u

if [ "$#" -lt 3 ]; then
	printf 'Usage: %s OUTPUT_FILE LABEL COMMAND [ARG...]\n' "$0" >&2
	exit 2
fi

output_file=$1
label=$2
shift 2
start_seconds=$SECONDS

(
	sleep_pid=
	trap 'if [ -n "$sleep_pid" ]; then kill "$sleep_pid" 2>/dev/null || :; fi; exit 0' TERM INT
	elapsed_seconds=0
	while :; do
		printf '%s: %d seconds elapsed\n' "$label" "$elapsed_seconds" >&2
		sleep 1 &
		sleep_pid=$!
		wait "$sleep_pid" || :
		sleep_pid=
		elapsed_seconds=$((elapsed_seconds + 1))
	done
) &
timer_pid=$!

set +e
"$@" >"$output_file" 2>&1
command_status=$?
set -e

kill "$timer_pid" 2>/dev/null || :
wait "$timer_pid" 2>/dev/null || :

elapsed_seconds=$((SECONDS - start_seconds))
printf '%s: finished in %d seconds\n' "$label" "$elapsed_seconds" >&2
exit "$command_status"
