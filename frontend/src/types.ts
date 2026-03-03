export interface Tag {
  key: string;
  value: string;
}

export interface TestCase {
  id: string;
  suite_id: string;
  name: string;
  test_key: string;
  classname?: string;
  duration_sec: number;
  status: 'passed' | 'failed' | 'error' | 'skipped';
  failure_msg?: string;
  failure_type?: string;
  failure_text?: string;
  system_out?: string;
  system_err?: string;
}

export interface TestSuite {
  id: string;
  report_id: string;
  name: string;
  suite_key: string;
  total_tests: number;
  passed: number;
  failed: number;
  skipped: number;
  duration_sec: number;
  timestamp?: string;
  cases: TestCase[];
}

export interface TestReport {
  id: string;
  execution_name: string;
  name?: string;
  source: string;
  uploaded_at: string;
  total_tests: number;
  passed: number;
  failed: number;
  skipped: number;
  duration_sec: number;
  suites?: TestSuite[];
  tags?: Tag[];
}

export interface ListReportsResponse {
  reports: TestReport[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ExecutionHistoryItem {
  report_id: string;
  uploaded_at: string;
  total_tests: number;
  passed: number;
  failed: number;
  skipped: number;
  duration_sec: number;
  tags?: Tag[];
}

export interface TestHistoryItem {
  report_id: string;
  uploaded_at: string;
  status: string;
  duration_sec: number;
  failure_msg?: string;
  failure_type?: string;
}

export interface Stats {
  total_reports: number;
  reports_last_7_days: number;
  overall_pass_rate: number;
  total_failed_last_7_days: number;
}
