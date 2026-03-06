#!/usr/bin/env bash
set -euo pipefail

MODE="advisory"
BASE_REF=""
HEAD_REF="HEAD"
GIT_BIN="${GIT_BIN:-git}"

usage() {
  cat <<'EOF'
usage: ./scripts/check-doc-drift.sh [--mode advisory|strict] [--base-ref <ref>] [--head-ref <ref>]

examples:
  ./scripts/check-doc-drift.sh --mode advisory
  ./scripts/check-doc-drift.sh --mode strict --base-ref origin/main --head-ref HEAD
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode)
      MODE="${2:-}"
      shift 2
      ;;
    --base-ref)
      BASE_REF="${2:-}"
      shift 2
      ;;
    --head-ref)
      HEAD_REF="${2:-}"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    --)
      shift
      break
      ;;
    -*)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
    *)
      if [[ -z "$BASE_REF" ]]; then
        BASE_REF="$1"
      elif [[ "$HEAD_REF" == "HEAD" ]]; then
        HEAD_REF="$1"
      else
        echo "too many positional arguments" >&2
        usage >&2
        exit 1
      fi
      shift
      ;;
  esac
done

resolve_git_bin() {
  if command -v "$GIT_BIN" >/dev/null 2>&1; then
    return 0
  fi

  if [[ "$GIT_BIN" != "git" ]]; then
    return 1
  fi

  local candidate
  for candidate in \
    "/mnt/c/Program Files/Git/cmd/git.exe" \
    "/mnt/c/Program Files/Git/bin/git.exe"; do
    if [[ -x "$candidate" ]]; then
      GIT_BIN="$candidate"
      return 0
    fi
  done

  return 1
}

normalize_changed_files_override() {
  local raw="${DOC_DRIFT_CHANGED_FILES:-}"
  if [[ -z "$raw" ]]; then
    return 0
  fi
  printf '%s\n' "$raw" \
    | tr ',' '\n' \
    | sed -E 's/^[[:space:]]+//; s/[[:space:]]+$//' \
    | sed '/^$/d'
}

semantic_failures=()

note_semantic_failure() {
  semantic_failures+=("$1")
}

require_file_contains() {
  local path="$1"
  local pattern="$2"
  local description="$3"

  if [[ ! -f "$path" ]]; then
    note_semantic_failure "$path: expected file is missing ($description)"
    return
  fi

  if ! grep -Fq -- "$pattern" "$path"; then
    note_semantic_failure "$path: missing '$pattern' ($description)"
  fi
}

require_file_contains_any() {
  local path="$1"
  local description="$2"
  shift 2

  if [[ ! -f "$path" ]]; then
    note_semantic_failure "$path: expected file is missing ($description)"
    return
  fi

  local pattern
  for pattern in "$@"; do
    if grep -Fq -- "$pattern" "$path"; then
      return
    fi
  done

  note_semantic_failure "$path: missing one of [$*] ($description)"
}

require_file_not_contains() {
  local path="$1"
  local pattern="$2"
  local description="$3"

  if [[ ! -f "$path" ]]; then
    note_semantic_failure "$path: expected file is missing ($description)"
    return
  fi

  if grep -Fq -- "$pattern" "$path"; then
    note_semantic_failure "$path: contains forbidden text '$pattern' ($description)"
  fi
}

run_semantic_doc_truth_checks() {
  require_file_not_contains \
    "docs/deployment-strategy.md" \
    "./scripts/archive-runtime-report.sh https://api.example.com admin strong_password" \
    "deployment docs must not demonstrate positional admin passwords"

  local archive_docs=(
    "docs/deployment-strategy.md"
    "docs/en/deployment.md"
    "docs/tr/deployment.md"
    "scripts/README.md"
  )
  local doc
  for doc in "${archive_docs[@]}"; do
    if [[ -f "$doc" ]] && grep -Fq "archive-runtime-report.sh" "$doc"; then
      require_file_contains_any \
        "$doc" \
        "archive-runtime-report usage should point operators to env/stdin credentials" \
        "--password-stdin" \
        "DEPLOY_ADMIN_PASSWORD"
    fi
  done

  if grep -R -Fq "LEDGER_ATTESTATION_REQUIRE_SIGNED: 'true'" .github/workflows; then
    local evidence_docs=(
      "docs/deployment-strategy.md"
      "docs/en/deployment.md"
      "docs/tr/deployment.md"
      "docs/en/operations.md"
      "docs/tr/operasyonlar.md"
    )
    for doc in "${evidence_docs[@]}"; do
      require_file_contains "$doc" "RELEASE_EVIDENCE_AUDIT_TARGET" "signed evidence workflows require an audit export target"
      require_file_contains_any \
        "$doc" \
        "signed evidence workflows require the attestation signing secret" \
        "RELEASE_LEDGER_ATTESTATION_KEY" \
        "LEDGER_ATTESTATION_SIGNING_KEY"
    done
  fi

  if [[ -f "frontend/e2e/live-stack.spec.ts" ]]; then
    local smoke_docs=(
      "README.md"
      "docs/en/quick-start.md"
      "docs/tr/hizli-baslangic.md"
      "docs/en/testing-and-quality.md"
      "docs/tr/test-ve-kalite.md"
    )
    for doc in "${smoke_docs[@]}"; do
      require_file_contains "$doc" "e2e:live" "docs should mention the real-stack browser smoke"
      require_file_contains "$doc" "e2e" "docs should mention the mock-backed browser regression lane"
    done
  fi

  require_file_not_contains \
    "docs/en/testing-and-quality.md" \
    "until Phase 1 toolchain alignment is complete" \
    "testing docs must not describe already-closed toolchain work as open"
  require_file_not_contains \
    "docs/en/testing-and-quality.md" \
    "SystemStatus.svelte is still too large and needs decomposition" \
    "testing docs must not describe the completed SystemStatus split as open"
  require_file_not_contains \
    "docs/en/testing-and-quality.md" \
    'repo-local Go toolchain alignment remains open in `PROGRESS.md`' \
    "testing docs must not point at stale backlog items"

  require_file_not_contains \
    "docs/tr/test-ve-kalite.md" \
    "Faz 1 toolchain hizalamasina kadar" \
    "TR testing docs must not describe already-closed toolchain work as open"
  require_file_not_contains \
    "docs/tr/test-ve-kalite.md" \
    "SystemStatus.svelte halen fazla buyuk ve parcali bakima ihtiyac duyuyor" \
    "TR testing docs must not describe the completed SystemStatus split as open"
  require_file_not_contains \
    "docs/tr/test-ve-kalite.md" \
    'repo-ici Go toolchain hizalamasi `PROGRESS.md` icinde halen acik' \
    "TR testing docs must not point at stale backlog items"
}

