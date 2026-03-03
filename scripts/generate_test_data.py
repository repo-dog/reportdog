#!/usr/bin/env python3
"""
Generate ~500 diverse JUnit XML test reports and ingest them via the ReportDog API.

Usage:
    python3 scripts/generate_test_data.py [--api http://localhost:8080]
"""

import argparse
import random
import string
import time
import xml.etree.ElementTree as ET
from datetime import datetime, timedelta, timezone
import requests
from typing import Optional, Dict, List, Tuple

# ---------------------------------------------------------------------------
# Configuration knobs
# ---------------------------------------------------------------------------
NUM_REPORTS = 500
API_BASE = "http://localhost:8080/api/v1"

# Execution names (simulating different CI pipelines / repos)
EXECUTIONS = [
    "backend-unit-tests",
    "backend-integration-tests",
    "frontend-unit-tests",
    "frontend-e2e-tests",
    "api-contract-tests",
    "mobile-ios-tests",
    "mobile-android-tests",
    "data-pipeline-tests",
    "infra-smoke-tests",
    "auth-service-tests",
    "payment-service-tests",
    "notification-service-tests",
    "search-service-tests",
    "ml-model-validation",
    "performance-benchmarks",
]

# Suite name templates per execution category
SUITE_POOLS = {
    "backend": [
        "UserService", "OrderService", "ProductService", "AuthController",
        "PaymentGateway", "NotificationWorker", "CacheLayer", "DatabaseMigrations",
        "RateLimiter", "HealthCheck", "ConfigLoader", "EventBus",
    ],
    "frontend": [
        "LoginPage", "Dashboard", "Settings", "ProfileEditor",
        "SearchResults", "CartCheckout", "Notifications", "DataTable",
        "FormValidation", "Accessibility", "ThemeSwitcher", "Routing",
    ],
    "api": [
        "GET /users", "POST /users", "PUT /users/:id", "DELETE /users/:id",
        "GET /orders", "POST /orders", "GET /products", "POST /products",
        "GET /health", "POST /auth/login", "POST /auth/register", "GET /metrics",
    ],
    "mobile": [
        "SplashScreen", "OnboardingFlow", "PushNotifications", "DeepLinks",
        "OfflineMode", "BiometricAuth", "CameraCapture", "LocationServices",
    ],
    "data": [
        "ETL_Users", "ETL_Transactions", "DataQuality", "SchemaValidation",
        "Aggregation_Daily", "Aggregation_Weekly", "ExportCSV", "ImportBatch",
    ],
    "infra": [
        "DNS_Resolution", "SSL_Certs", "LoadBalancer", "CDN_Cache",
        "DatabaseReplication", "RedisCluster", "K8s_Pods", "Monitoring",
    ],
    "ml": [
        "ModelAccuracy", "FeatureImportance", "DataDrift", "PredictionLatency",
        "A_B_TestResults", "BiasDetection", "ModelVersioning", "TrainingPipeline",
    ],
    "perf": [
        "Latency_P50", "Latency_P99", "Throughput", "MemoryUsage",
        "CPUUtilization", "ConnectionPool", "GarbageCollection", "CacheHitRate",
    ],
}

# Test case name templates
CASE_TEMPLATES = [
    "test_{action}_{entity}",
    "should_{action}_{entity}_correctly",
    "it_{action}s_{entity}",
    "verify_{entity}_{action}",
    "{entity}_{action}_test",
]
ACTIONS = [
    "create", "read", "update", "delete", "validate", "process",
    "serialize", "deserialize", "render", "calculate", "transform",
    "filter", "sort", "paginate", "cache", "retry", "timeout",
    "authenticate", "authorize", "encrypt", "decrypt", "compress",
    "parse", "format", "notify", "schedule", "batch", "stream",
]
ENTITIES = [
    "user", "order", "product", "payment", "session", "config",
    "email", "report", "token", "file", "event", "metric",
    "invoice", "subscription", "webhook", "rule", "policy", "role",
]

