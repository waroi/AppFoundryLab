#!/usr/bin/env python3
import argparse
import json
from pathlib import Path


def build_summary(mode: str, status: str, automated_count: int, manual_count: int) -> str:
    return "\n".join(
        [
            "## Release Gate Fast Summary",
            f"- Status: `{status}`",
            f"- Mode: `{mode}`",
            f"- Automated checks: `{automated_count}`",
            f"- Manual checks to review: `{manual_count}`",
        ]
    ) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(description="Parse release-gate JSON report.")
    parser.add_argument("--input", required=True, help="Input JSON report path.")
    parser.add_argument("--summary-md", help="Markdown summary output path.")
    parser.add_argument("--github-output", help="Path to $GITHUB_OUTPUT file.")
    args = parser.parse_args()

    report_path = Path(args.input)
    data = json.loads(report_path.read_text(encoding="utf-8"))

    status = data.get("status", "")
    mode = data.get("mode", "")
    automated_checks = data.get("automatedChecks", [])
    manual_checks = data.get("manualChecks", [])

    if status != "passed":
        raise SystemExit(f"release gate status is not passed: {status}")
    if mode not in {"fast", "full"}:
        raise SystemExit(f"release gate mode is invalid: {mode}")
    if not isinstance(automated_checks, list) or not isinstance(manual_checks, list):
        raise SystemExit("release gate report has invalid check arrays")

    automated_count = len(automated_checks)
    manual_count = len(manual_checks)
    summary = build_summary(mode, status, automated_count, manual_count)

    if args.summary_md:
        Path(args.summary_md).write_text(summary, encoding="utf-8")

    if args.github_output:
        with Path(args.github_output).open("a", encoding="utf-8") as output:
            output.write(f"release_gate_status={status}\n")
            output.write(f"release_gate_mode={mode}\n")
            output.write(f"release_gate_automated_count={automated_count}\n")
            output.write(f"release_gate_manual_count={manual_count}\n")

    print(
        f"release gate report parsed: status={status} mode={mode} "
        f"automated={automated_count} manual={manual_count}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
