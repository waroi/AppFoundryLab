#!/usr/bin/env bash
set -euo pipefail

MODE="advisory"
BASE_REF=""
HEAD_REF="HEAD"

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

if [[ "$MODE" != "advisory" && "$MODE" != "strict" ]]; then
  echo "invalid mode: $MODE" >&2
  usage >&2
  exit 1
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "doc drift check skipped: not a git work tree"
  exit 0
fi

if [[ -z "$BASE_REF" ]]; then
  if [[ -n "${GITHUB_BASE_REF:-}" ]]; then
    BASE_REF="origin/${GITHUB_BASE_REF}"
  else
    BASE_REF="HEAD~1"
  fi
fi

RANGE="${BASE_REF}...${HEAD_REF}"

CHANGED_FILES="$(git diff --name-only "$RANGE" || true)"
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

echo "doc drift check passed"
