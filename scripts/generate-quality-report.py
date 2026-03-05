#!/usr/bin/env python3
import argparse
import json
import re
from pathlib import Path


def parse_readme_scores(readme_text: str):
    rows = []
    for line in readme_text.splitlines():
        if not line.startswith("|"):
            continue
        if "Skor (10)" in line or line.startswith("|---"):
            continue
        parts = [p.strip() for p in line.strip().strip("|").split("|")]
        if len(parts) < 4:
            continue
        try:
            score = float(parts[1].replace(",", "."))
        except ValueError:
            continue
        rows.append(
            {
                "criterion": parts[0],
                "score": score,
                "level": parts[2],
                "summary": parts[3],
            }
        )
    return rows


def parse_plan_tasks(plan_text: str):
    completed = len(re.findall(r"^- \[x\]", plan_text, flags=re.MULTILINE))
    pending = len(re.findall(r"^- \[ \]", plan_text, flags=re.MULTILINE))
    return completed, pending


def parse_overall_score(readme_text: str):
    m = re.search(r"Genel skor:\s+\*\*([0-9]+(?:\.[0-9]+)?)\s*/\s*10\*\*", readme_text)
    if not m:
        return None
    return float(m.group(1))


def build_governance_coverage(checklist_path: str, doc_paths):
    if not checklist_path or not doc_paths:
        return None

    data = json.loads(Path(checklist_path).read_text(encoding="utf-8"))
    doc_texts = {path: Path(path).read_text(encoding="utf-8") for path in doc_paths}

    checklist_ids = [item.get("id") for item in data.get("releaseChecklist", []) if item.get("id")]
    ci_refs = [item for item in data.get("ciReferences", []) if item]

    checklist_missing = []
    for item_id in checklist_ids:
        if not any(item_id in text for text in doc_texts.values()):
            checklist_missing.append(item_id)

    ci_ref_missing = []
    for ci_ref in ci_refs:
        if not any(ci_ref in text for text in doc_texts.values()):
            ci_ref_missing.append(ci_ref)

    covered_checklist = len(checklist_ids) - len(checklist_missing)
    covered_ci_refs = len(ci_refs) - len(ci_ref_missing)
    total = len(checklist_ids) + len(ci_refs)
    covered_total = covered_checklist + covered_ci_refs
    coverage_pct = (covered_total / total * 100.0) if total else 100.0

    return {
        "status": "PASS" if not checklist_missing and not ci_ref_missing else "WARN",
        "coveragePct": coverage_pct,
        "checklist": {
            "total": len(checklist_ids),
            "covered": covered_checklist,
            "missing": checklist_missing,
        },
        "ciReferences": {
            "total": len(ci_refs),
            "covered": covered_ci_refs,
            "missing": ci_ref_missing,
        },
        "documents": doc_paths,
    }