if [[ "$MODE" != "advisory" && "$MODE" != "strict" ]]; then
  echo "invalid mode: $MODE" >&2
  usage >&2
  exit 1
fi

CHANGED_FILES=""
if resolve_git_bin; then
  if ! "$GIT_BIN" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    echo "doc drift check skipped: not a git work tree"
    exit 0
  fi

  if [[ -z "$BASE_REF" ]]; then
    if [[ -n "${GITHUB_BASE_REF:-}" ]]; then
      BASE_REF="origin/${GITHUB_BASE_REF}"
    else
      CHANGED_FILES="$("$GIT_BIN" -c core.quotePath=false diff --name-only HEAD || true)"
    fi
  fi

  if [[ -n "$BASE_REF" ]]; then
    RANGE="${BASE_REF}...${HEAD_REF}"
    CHANGED_FILES="$("$GIT_BIN" -c core.quotePath=false diff --name-only "$RANGE" || true)"
  fi
else
  CHANGED_FILES="$(normalize_changed_files_override)"
  if [[ -z "$CHANGED_FILES" ]]; then
    if [[ "$MODE" == "strict" ]]; then
      echo "doc drift check failed (strict): git command not found and DOC_DRIFT_CHANGED_FILES is empty" >&2
      exit 1
    fi
    echo "doc drift check skipped: git command not found and DOC_DRIFT_CHANGED_FILES is empty"
    exit 0
  fi
  echo "doc drift check: using DOC_DRIFT_CHANGED_FILES fallback"
fi

if [[ -z "$CHANGED_FILES" ]]; then
  echo "doc drift check skipped: no changed files"
  exit 0
fi

if ! echo "$CHANGED_FILES" | grep -Eq '^(backend/|frontend/|docker-compose\.yml|docker-compose\.security\.yml|\.github/workflows/|scripts/|toolchain\.versions\.json)'; then
  echo "doc drift check passed: no code/infrastructure changes requiring docs"
  exit 0
fi

required_docs=(
  "README.md"
  "docs/appfoundrylab-teknik-analiz.md"
  "docs/gelistirmePlanı.md"
)

missing_docs=()
for doc in "${required_docs[@]}"; do
  if ! echo "$CHANGED_FILES" | grep -Fxq "$doc"; then
    missing_docs+=("$doc")
  fi
done

if [[ "${#missing_docs[@]}" -gt 0 ]]; then
  if [[ "$MODE" == "strict" ]]; then
    echo "doc drift check failed (strict): missing updated docs"
    for doc in "${missing_docs[@]}"; do
      echo "  - $doc"
    done
    echo "changed files:"
    echo "$CHANGED_FILES"
    exit 1
  fi

  echo "doc drift check advisory: missing updated docs"
  for doc in "${missing_docs[@]}"; do
    echo "  - $doc"
  done
  echo "changed files:"
  echo "$CHANGED_FILES"
  echo "advisory mode does not fail the command"
  exit 0
fi

run_semantic_doc_truth_checks

if [[ "${#semantic_failures[@]}" -gt 0 ]]; then
  if [[ "$MODE" == "strict" ]]; then
    echo "doc drift check failed (strict): semantic doc truth checks failed"
    for failure in "${semantic_failures[@]}"; do
      echo "  - $failure"
    done
    exit 1
  fi

  echo "doc drift check advisory: semantic doc truth checks failed"
  for failure in "${semantic_failures[@]}"; do
    echo "  - $failure"
  done
  echo "advisory mode does not fail the command"
  exit 0
fi

echo "doc drift check passed"