# Tag pools
TAG_KEYS_VALUES = {
    "env":       ["dev", "staging", "production", "sandbox", "qa"],
    "branch":    ["main", "develop", "feature/auth", "feature/payments", "hotfix/login", "release/v2.1", "release/v2.2", "release/v3.0"],
    "runner":    ["github-actions", "jenkins", "gitlab-ci", "circleci", "buildkite", "self-hosted"],
    "team":      ["platform", "frontend", "backend", "mobile", "data", "infra", "security", "qa"],
    "priority":  ["critical", "high", "medium", "low"],
    "region":    ["us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"],
    "os":        ["linux", "macos", "windows"],
    "arch":      ["amd64", "arm64"],
    "node":      ["node-18", "node-20", "node-22"],
    "go":        ["go1.21", "go1.22", "go1.23"],
    "db":        ["postgres-15", "postgres-16", "mysql-8", "sqlite"],
    "trigger":   ["push", "pull_request", "schedule", "manual", "tag"],
    "component": ["api", "worker", "scheduler", "gateway", "cli"],
    "release":   ["v2.0.0", "v2.1.0", "v2.2.0", "v3.0.0-beta", "v3.0.0-rc1", "v3.0.0"],
}

# Failure message templates
FAILURE_MESSAGES = [
    "Expected {expected} but got {actual}",
    "AssertionError: {entity} should not be nil",
    "Timeout after 30s waiting for {entity}",
    "ConnectionRefusedError: could not connect to {entity}",
    "ValidationError: field '{field}' is required",
    "404 Not Found: /{entity}/{id}",
    "500 Internal Server Error: {entity} processing failed",
    "DeadlineExceeded: context deadline exceeded",
    "PermissionDenied: insufficient permissions for {action}",
    "RateLimitExceeded: too many requests",
    "OutOfMemoryError: heap space exhausted",
    "ConcurrencyError: optimistic lock failed on {entity}",
    "SchemaError: column '{field}' does not exist",
    "SerializationError: cannot marshal {entity}",
    "panic: runtime error: index out of range",
]

FAILURE_TYPES = [
    "AssertionError", "TimeoutError", "ConnectionError", "ValidationError",
    "HTTPError", "RuntimeError", "PermissionError", "ResourceError",
    "ConcurrencyError", "SerializationError",
]

STACK_TRACE_TEMPLATE = """    at {pkg}.{func}({file}:{line})
    at {pkg}.run({file}:{line2})
    at testing.tRunner(testing.go:1595)
    at runtime.goexit(asm_{arch}.s:1650)"""


def _suite_pool(exec_name: str) -> List[str]:
    """Pick the right suite pool based on execution name prefix."""
    for prefix, pool in SUITE_POOLS.items():
        if prefix in exec_name:
            return pool
    return SUITE_POOLS["backend"]


def _random_case_name() -> str:
    tpl = random.choice(CASE_TEMPLATES)
    return tpl.format(action=random.choice(ACTIONS), entity=random.choice(ENTITIES))


def _random_failure_msg() -> str:
    msg = random.choice(FAILURE_MESSAGES)
    return msg.format(
        expected=random.randint(1, 100),
        actual=random.randint(1, 100),
        entity=random.choice(ENTITIES),
        field=random.choice(["name", "email", "id", "status", "amount", "type"]),
        id=random.randint(1000, 9999),
        action=random.choice(ACTIONS),
    )


def _random_stack_trace() -> str:
    pkg = random.choice(["service", "handler", "repo", "worker", "controller", "middleware"])
    func = random.choice(ACTIONS).capitalize()
    file = f"{pkg}/{random.choice(ENTITIES)}.go"
    return STACK_TRACE_TEMPLATE.format(
        pkg=pkg, func=func, file=file,
        line=random.randint(20, 500),
        line2=random.randint(10, 200),
        arch=random.choice(["amd64", "arm64"]),
    )


