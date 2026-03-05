#!/usr/bin/env python3
import argparse
import json
from pathlib import Path


def load_metric(summary: dict, metric: str, field: str) -> float:
    try:
        value = summary["metrics"][metric]["values"][field]
        return float(value)
    except Exception as exc:  # noqa: BLE001
        raise ValueError(f"missing metric {metric}.{field}") from exc


def build_markdown(report: dict) -> str:
    scenario = report["scenario"]
    metrics = report["metrics"]
    thresholds = report["thresholds"]
    lines = [
        f"## k6 {scenario.capitalize()} Summary",
        "",
        f"- Status: `{'PASS' if report['passed'] else 'FAIL'}`",
        "",
        "| Metric | Value | Threshold |",
        "|---|---:|---:|",
        f"| `http_req_failed.rate` | {metrics['http_req_failed_rate']:.6f} | < {thresholds['http_req_failed_rate_max']:.6f} |",
        f"| `http_req_duration.p(95)` | {metrics['http_req_duration_p95_ms']:.2f} ms | < {thresholds['http_req_duration_p95_ms_max']:.2f} ms |",
        f"| `checks.rate` | {metrics['checks_rate']:.6f} | > {thresholds['checks_rate_min']:.6f} |",
        "",
    ]
    return "\n".join(lines) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(description="Parse k6 summary json into compact markdown/json report.")
    parser.add_argument("--input", required=True, help="Input k6 summary json path.")
    parser.add_argument("--scenario", required=True, choices=["smoke", "spike", "soak"])
    parser.add_argument("--out-json", required=True)
    parser.add_argument("--out-md", required=True)
    args = parser.parse_args()

    summary = json.loads(Path(args.input).read_text(encoding="utf-8"))

    failed_rate = load_metric(summary, "http_req_failed", "rate")
    p95_ms = load_metric(summary, "http_req_duration", "p(95)")
    checks_rate = load_metric(summary, "checks", "rate")

    thresholds = {
        "smoke": {
            "http_req_failed_rate_max": 0.02,
            "http_req_duration_p95_ms_max": 800.0,
            "checks_rate_min": 0.99,
        },
        "spike": {
            "http_req_failed_rate_max": 0.03,
            "http_req_duration_p95_ms_max": 1200.0,
            "checks_rate_min": 0.98,
        },
        "soak": {
            "http_req_failed_rate_max": 0.02,
            "http_req_duration_p95_ms_max": 900.0,
            "checks_rate_min": 0.99,
        },
    }[args.scenario]

    passed = (
        failed_rate < thresholds["http_req_failed_rate_max"]
        and p95_ms < thresholds["http_req_duration_p95_ms_max"]
        and checks_rate > thresholds["checks_rate_min"]
    )

    report = {
        "scenario": args.scenario,
        "passed": passed,
        "metrics": {
            "http_req_failed_rate": failed_rate,
            "http_req_duration_p95_ms": p95_ms,
            "checks_rate": checks_rate,
        },
        "thresholds": thresholds,
    }

    Path(args.out_json).write_text(json.dumps(report, ensure_ascii=True, indent=2) + "\n", encoding="utf-8")
    Path(args.out_md).write_text(build_markdown(report), encoding="utf-8")

    print(f"k6 summary parsed: scenario={args.scenario} passed={passed}")
    return 0 if passed else 1


if __name__ == "__main__":
    raise SystemExit(main())
