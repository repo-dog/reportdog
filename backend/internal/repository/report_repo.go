package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/repo-dog/reportdog/backend/internal/models"
)

// ReportRepo handles all database operations for reports.
type ReportRepo struct {
	db *sql.DB
}

// NewReportRepo creates a new ReportRepo.
func NewReportRepo(db *sql.DB) *ReportRepo {
	return &ReportRepo{db: db}
}

// ---------- Create ----------

// CreateReport inserts a report, its suites, cases, and registers tag keys — all in one transaction.
func (r *ReportRepo) CreateReport(report *models.TestReport, suites []models.TestSuite) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	tagsVal, err := report.Tags.Value()
	if err != nil {
		return fmt.Errorf("marshal tags: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO test_reports
			(id, execution_name, name, source, uploaded_at, timestamp, raw_xml, raw_xml_sha256,
			 total_tests, passed, failed, skipped, duration_sec, tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		report.ID, report.ExecutionName, report.Name, report.Source,
		report.UploadedAt, report.Timestamp,
		report.RawXML, report.RawXMLSHA256,
		report.TotalTests, report.Passed, report.Failed, report.Skipped,
		report.DurationSec, tagsVal,
	)
	if err != nil {
		return fmt.Errorf("insert report: %w", err)
	}

	for _, s := range suites {
		_, err = tx.Exec(`
			INSERT INTO test_suites
				(id, report_id, name, suite_key, total_tests, passed, failed, skipped, duration_sec, timestamp)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			s.ID, report.ID, s.Name, s.SuiteKey,
			s.TotalTests, s.Passed, s.Failed, s.Skipped, s.DurationSec, s.Timestamp,
		)
		if err != nil {
			return fmt.Errorf("insert suite %q: %w", s.Name, err)
		}

		for _, c := range s.Cases {
			_, err = tx.Exec(`
				INSERT INTO test_cases
					(id, suite_id, name, test_key, class_name, duration_sec, status,
					 failure_msg, failure_type, failure_text, system_out, system_err)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
				c.ID, s.ID, c.Name, c.TestKey, c.ClassName,
				c.DurationSec, c.Status,
				c.FailureMsg, c.FailureType, c.FailureText,
				c.SystemOut, c.SystemErr,
			)
			if err != nil {
				return fmt.Errorf("insert case %q: %w", c.Name, err)
			}
		}
	}

	// Register known tag keys.
	seen := make(map[string]bool)
	for _, t := range report.Tags {
		if seen[t.Key] {
			continue
		}
		seen[t.Key] = true
		_, err = tx.Exec(`
			INSERT INTO known_tag_keys (key, last_seen_at)
			VALUES ($1, $2)
			ON CONFLICT (key) DO UPDATE SET last_seen_at = EXCLUDED.last_seen_at`,
			t.Key, time.Now().UTC(),
		)
		if err != nil {
			return fmt.Errorf("upsert tag key %q: %w", t.Key, err)
		}
	}

	return tx.Commit()
}

// ---------- List ----------

// ListReportsFilter holds filter/pagination parameters.
type ListReportsFilter struct {
	ExecutionName string
	TagKey        string
	TagValue      string
	Status        string
	Search        string
	From          *time.Time
	To            *time.Time
	Page          int
	PageSize      int
	Sort          string
	Order         string
}