def generate_junit_xml(
    suites_count: int,
    cases_per_suite: tuple[int, int],
    fail_rate: float,
    skip_rate: float,
    timestamp: Optional[datetime] = None,
) -> str:
    """Build a JUnit XML string with the given parameters."""
    root = ET.Element("testsuites")

    pool = random.sample(_suite_pool(random.choice(EXECUTIONS)), min(suites_count, 8))
    if len(pool) < suites_count:
        pool = pool * ((suites_count // len(pool)) + 1)
    suite_names = pool[:suites_count]

    total_t = total_p = total_f = total_s = 0
    total_dur = 0.0

    for idx, sname in enumerate(suite_names):
        n_cases = random.randint(*cases_per_suite)
        suite_el = ET.SubElement(root, "testsuite", name=sname)
        s_pass = s_fail = s_skip = 0
        s_dur = 0.0

        for _ in range(n_cases):
            cname = _random_case_name()
            dur = round(random.uniform(0.001, 15.0), 3)
            if random.random() < 0.08:        # occasional slow test
                dur = round(random.uniform(15.0, 120.0), 3)
            s_dur += dur

            tc = ET.SubElement(suite_el, "testcase", name=cname,
                               classname=f"{sname}.{cname.split('_')[0]}",
                               time=str(dur))

            roll = random.random()
            if roll < fail_rate:
                # failed or error
                ftype = random.choice(FAILURE_TYPES)
                fmsg = _random_failure_msg()
                if random.random() < 0.3:
                    err_el = ET.SubElement(tc, "error", type=ftype, message=fmsg)
                else:
                    err_el = ET.SubElement(tc, "failure", type=ftype, message=fmsg)
                err_el.text = _random_stack_trace()
                s_fail += 1
            elif roll < fail_rate + skip_rate:
                ET.SubElement(tc, "skipped", message="Skipped: " + random.choice([
                    "flaky test disabled", "not applicable on this OS",
                    "requires external service", "pending implementation",
                    "known issue #" + str(random.randint(100, 9999)),
                ]))
                s_skip += 1
            else:
                s_pass += 1

            # Occasional system-out / system-err
            if random.random() < 0.15:
                so = ET.SubElement(tc, "system-out")
                so.text = f"[INFO] {cname} completed in {dur}s"
            if random.random() < 0.05:
                se = ET.SubElement(tc, "system-err")
                se.text = f"[WARN] slow query detected: {random.randint(100,2000)}ms"

        suite_el.set("tests", str(s_pass + s_fail + s_skip))
        suite_el.set("failures", str(s_fail))
        suite_el.set("skipped", str(s_skip))
        suite_el.set("time", f"{s_dur:.3f}")
        suite_ts = timestamp or datetime.now(timezone.utc)
        # Add small per-suite jitter (0-60 min) so suites aren't identical
        suite_ts = suite_ts + timedelta(minutes=random.uniform(0, 60) * idx)
        suite_el.set("timestamp", suite_ts.isoformat())

        total_t += s_pass + s_fail + s_skip
        total_p += s_pass
        total_f += s_fail
        total_s += s_skip
        total_dur += s_dur

    root.set("tests", str(total_t))
    root.set("failures", str(total_f))
    root.set("time", f"{total_dur:.3f}")

    return ET.tostring(root, encoding="unicode", xml_declaration=True)


def random_tags(min_tags: int = 0, max_tags: int = 6) -> str:
    """Generate comma-separated key:value tag string."""
    n = random.randint(min_tags, max_tags)
    keys = random.sample(list(TAG_KEYS_VALUES.keys()), min(n, len(TAG_KEYS_VALUES)))
    pairs = []
    for k in keys:
        v = random.choice(TAG_KEYS_VALUES[k])
        pairs.append(f"{k}:{v}")
    return ",".join(pairs)


def ingest_report(api_base: str, exec_name: str, name: Optional[str], xml: str, tags: str) -> dict:
    headers = {
        "Content-Type": "application/xml",
        "X-Execution-Name": exec_name,
    }
    if name:
        headers["X-Report-Name"] = name
    if tags:
        headers["X-Tags"] = tags

    resp = requests.post(f"{api_base}/reports/ingest", data=xml.encode(), headers=headers, timeout=30)
    resp.raise_for_status()
    return resp.json()


def main():
    parser = argparse.ArgumentParser(description="Generate test data for ReportDog")
    parser.add_argument("--api", default=API_BASE, help="API base URL (default: %(default)s)")
    parser.add_argument("-n", "--count", type=int, default=NUM_REPORTS, help="Number of reports")
    args = parser.parse_args()

    api = args.api.rstrip("/")
    count = args.count

    # Verify API is reachable
    try:
        r = requests.get(f"{api.rsplit('/api', 1)[0]}/health", timeout=5)
        r.raise_for_status()
        print(f"✓ API healthy at {api}")
    except Exception as e:
        print(f"✗ Cannot reach API at {api}: {e}")
        return

    # Pre-compute execution name "runs" — simulate multiple runs of same pipeline
    # Each execution gets a variable number of reports spread over a time window
    exec_run_counts: Dict[str, int] = {}
    remaining = count
    for ex in EXECUTIONS:
        n = max(5, remaining // (len(EXECUTIONS) - len(exec_run_counts)))
        n = min(n, remaining)
        exec_run_counts[ex] = n
        remaining -= n
    # Distribute any leftover
    while remaining > 0:
        ex = random.choice(EXECUTIONS)
        exec_run_counts[ex] += 1
        remaining -= 1

    report_configs = []
    for exec_name, n_runs in exec_run_counts.items():
        # Simulate time series: reports spread over last 30 days
        base_time = datetime.now(timezone.utc) - timedelta(days=30)
        for i in range(n_runs):
            # Spread evenly + jitter
            offset_hours = (30 * 24 / n_runs) * i + random.uniform(-2, 2)
            ts = base_time + timedelta(hours=max(0, offset_hours))

            # Vary failure/skip rates per execution to create diverse patterns
            # Some executions are stable, some are flaky
            if "smoke" in exec_name or "health" in exec_name:
                fail_rate = random.choice([0.0, 0.0, 0.0, 0.02])
                skip_rate = 0.01
            elif "e2e" in exec_name:
                fail_rate = random.uniform(0.05, 0.25)
                skip_rate = random.uniform(0.02, 0.10)
            elif "integration" in exec_name:
                fail_rate = random.uniform(0.02, 0.15)
                skip_rate = random.uniform(0.01, 0.05)
            elif "perf" in exec_name or "benchmark" in exec_name:
                fail_rate = random.uniform(0.0, 0.10)
                skip_rate = random.uniform(0.0, 0.03)
            elif "ml" in exec_name or "model" in exec_name:
                fail_rate = random.uniform(0.0, 0.20)
                skip_rate = random.uniform(0.05, 0.15)
            else:
                fail_rate = random.uniform(0.0, 0.15)
                skip_rate = random.uniform(0.01, 0.08)

            # Vary suite/case counts
            suites = random.randint(1, 6)
            cases_per_suite = (
                random.choice([2, 3, 5, 8]),
                random.choice([10, 15, 20, 30, 50]),
            )

            # Build report name — sometimes None
            name = None
            if random.random() < 0.7:
                run_id = ''.join(random.choices(string.ascii_lowercase + string.digits, k=7))
                name = f"Run #{i+1} ({run_id})"

            report_configs.append({
                "exec_name": exec_name,
                "name": name,
                "suites": suites,
                "cases_per_suite": cases_per_suite,
                "fail_rate": fail_rate,
                "skip_rate": skip_rate,
                "tags": random_tags(1, 6),
                "timestamp": ts,
            })

    # Shuffle to simulate real-world interleaving
    random.shuffle(report_configs)

    print(f"Generating and ingesting {len(report_configs)} reports...")
    success = 0
    errors = 0
    start = time.time()

    for i, cfg in enumerate(report_configs):
        try:
            xml = generate_junit_xml(
                suites_count=cfg["suites"],
                cases_per_suite=cfg["cases_per_suite"],
                fail_rate=cfg["fail_rate"],
                skip_rate=cfg["skip_rate"],
                timestamp=cfg["timestamp"],
            )
            result = ingest_report(
                api_base=api,
                exec_name=cfg["exec_name"],
                name=cfg["name"],
                xml=xml,
                tags=cfg["tags"],
            )
            success += 1
            rid = result.get("report_id", "?")
            total = result.get("total_tests", 0)
            failed = result.get("failed", 0)
            status = "✓" if failed == 0 else "✗"
            if (i + 1) % 25 == 0 or i == 0:
                elapsed = time.time() - start
                rate = (i + 1) / elapsed if elapsed > 0 else 0
                print(f"  [{i+1:3d}/{len(report_configs)}] {status} {cfg['exec_name']:<35s} "
                      f"tests={total:<4d} failed={failed:<3d} "
                      f"({rate:.1f} reports/s)")
        except Exception as e:
            errors += 1
            print(f"  [{i+1:3d}/{len(report_configs)}] ERROR: {e}")

    elapsed = time.time() - start
    print(f"\nDone in {elapsed:.1f}s — {success} succeeded, {errors} failed")
    print(f"  Executions: {len(EXECUTIONS)}")
    print(f"  Tag keys:   {len(TAG_KEYS_VALUES)}")
    print(f"  Tag combos: {sum(len(v) for v in TAG_KEYS_VALUES.values())}")


if __name__ == "__main__":
    main()