def render_markdown(
    rows,
    overall_score,
    completed_tasks,
    pending_tasks,
    workflow_name,
    needs_json,
    extended_perf_reports,
    governance_coverage,
):
    lines = []
    lines.append("# Boilerplate Kalite Skoru Raporu")
    lines.append("")
    lines.append(f"- Workflow: `{workflow_name}`")
    lines.append(f"- Genel skor: `{overall_score:.1f}/10`" if overall_score is not None else "- Genel skor: `n/a`")
    lines.append(f"- Tamamlanan backlog maddeleri: `{completed_tasks}`")
    lines.append(f"- Bekleyen backlog maddeleri: `{pending_tasks}`")
    lines.append("")
    lines.append("## Kriter Tablosu")
    lines.append("")
    lines.append("| Kriter | Skor | Seviye | Ozet |")
    lines.append("|---|---:|---|---|")
    for row in rows:
        lines.append(f"| {row['criterion']} | {row['score']:.1f} | {row['level']} | {row['summary']} |")
    lines.append("")
    lines.append("## Job Sonuclari")
    lines.append("")
    lines.append("| Job | Sonuc |")
    lines.append("|---|---|")
    for job_name, details in sorted(needs_json.items()):
        result = details.get("result", "unknown")
        lines.append(f"| {job_name} | {result} |")
    lines.append("")
    if governance_coverage:
        lines.append("## Governance Coverage")
        lines.append("")
        lines.append(f"- Status: `{governance_coverage['status']}`")
        lines.append(f"- Coverage: `{governance_coverage['coveragePct']:.1f}%`")
        lines.append(
            "- Checklist IDs: `{covered}/{total}`".format(
                covered=governance_coverage["checklist"]["covered"],
                total=governance_coverage["checklist"]["total"],
            )
        )
        lines.append(
            "- CI references: `{covered}/{total}`".format(
                covered=governance_coverage["ciReferences"]["covered"],
                total=governance_coverage["ciReferences"]["total"],
            )
        )
        missing_checklist = governance_coverage["checklist"]["missing"]
        missing_ci = governance_coverage["ciReferences"]["missing"]
        if missing_checklist:
            lines.append(f"- Missing checklist IDs: `{', '.join(missing_checklist)}`")
        if missing_ci:
            lines.append(f"- Missing CI references: `{', '.join(missing_ci)}`")
        lines.append("")
    if extended_perf_reports:
        lines.append("## Extended Perf")
        lines.append("")
        lines.append("| Scenario | Status | p95 (ms) | Failed Rate | Checks Rate |")
        lines.append("|---|---|---:|---:|---:|")
        for report in extended_perf_reports:
            metrics = report.get("metrics", {})
            lines.append(
                "| {scenario} | {status} | {p95:.2f} | {failed:.6f} | {checks:.6f} |".format(
                    scenario=report.get("scenario", "unknown"),
                    status="PASS" if report.get("passed") else "FAIL",
                    p95=float(metrics.get("http_req_duration_p95_ms", 0.0)),
                    failed=float(metrics.get("http_req_failed_rate", 0.0)),
                    checks=float(metrics.get("checks_rate", 0.0)),
                )
            )
        lines.append("")
    return "\n".join(lines) + "\n"


def main():
    parser = argparse.ArgumentParser(description="Generate boilerplate quality report markdown.")
    parser.add_argument("--readme", required=True)
    parser.add_argument("--plan", required=True)
    parser.add_argument("--out-md", required=True)
    parser.add_argument("--out-json")
    parser.add_argument("--workflow-name", default="unknown")
    parser.add_argument("--needs-json", default="{}")
    parser.add_argument("--extended-perf-json", action="append", default=[])
    parser.add_argument("--governance-checklist")
    parser.add_argument("--governance-doc", action="append", default=[])
    args = parser.parse_args()

    readme_text = Path(args.readme).read_text(encoding="utf-8")
    plan_text = Path(args.plan).read_text(encoding="utf-8")

    rows = parse_readme_scores(readme_text)
    completed_tasks, pending_tasks = parse_plan_tasks(plan_text)
    overall_score = parse_overall_score(readme_text)
    needs_json = json.loads(args.needs_json or "{}")
    extended_perf_reports = [
        json.loads(Path(path).read_text(encoding="utf-8")) for path in (args.extended_perf_json or [])
    ]
    governance_coverage = build_governance_coverage(args.governance_checklist, args.governance_doc or [])

    report_md = render_markdown(
        rows=rows,
        overall_score=overall_score,
        completed_tasks=completed_tasks,
        pending_tasks=pending_tasks,
        workflow_name=args.workflow_name,
        needs_json=needs_json,
        extended_perf_reports=extended_perf_reports,
        governance_coverage=governance_coverage,
    )
    Path(args.out_md).write_text(report_md, encoding="utf-8")

    if args.out_json:
        output = {
            "overallScore": overall_score,
            "criteria": rows,
            "completedTasks": completed_tasks,
            "pendingTasks": pending_tasks,
            "jobs": needs_json,
            "extendedPerf": extended_perf_reports,
            "governanceCoverage": governance_coverage,
        }
        Path(args.out_json).write_text(json.dumps(output, ensure_ascii=True, indent=2), encoding="utf-8")

    print("quality report generated")


if __name__ == "__main__":
    raise SystemExit(main())
