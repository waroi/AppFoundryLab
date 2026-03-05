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


def main() -> int:
    parser = argparse.ArgumentParser(description="Compare k6 summary against baseline")
    parser.add_argument("--base", required=True, help="Base summary json path")
    parser.add_argument("--current", required=True, help="Current summary json path")
    parser.add_argument("--out-json", required=True, help="Comparison result json output path")
    parser.add_argument("--out-md", required=True, help="Comparison markdown output path")
    parser.add_argument("--p95-regression-pct", type=float, default=15.0)
    parser.add_argument("--p95-regression-abs-ms", type=float, default=60.0)
    parser.add_argument("--failed-regression-abs", type=float, default=0.003)
    parser.add_argument("--checks-regression-abs", type=float, default=0.003)
    args = parser.parse_args()

    base = json.loads(Path(args.base).read_text(encoding="utf-8"))
    cur = json.loads(Path(args.current).read_text(encoding="utf-8"))

    base_failed = load_metric(base, "http_req_failed", "rate")
    cur_failed = load_metric(cur, "http_req_failed", "rate")
    base_p95 = load_metric(base, "http_req_duration", "p(95)")
    cur_p95 = load_metric(cur, "http_req_duration", "p(95)")
    base_checks = load_metric(base, "checks", "rate")
    cur_checks = load_metric(cur, "checks", "rate")

    failed_limit = base_failed + args.failed_regression_abs
    p95_limit = (base_p95 * (1 + (args.p95_regression_pct / 100.0))) + args.p95_regression_abs_ms
    checks_limit = base_checks - args.checks_regression_abs

    failed_regressed = cur_failed > failed_limit
    p95_regressed = cur_p95 > p95_limit
    checks_regressed = cur_checks < checks_limit
    regressed = failed_regressed or p95_regressed or checks_regressed

    result = {
        "regressed": regressed,
        "limits": {
            "http_req_failed_rate_max": failed_limit,
            "http_req_duration_p95_ms_max": p95_limit,
            "checks_rate_min": checks_limit,
        },
        "base": {
            "http_req_failed_rate": base_failed,
            "http_req_duration_p95_ms": base_p95,
            "checks_rate": base_checks,
        },
        "current": {
            "http_req_failed_rate": cur_failed,
            "http_req_duration_p95_ms": cur_p95,
            "checks_rate": cur_checks,
        },
        "regressions": {
            "http_req_failed_rate": failed_regressed,
            "http_req_duration_p95_ms": p95_regressed,
            "checks_rate": checks_regressed,
        },
    }

    Path(args.out_json).write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")

    lines = [
        "# k6 Trend Comparison",
        "",
        "| Metric | Base | Current | Limit | Status |",
        "|---|---:|---:|---:|---|",
        f"| `http_req_failed.rate` | {base_failed:.6f} | {cur_failed:.6f} | <= {failed_limit:.6f} | {'FAIL' if failed_regressed else 'PASS'} |",
        f"| `http_req_duration.p(95)` | {base_p95:.2f} ms | {cur_p95:.2f} ms | <= {p95_limit:.2f} ms | {'FAIL' if p95_regressed else 'PASS'} |",
        f"| `checks.rate` | {base_checks:.6f} | {cur_checks:.6f} | >= {checks_limit:.6f} | {'FAIL' if checks_regressed else 'PASS'} |",
        "",
        f"Overall: {'FAIL (regression detected)' if regressed else 'PASS'}",
    ]
    Path(args.out_md).write_text("\n".join(lines) + "\n", encoding="utf-8")

    return 1 if regressed else 0


if __name__ == "__main__":
    raise SystemExit(main())
