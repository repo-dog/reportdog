-- 002_create_indexes.sql
-- Creates indexes for common query patterns.

CREATE INDEX IF NOT EXISTS idx_exec_uploaded      ON test_reports  (execution_name, uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_reports_sha   ON test_reports  (raw_xml_sha256);
CREATE INDEX IF NOT EXISTS idx_test_reports_tags  ON test_reports  USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_test_suites_report ON test_suites   (report_id);
CREATE INDEX IF NOT EXISTS idx_test_suites_key    ON test_suites   (suite_key);
CREATE INDEX IF NOT EXISTS idx_test_cases_suite   ON test_cases    (suite_id);
CREATE INDEX IF NOT EXISTS idx_test_cases_key     ON test_cases    (test_key);
