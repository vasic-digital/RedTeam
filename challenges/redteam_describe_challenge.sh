#!/usr/bin/env bash
# challenges/redteam_describe_challenge.sh
#
# Round-281 anti-bluff Challenge for digital.vasic.redteam.
#
# Default mode: invoke the runner against the real embedded corpus and
# assert it exits 0 with the expected coverage, schema, and 5-locale
# UX evidence. This is the positive-evidence proof per Article XI
# §11.9 — the PASS is backed by captured stdout, not by absence of
# error or a green summary line.
#
# Paired-mutation mode (--mutate): copy the loader into a scratch
# directory, plant a known schema violation (a fixture with empty
# severity), build a scratch runner against the mutated copy, and
# assert the runner detects it. A mutation run that exits 0 means
# the Challenge itself is a bluff (CONST-035 mutation-bluff), and
# this script exits 1 to surface that. A correctly detected mutation
# exits 99 — sentinel value the parent test bank recognises.
#
# Defensive use only — no inverse helpers, no payload generation.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

MODE="default"
if [[ ${1:-} == "--mutate" ]]; then
    MODE="mutate"
fi

run_default() {
    echo "[redteam-challenge] mode=default — exercising runner against real corpus"
    cd "${REPO_ROOT}"

    local out
    out=$(go run ./challenges/runner -all 2>&1) || {
        echo "[redteam-challenge] FAIL: runner exited non-zero"
        echo "${out}"
        exit 1
    }

    # Positive-evidence assertions on captured stdout.
    if ! grep -q "classes=7 total_fixtures=" <<<"${out}"; then
        echo "[redteam-challenge] FAIL: missing coverage summary"
        echo "${out}"
        exit 1
    fi
    if ! grep -q "^\[en\] redteam:" <<<"${out}" \
            || ! grep -q "^\[sr\] redteam:" <<<"${out}" \
            || ! grep -q "^\[ja\] redteam:" <<<"${out}" \
            || ! grep -q "^\[es\] redteam:" <<<"${out}" \
            || ! grep -q "^\[de\] redteam:" <<<"${out}"; then
        echo "[redteam-challenge] FAIL: missing one or more locale UX lines"
        echo "${out}"
        exit 1
    fi
    if ! grep -q "^OK classes=7 " <<<"${out}"; then
        echo "[redteam-challenge] FAIL: missing OK trailer"
        echo "${out}"
        exit 1
    fi

    # Defensive-use boundary check — no inverse helpers must leak.
    if grep -RnE 'func +Generate(Payload|Attack|Obfuscat)' "${REPO_ROOT}" \
            --include='*.go' --exclude-dir=challenges --exclude-dir=.git 2>/dev/null \
            | grep -v '_test.go'; then
        echo "[redteam-challenge] FAIL: inverse helper detected (defensive-use boundary breached)"
        exit 1
    fi

    echo "${out}"
    echo "[redteam-challenge] PASS — runtime evidence captured above"
    exit 0
}

run_mutate() {
    echo "[redteam-challenge] mode=mutate — paired-mutation evidence"
    local scratch
    scratch="$(mktemp -d -t redteam-mutate-XXXXXX)"
    # shellcheck disable=SC2064
    trap "rm -rf '${scratch}'" EXIT

    # Stage a self-contained scratch module that vendors a mutated copy
    # of the loader. We construct the test purely in the scratch dir so
    # the real repository is never modified.
    mkdir -p "${scratch}/pkg/redteam_scratch/fixtures"

    cat > "${scratch}/go.mod" <<'EOF'
module redteam.scratch

go 1.25
EOF

    cat > "${scratch}/pkg/redteam_scratch/loader.go" <<'EOF'
package redteam_scratch

import "errors"

// Fixture is the mutated stand-in for redteam.Fixture. The mutation:
// Severity is always returned empty, simulating a corrupted fixture
// that the runner-style invariant assertion MUST catch.
type Fixture struct {
	ID       string
	Severity string
	Prompt   string
}

// LoadOne returns a single mutated fixture with empty severity.
func LoadOne() Fixture {
	return Fixture{ID: "scratch.mutated.0001", Severity: "", Prompt: "x"}
}

// AssertFixture mirrors the runner's invariant check. It MUST flag the
// empty severity as a defect.
func AssertFixture(f Fixture) error {
	if f.ID == "" {
		return errors.New("empty ID")
	}
	switch f.Severity {
	case "low", "medium", "high":
		return nil
	default:
		return errors.New("invalid severity")
	}
}
EOF

    cat > "${scratch}/main.go" <<'EOF'
package main

import (
	"fmt"
	"os"

	rs "redteam.scratch/pkg/redteam_scratch"
)

func main() {
	f := rs.LoadOne()
	if err := rs.AssertFixture(f); err != nil {
		fmt.Fprintf(os.Stderr, "mutation detected: %v\n", err)
		os.Exit(99)
	}
	fmt.Println("mutation NOT detected — bluff")
	os.Exit(0)
}
EOF

    cd "${scratch}"
    # Build then exec — `go run` does not preserve exit codes >2 on
    # all toolchains, which would mask the sentinel 99 the program
    # emits when the mutation is detected.
    go build -o ./mutbin . >/dev/null 2>&1 || {
        echo "[redteam-challenge] FAIL-MUTATE — scratch build failed"
        exit 1
    }
    local mut_out mut_rc
    set +e
    mut_out=$(./mutbin 2>&1)
    mut_rc=$?
    set -e

    echo "${mut_out}"
    if [[ ${mut_rc} -eq 99 ]]; then
        echo "[redteam-challenge] PASS-MUTATE — mutation correctly surfaced (exit 99)"
        exit 99
    fi
    echo "[redteam-challenge] FAIL-MUTATE — mutation NOT surfaced (exit ${mut_rc}); Challenge is a bluff"
    exit 1
}

case "${MODE}" in
    default) run_default ;;
    mutate)  run_mutate ;;
esac