// ListReportsResult is a paginated list of reports.
type ListReportsResult struct {
	Reports    []models.TestReport `json:"reports"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// ListReports returns a filtered, paginated list of reports.
func (r *ReportRepo) ListReports(f ListReportsFilter) (*ListReportsResult, error) {
	var where []string
	var args []interface{}
	argN := 0

	addArg := func(val interface{}) string {
		argN++
		args = append(args, val)
		p := fmt.Sprintf("$%d", argN)
		return p
	}

	if f.ExecutionName != "" {
		where = append(where, "execution_name ILIKE "+addArg("%"+f.ExecutionName+"%"))
	}
	if f.Status == "passed" {
		where = append(where, "failed = 0")
	} else if f.Status == "failed" {
		where = append(where, "failed > 0")
	}
	if f.From != nil {
		where = append(where, "uploaded_at >= "+addArg(f.From))
	}
	if f.To != nil {
		where = append(where, "uploaded_at <= "+addArg(f.To))
	}
	if f.Search != "" {
		p := addArg("%" + f.Search + "%")
		where = append(where, fmt.Sprintf(
			"(execution_name ILIKE %s OR name ILIKE %s OR EXISTS (SELECT 1 FROM jsonb_array_elements(tags) elem WHERE elem->>'key' ILIKE %s OR elem->>'value' ILIKE %s))",
			p, p, p, p,
		))
	}
	if f.TagKey != "" && f.TagValue != "" {
		fragment, _ := json.Marshal([]models.TagPair{{Key: f.TagKey, Value: f.TagValue}})
		where = append(where, "tags @> "+addArg(string(fragment))+"::jsonb")
	} else if f.TagKey != "" {
		where = append(where, "EXISTS (SELECT 1 FROM jsonb_array_elements(tags) elem WHERE elem->>'key' = "+addArg(f.TagKey)+")")
	} else if f.TagValue != "" {
		where = append(where, "EXISTS (SELECT 1 FROM jsonb_array_elements(tags) elem WHERE elem->>'value' = "+addArg(f.TagValue)+")")
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Count total.
	var total int64
	countSQL := "SELECT COUNT(*) FROM test_reports " + whereClause
	if err := r.db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}

	// Sort & order.
	sortCol := "uploaded_at"
	allowed := map[string]bool{
		"uploaded_at": true, "execution_name": true,
		"total_tests": true, "failed": true, "duration_sec": true,
	}
	if allowed[f.Sort] {
		sortCol = f.Sort
	}
	order := "DESC"
	if strings.EqualFold(f.Order, "asc") {
		order = "ASC"
	}

	offset := (f.Page - 1) * f.PageSize

	dataSQL := fmt.Sprintf(`
		SELECT id, execution_name, name, source, uploaded_at, timestamp,
			   total_tests, passed, failed, skipped, duration_sec, tags
		FROM test_reports %s
		ORDER BY %s %s
		OFFSET %d LIMIT %d`,
		whereClause, sortCol, order, offset, f.PageSize,
	)

	rows, err := r.db.Query(dataSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var reports []models.TestReport
	for rows.Next() {
		var rpt models.TestReport
		if err := rows.Scan(
			&rpt.ID, &rpt.ExecutionName, &rpt.Name, &rpt.Source,
			&rpt.UploadedAt, &rpt.Timestamp,
			&rpt.TotalTests, &rpt.Passed, &rpt.Failed, &rpt.Skipped,
			&rpt.DurationSec, &rpt.Tags,
		); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		reports = append(reports, rpt)
	}

	if reports == nil {
		reports = []models.TestReport{}
	}

	totalPages := int(total) / f.PageSize
	if int(total)%f.PageSize != 0 {
		totalPages++
	}

	return &ListReportsResult{
		Reports:    reports,
		Total:      total,
		Page:       f.Page,
		PageSize:   f.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ---------- Get ----------

// GetReport returns a single report with suites and cases.
func (r *ReportRepo) GetReport(id uuid.UUID) (*models.TestReport, error) {
	var rpt models.TestReport
	err := r.db.QueryRow(`
		SELECT id, execution_name, name, source, uploaded_at, timestamp,
			   raw_xml, raw_xml_sha256,
			   total_tests, passed, failed, skipped, duration_sec, tags
		FROM test_reports WHERE id = $1`, id,
	).Scan(
		&rpt.ID, &rpt.ExecutionName, &rpt.Name, &rpt.Source,
		&rpt.UploadedAt, &rpt.Timestamp,
		&rpt.RawXML, &rpt.RawXMLSHA256,
		&rpt.TotalTests, &rpt.Passed, &rpt.Failed, &rpt.Skipped,
		&rpt.DurationSec, &rpt.Tags,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("report not found")
		}
		return nil, fmt.Errorf("get report: %w", err)
	}

	// Load suites.
	suiteRows, err := r.db.Query(`
		SELECT id, report_id, name, suite_key, total_tests, passed, failed, skipped, duration_sec, timestamp
		FROM test_suites WHERE report_id = $1 ORDER BY name ASC`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("get suites: %w", err)
	}
	defer suiteRows.Close()

	suiteMap := make(map[uuid.UUID]*models.TestSuite)
	var suiteOrder []uuid.UUID
	for suiteRows.Next() {
		var s models.TestSuite
		if err := suiteRows.Scan(
			&s.ID, &s.ReportID, &s.Name, &s.SuiteKey,
			&s.TotalTests, &s.Passed, &s.Failed, &s.Skipped,
			&s.DurationSec, &s.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scan suite: %w", err)
		}
		s.Cases = []models.TestCase{}
		suiteMap[s.ID] = &s
		suiteOrder = append(suiteOrder, s.ID)
	}

	// Load cases for all suites.
	if len(suiteOrder) > 0 {
		suiteIDs := make([]interface{}, len(suiteOrder))
		placeholders := make([]string, len(suiteOrder))
		for i, sid := range suiteOrder {
			suiteIDs[i] = sid
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}

		caseSQL := fmt.Sprintf(`
			SELECT id, suite_id, name, test_key, class_name, duration_sec, status,
				   failure_msg, failure_type, failure_text, system_out, system_err
			FROM test_cases WHERE suite_id IN (%s) ORDER BY name ASC`,
			strings.Join(placeholders, ","),
		)

		caseRows, err := r.db.Query(caseSQL, suiteIDs...)
		if err != nil {
			return nil, fmt.Errorf("get cases: %w", err)
		}
		defer caseRows.Close()

		for caseRows.Next() {
			var c models.TestCase
			if err := caseRows.Scan(
				&c.ID, &c.SuiteID, &c.Name, &c.TestKey, &c.ClassName,
				&c.DurationSec, &c.Status,
				&c.FailureMsg, &c.FailureType, &c.FailureText,
				&c.SystemOut, &c.SystemErr,
			); err != nil {
				return nil, fmt.Errorf("scan case: %w", err)
			}
			if s, ok := suiteMap[c.SuiteID]; ok {
				s.Cases = append(s.Cases, c)
			}
		}
	}

	rpt.Suites = make([]models.TestSuite, 0, len(suiteOrder))
	for _, sid := range suiteOrder {
		rpt.Suites = append(rpt.Suites, *suiteMap[sid])
	}

	return &rpt, nil
}

// GetRawXML returns the raw XML for a report.
func (r *ReportRepo) GetRawXML(id uuid.UUID) (string, error) {
	var raw string
	err := r.db.QueryRow(`SELECT raw_xml FROM test_reports WHERE id = $1`, id).Scan(&raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("report not found")
		}
		return "", err
	}
	return raw, err
}

// ---------- Tags ----------

// UpdateTags replaces the tags column for a report.
func (r *ReportRepo) UpdateTags(reportID uuid.UUID, tags models.TagList) error {
	tagsVal, err := tags.Value()
	if err != nil {
		return err
	}
	res, err := r.db.Exec(`UPDATE test_reports SET tags = $1 WHERE id = $2`, tagsVal, reportID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("report not found")
	}
	return nil
}

// GetReportTags returns the tags for a report.
func (r *ReportRepo) GetReportTags(reportID uuid.UUID) (models.TagList, error) {
	var tags models.TagList
	err := r.db.QueryRow(`SELECT tags FROM test_reports WHERE id = $1`, reportID).Scan(&tags)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("report not found")
		}
		return nil, err
	}
	return tags, err
}

// RegisterTagKeys upserts tag keys.
func (r *ReportRepo) RegisterTagKeys(keys []string) error {
	for _, k := range keys {
		_, err := r.db.Exec(`
			INSERT INTO known_tag_keys (key, last_seen_at)
			VALUES ($1, $2)
			ON CONFLICT (key) DO UPDATE SET last_seen_at = EXCLUDED.last_seen_at`,
			k, time.Now().UTC(),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// TagInfo is a tag with its usage count.
type TagInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Count int    `json:"count"`
}

// ListTags returns all distinct tags, optionally filtered by key.
func (r *ReportRepo) ListTags(key string) ([]TagInfo, error) {
	q := `SELECT elem->>'key', elem->>'value', COUNT(*)
		  FROM test_reports, jsonb_array_elements(tags) elem`
	var args []interface{}
	if key != "" {
		args = append(args, key)
		q += ` WHERE elem->>'key' = $1`
	}
	q += ` GROUP BY elem->>'key', elem->>'value' ORDER BY elem->>'key', elem->>'value'`

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []TagInfo
	for rows.Next() {
		var t TagInfo
		if err := rows.Scan(&t.Key, &t.Value, &t.Count); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// ListKnownTagKeys returns known tag keys for autocomplete.
func (r *ReportRepo) ListKnownTagKeys() ([]string, error) {
	rows, err := r.db.Query(`SELECT key FROM known_tag_keys ORDER BY key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

// ---------- History ----------

// ExecutionHistoryItem is one row in execution history.
type ExecutionHistoryItem struct {
	ReportID    uuid.UUID      `json:"report_id"`
	UploadedAt  time.Time      `json:"uploaded_at"`
	TotalTests  int            `json:"total_tests"`
	Passed      int            `json:"passed"`
	Failed      int            `json:"failed"`
	Skipped     int            `json:"skipped"`
	DurationSec float64        `json:"duration_sec"`
	Tags        models.TagList `json:"tags,omitempty"`
}

// GetExecutionHistory returns the last N reports for an execution_name.
func (r *ReportRepo) GetExecutionHistory(executionName string, limit int) ([]ExecutionHistoryItem, error) {
	rows, err := r.db.Query(`
		SELECT id, uploaded_at, total_tests, passed, failed, skipped, duration_sec, tags
		FROM test_reports
		WHERE execution_name = $1
		ORDER BY uploaded_at DESC
		LIMIT $2`, executionName, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ExecutionHistoryItem
	for rows.Next() {
		var it ExecutionHistoryItem
		if err := rows.Scan(
			&it.ReportID, &it.UploadedAt, &it.TotalTests,
			&it.Passed, &it.Failed, &it.Skipped,
			&it.DurationSec, &it.Tags,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

// TestHistoryItem is one occurrence of a test across runs.
type TestHistoryItem struct {
	ReportID    uuid.UUID `json:"report_id"`
	UploadedAt  time.Time `json:"uploaded_at"`
	Status      string    `json:"status"`
	DurationSec float64   `json:"duration_sec"`
	FailureMsg  *string   `json:"failure_msg,omitempty"`
	FailureType *string   `json:"failure_type,omitempty"`
}

// GetTestHistory returns the last N results for a test in an execution series.
func (r *ReportRepo) GetTestHistory(executionName, testName string, limit int) ([]TestHistoryItem, error) {
	rows, err := r.db.Query(`
		SELECT tr.id, tr.uploaded_at, tc.status, tc.duration_sec, tc.failure_msg, tc.failure_type
		FROM test_cases tc
			JOIN test_suites ts ON ts.id = tc.suite_id
			JOIN test_reports tr ON tr.id = ts.report_id
		WHERE tr.execution_name = $1 AND tc.test_key = $2
		ORDER BY tr.uploaded_at DESC
		LIMIT $3`, executionName, testName, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TestHistoryItem
	for rows.Next() {
		var it TestHistoryItem
		if err := rows.Scan(
			&it.ReportID, &it.UploadedAt, &it.Status,
			&it.DurationSec, &it.FailureMsg, &it.FailureType,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

// ---------- Stats ----------

// Stats holds quick stats for the home page.
type Stats struct {
	TotalReports     int64   `json:"total_reports"`
	ReportsLast7Days int64   `json:"reports_last_7_days"`
	OverallPassRate  float64 `json:"overall_pass_rate"`
	TotalFailed      int64   `json:"total_failed_last_7_days"`
}

// GetStats returns quick stats for the home page.
func (r *ReportRepo) GetStats() (*Stats, error) {
	var stats Stats

	r.db.QueryRow(`SELECT COUNT(*) FROM test_reports`).Scan(&stats.TotalReports)

	cutoff := time.Now().UTC().AddDate(0, 0, -7)
	r.db.QueryRow(`SELECT COUNT(*) FROM test_reports WHERE uploaded_at >= $1`, cutoff).Scan(&stats.ReportsLast7Days)

	var totalTests, totalPassed, totalFailed int64
	r.db.QueryRow(`
		SELECT COALESCE(SUM(total_tests),0), COALESCE(SUM(passed),0), COALESCE(SUM(failed),0)
		FROM test_reports WHERE uploaded_at >= $1`, cutoff,
	).Scan(&totalTests, &totalPassed, &totalFailed)

	if totalTests > 0 {
		stats.OverallPassRate = float64(totalPassed) / float64(totalTests) * 100
	}
	stats.TotalFailed = totalFailed

	return &stats, nil
}

// ---------- Misc ----------

// GetDistinctExecutionNames returns unique execution names.
func (r *ReportRepo) GetDistinctExecutionNames() ([]string, error) {
	rows, err := r.db.Query(`SELECT DISTINCT execution_name FROM test_reports ORDER BY execution_name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, nil
}
