#!/usr/bin/env bash
# crud_reprofile_compare.sh — Re-run internal/database CRUD benchmarks, capture pprof, compare to baseline.
#
# Usage (from exarp-go repo root):
#   bash scripts/crud_reprofile_compare.sh bench           # run benches → logs/crud_bench_latest.txt; benchstat if baseline exists
#   bash scripts/crud_reprofile_compare.sh save-baseline   # copy latest → logs/crud_bench_baseline.txt
#   bash scripts/crud_reprofile_compare.sh compare         # benchstat only (no go test)
#   bash scripts/crud_reprofile_compare.sh pprof           # CPU + heap profiles with UTC timestamp
#   bash scripts/crud_reprofile_compare.sh pprof-diff OLD.pprof NEW.pprof [binary]
#
# Env: CRUD_BENCH_COUNT (default 5), CRUD_BENCH_TIME (optional, e.g. 2s), GO (go binary).
# benchstat: go install golang.org/x/perf/cmd/benchstat@latest
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

GO="${GO:-go}"
CRUD_BENCH='Benchmark(CreateTask|GetTask|UpdateTask|DeleteTask|BatchUpdateTaskStatus_64)'
PKG="./internal/database/"
LATEST="${CRUD_BENCH_LATEST:-logs/crud_bench_latest.txt}"
BASELINE="${CRUD_BENCH_BASELINE:-logs/crud_bench_baseline.txt}"
COUNT="${CRUD_BENCH_COUNT:-5}"
BENCHTIME_ARG=()
if [ -n "${CRUD_BENCH_TIME:-}" ]; then
	BENCHTIME_ARG=(-benchtime="${CRUD_BENCH_TIME}")
fi

usage() {
	sed -n '1,20p' "$0" | sed 's/^# \{0,1\}//'
	exit "${1:-0}"
}

run_bench() {
	mkdir -p logs
	echo "Running CRUD benchmarks (CGO_ENABLED=0, -count=${COUNT}) → ${LATEST}"
	# shellcheck disable=SC2086
	CGO_ENABLED=0 "$GO" test -run='^$' -bench="${CRUD_BENCH}" -benchmem -count="${COUNT}" \
		"${BENCHTIME_ARG[@]}" "$PKG" 2>&1 | tee "${LATEST}"
}

maybe_benchstat() {
	if [ ! -f "${BASELINE}" ]; then
		echo "No baseline at ${BASELINE}. After a change, save one: bash scripts/crud_reprofile_compare.sh save-baseline"
		return 0
	fi
	if ! command -v benchstat >/dev/null 2>&1; then
		echo "benchstat not in PATH; install: go install golang.org/x/perf/cmd/benchstat@latest"
		return 0
	fi
	echo ""
	echo "=== benchstat: ${BASELINE} vs ${LATEST} ==="
	benchstat "${BASELINE}" "${LATEST}"
}

cmd_bench() {
	run_bench
	maybe_benchstat
}

cmd_save_baseline() {
	if [ ! -f "${LATEST}" ]; then
		echo "Missing ${LATEST}; run: bash scripts/crud_reprofile_compare.sh bench" >&2
		exit 1
	fi
	cp -f "${LATEST}" "${BASELINE}"
	echo "Saved baseline → ${BASELINE}"
}

cmd_compare() {
	if [ ! -f "${BASELINE}" ] || [ ! -f "${LATEST}" ]; then
		echo "Need both ${BASELINE} and ${LATEST}" >&2
		exit 1
	fi
	if ! command -v benchstat >/dev/null 2>&1; then
		echo "benchstat not in PATH; install: go install golang.org/x/perf/cmd/benchstat@latest" >&2
		exit 1
	fi
	benchstat "${BASELINE}" "${LATEST}"
}

cmd_pprof() {
	mkdir -p logs
	echo "Building test binary → logs/database.test"
	CGO_ENABLED=0 "$GO" test -c -o logs/database.test "${PKG}"
	ts="$(date -u +%Y%m%dT%H%M%SZ)"
	cpu="logs/crud_cpu_${ts}.pprof"
	mem="logs/crud_mem_${ts}.pprof"
	echo "Capturing CPU → ${cpu} and heap → ${mem} via logs/database.test (-count=1)"
	./logs/database.test -test.v=false -test.run='^$' -test.bench="${CRUD_BENCH}" -test.count=1 \
		-test.cpuprofile="${cpu}" -test.memprofile="${mem}"
	echo ""
	echo "Top CPU (text):"
	"$GO" tool pprof -text -nodecount=25 logs/database.test "${cpu}" || true
	echo ""
	echo "Top alloc_space:"
	"$GO" tool pprof -text -sample_index=alloc_space -nodecount=25 logs/database.test "${mem}" || true
	echo ""
	echo "Compare CPU after a change (same binary build recommended):"
	echo "  bash scripts/crud_reprofile_compare.sh pprof-diff ${cpu} <new_cpu.pprof>"
}

cmd_pprof_diff() {
	if [ "${#}" -lt 2 ]; then
		echo "usage: pprof-diff <old_cpu.pprof> <new_cpu.pprof> [path_to_test_binary]" >&2
		exit 1
	fi
	old_pf="$1"
	new_pf="$2"
	bin="${3:-logs/database.test}"
	if [ ! -f "${bin}" ]; then
		echo "Binary not found: ${bin} (run pprof command first or pass path)" >&2
		exit 1
	fi
	if [ ! -f "${old_pf}" ] || [ ! -f "${new_pf}" ]; then
		echo "Profile file missing." >&2
		exit 1
	fi
	echo "=== pprof -base (delta vs baseline) ==="
	"$GO" tool pprof -text -nodecount=40 -base="${old_pf}" "${bin}" "${new_pf}"
}

case "${1:-bench}" in
-h | --help | help)
	usage 0
	;;
bench)
	cmd_bench
	;;
save-baseline)
	cmd_save_baseline
	;;
compare)
	cmd_compare
	;;
pprof)
	cmd_pprof
	;;
pprof-diff)
	shift
	cmd_pprof_diff "$@"
	;;
*)
	echo "Unknown command: $1" >&2
	usage 1
	;;
esac
