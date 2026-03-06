#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
POLICY_FILE="$ROOT_DIR/docs/release-policy.md"
CHECKLIST_JSON="$ROOT_DIR/docs/release-checklist.json"
WORKFLOW_FILE="$ROOT_DIR/.github/workflows/appfoundrylab-ci.yml"
NIGHTLY_WORKFLOW_FILE="$ROOT_DIR/.github/workflows/release-gate-full-nightly.yml"
DELIVERY_WORKFLOW_MATRIX="$ROOT_DIR/docs/delivery-workflow-governance-matrix.md"

require_contains() {
  local file="$1"
  local pattern="$2"
  local message="$3"
  if ! grep -Fq "$pattern" "$file"; then
    echo "release policy drift check failed: $message" >&2
    exit 1
  fi
}

require_contains "$POLICY_FILE" "Release Gate (fast)" "release policy missing 'Release Gate (fast)' reference"
require_contains "$POLICY_FILE" "docs/release-checklist.json" "release policy missing canonical checklist json reference"
require_contains "$WORKFLOW_FILE" "release-gate-fast:" "workflow missing release-gate-fast job"

require_contains "$POLICY_FILE" "Perf benchmark smoke + trend diff" "release policy missing perf trend review requirement"
require_contains "$WORKFLOW_FILE" "Compare k6 summary vs base (PR)" "workflow missing k6 trend diff step"

require_contains "$POLICY_FILE" "Release Gate (full nightly)" "release policy missing full-nightly reference"
require_contains "$NIGHTLY_WORKFLOW_FILE" "release-gate-full-nightly:" "nightly workflow missing release-gate-full-nightly job"
require_contains "$POLICY_FILE" "perf-extended-nightly" "release policy missing perf-extended-nightly reference"
require_contains "$ROOT_DIR/.github/workflows/perf-extended-nightly.yml" "perf-extended-nightly:" "workflow missing perf-extended-nightly job"

require_contains "$POLICY_FILE" "frontend-api-contract-matrix" "release policy missing frontend api contract matrix reference"
require_contains "$WORKFLOW_FILE" "frontend-api-contract-matrix:" "workflow missing frontend-api-contract-matrix job"
require_contains "$POLICY_FILE" "profile-capacity-matrix" "release policy missing profile capacity matrix reference"
require_contains "$WORKFLOW_FILE" "profile-capacity-matrix:" "workflow missing profile-capacity-matrix job"
require_contains "$WORKFLOW_FILE" "check-doc-drift.sh --mode strict" "workflow missing strict doc drift mode"
require_contains "$POLICY_FILE" "boilerplate-quality-report" "release policy missing boilerplate quality report reference"
require_contains "$WORKFLOW_FILE" "boilerplate-quality-report:" "workflow missing boilerplate-quality-report job"
require_contains "$POLICY_FILE" "docs/load-shedding-policy.json" "release policy missing load-shedding policy reference"
require_contains "$POLICY_FILE" "docs/load-shedding-runbook.md" "release policy missing load-shedding runbook reference"
require_contains "$POLICY_FILE" "docs/nightly-workflow-governance-matrix.md" "release policy missing nightly workflow governance matrix reference"
require_contains "$POLICY_FILE" "docs/delivery-workflow-governance-matrix.md" "release policy missing delivery workflow governance matrix reference"
require_contains "$POLICY_FILE" "docs/branch-protection-required-checks.md" "release policy missing branch protection mapping reference"

if [[ ! -f "$CHECKLIST_JSON" ]]; then
  echo "release policy drift check failed: missing canonical checklist json ($CHECKLIST_JSON)" >&2
  exit 1
fi

if [[ ! -f "$DELIVERY_WORKFLOW_MATRIX" ]]; then
  echo "release policy drift check failed: missing delivery workflow governance matrix ($DELIVERY_WORKFLOW_MATRIX)" >&2
  exit 1
fi

python3 - "$CHECKLIST_JSON" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
data = json.loads(path.read_text(encoding="utf-8"))
root = path.parent
delivery_doc = root / "delivery-workflow-governance-matrix.md"
nightly_doc = root / "nightly-workflow-governance-matrix.md"
branch_doc = root / "branch-protection-required-checks.md"
release_policy_doc = root / "release-policy.md"
rg = data.get("releaseGate", {})
auto = rg.get("automatedChecks")
manual = rg.get("manualChecks")
ci_refs = data.get("ciReferences")
if not isinstance(auto, list) or not auto:
    raise SystemExit("release policy drift check failed: releaseGate.automatedChecks invalid")
if not isinstance(manual, list) or not manual:
    raise SystemExit("release policy drift check failed: releaseGate.manualChecks invalid")
if not isinstance(ci_refs, list):
    raise SystemExit("release policy drift check failed: ciReferences invalid")
required = {"boilerplate-quality-report", "profile-capacity-matrix", "perf-extended-nightly"}
missing = sorted(required - set(ci_refs))
if missing:
    raise SystemExit(f"release policy drift check failed: ciReferences missing {', '.join(missing)}")
manual_ids = {item.get("id") for item in data.get("releaseChecklist", []) if item.get("kind") == "manual"}
if "load-shed-policy-review" not in manual_ids:
    raise SystemExit("release policy drift check failed: releaseChecklist missing load-shed-policy-review")
if "load-shed-policy-review" not in set(manual):
    raise SystemExit("release policy drift check failed: releaseGate.manualChecks missing load-shed-policy-review")

docs = {
    "delivery": delivery_doc.read_text(encoding="utf-8"),
    "nightly": nightly_doc.read_text(encoding="utf-8"),
    "branch": branch_doc.read_text(encoding="utf-8"),
    "policy": release_policy_doc.read_text(encoding="utf-8"),
}

coverage_expectations = {
    "release-gate-fast": {"delivery"},
    "release-gate-full-optional": {"nightly"},
    "ci-required-jobs-green": {"delivery", "branch"},
    "trivy-gitleaks-review": {"delivery"},
    "dependabot-queue-review": {"delivery", "policy"},
    "perf-trend-diff-review": {"delivery"},
    "load-shed-policy-review": {"delivery", "policy"},
}

for checklist_id, locations in coverage_expectations.items():
    if checklist_id not in {item.get("id") for item in data.get("releaseChecklist", [])}:
        raise SystemExit(f"release policy drift check failed: releaseChecklist missing expected id {checklist_id}")
    for location in locations:
        if checklist_id not in docs[location]:
            raise SystemExit(
                f"release policy drift check failed: checklist id {checklist_id} missing in {location} governance docs"
            )

ci_ref_expectations = {
    "release-gate-fast": {"delivery", "policy"},
    "release-gate-full-nightly": {"nightly", "policy"},
    "perf-extended-nightly": {"nightly", "policy"},
    "frontend-api-contract-matrix": {"policy"},
    "profile-capacity-matrix": {"delivery", "policy"},
    "boilerplate-quality-report": {"delivery", "policy", "branch"},
    "compare-k6-summary-vs-base-pr": {"delivery", "policy"},
}

for ci_ref, locations in ci_ref_expectations.items():
    if ci_ref not in set(ci_refs):
        raise SystemExit(f"release policy drift check failed: ciReferences missing expected id {ci_ref}")
    for location in locations:
        if ci_ref not in docs[location]:
            raise SystemExit(
                f"release policy drift check failed: ci reference {ci_ref} missing in {location} governance docs"
            )
PY

echo "release policy drift check passed"
