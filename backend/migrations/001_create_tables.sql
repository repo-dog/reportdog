-- 001_create_tables.sql
-- Creates the core ReportDog schema: reports, suites, cases, and tag keys.

CREATE TABLE IF NOT EXISTS test_reports (
    id              UUID PRIMARY KEY,
    execution_name  VARCHAR(255) NOT NULL,
    name            VARCHAR(255),
    source          VARCHAR(50) NOT NULL,
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    timestamp       TIMESTAMPTZ,
    raw_xml         TEXT NOT NULL,
    raw_xml_sha256  VARCHAR(64),
    total_tests     INT NOT NULL,
    passed          INT NOT NULL,
    failed          INT NOT NULL,
    skipped         INT NOT NULL,
    duration_sec    DOUBLE PRECISION NOT NULL,
    tags            JSONB NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS test_suites (
    id              UUID PRIMARY KEY,
    report_id       UUID NOT NULL REFERENCES test_reports(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    suite_key       VARCHAR(255) NOT NULL,
    total_tests     INT NOT NULL,
    passed          INT NOT NULL,
    failed          INT NOT NULL,
    skipped         INT NOT NULL,
    duration_sec    DOUBLE PRECISION NOT NULL,
    timestamp       TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS test_cases (
    id              UUID PRIMARY KEY,
    suite_id        UUID NOT NULL REFERENCES test_suites(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    test_key        VARCHAR(255) NOT NULL,
    class_name      VARCHAR(255),
    duration_sec    DOUBLE PRECISION NOT NULL,
    status          VARCHAR(50) NOT NULL,
    failure_msg     TEXT,
    failure_type    VARCHAR(255),
    failure_text    TEXT,
    system_out      TEXT,
    system_err      TEXT
);

CREATE TABLE IF NOT EXISTS known_tag_keys (
    key             VARCHAR(255) PRIMARY KEY,
    last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
