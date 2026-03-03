package services

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/akhilbojedla/reportdog/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxTagsPerReport = 20

// ReportService provides report-related business logic.
type ReportService struct {
	db *gorm.DB
}

// NewReportService creates a new ReportService.
func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{db: db}
}

// IngestRequest contains all data needed to ingest a report.
type IngestRequest struct {
	RawXML        string
	ExecutionName string
	Name          *string
	Source        string // "upload" or "api"
	Tags          []models.TagPair
}

// IngestReport parses XML, stores the report with raw XML, and returns the created report.
func (s *ReportService) IngestReport(req IngestRequest) (*models.TestReport, error) {
	parsed, err := ParseJUnitXML([]byte(req.RawXML))
	if err != nil {
		return nil, fmt.Errorf("XML parse error: %w", err)
	}

	hash := sha256.Sum256([]byte(req.RawXML))
	sha := fmt.Sprintf("%x", hash)

	// Deduplicate and cap tags.
	tags := deduplicateTags(req.Tags)
	if len(tags) > maxTagsPerReport {
		tags = tags[:maxTagsPerReport]
	}

	report := models.TestReport{
		ID:            uuid.New(),
		ExecutionName: req.ExecutionName,
		Name:          req.Name,
		Source:        req.Source,
		UploadedAt:    time.Now().UTC(),
		RawXML:        req.RawXML,
		RawXMLSHA256:  &sha,
		TotalTests:    parsed.TotalTests,
		Passed:        parsed.Passed,
		Failed:        parsed.Failed,
		Skipped:       parsed.Skipped,
		DurationSec:   parsed.DurationSec,
		Tags:          models.TagList(tags),
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&report).Error; err != nil {
			return fmt.Errorf("failed to create report: %w", err)
		}

		for i := range parsed.Suites {
			parsed.Suites[i].ReportID = report.ID
			if err := tx.Create(&parsed.Suites[i]).Error; err != nil {
				return fmt.Errorf("failed to create suite: %w", err)
			}
		}

		// Register known tag keys.
		s.registerTagKeys(tx, tags)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &report, nil
}

// registerTagKeys upserts tag keys into the known_tag_keys registry.
func (s *ReportService) registerTagKeys(tx *gorm.DB, tags []models.TagPair) {
	seen := make(map[string]bool)
	for _, t := range tags {
		if seen[t.Key] {
			continue
		}
		seen[t.Key] = true
		tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"last_seen_at"}),
		}).Create(&models.KnownTagKey{Key: t.Key, LastSeenAt: time.Now().UTC()})
	}
}

// deduplicateTags removes duplicate key:value pairs.
func deduplicateTags(tags []models.TagPair) []models.TagPair {
	seen := make(map[string]bool)
	out := make([]models.TagPair, 0, len(tags))
	for _, t := range tags {
		k := t.Key + "\x00" + t.Value
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, t)
	}
	return out
}

// --- List Reports ---

// ListReportsRequest holds filter/pagination parameters.
type ListReportsRequest struct {
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

// ListReportsResponse is a paginated list of reports.
type ListReportsResponse struct {
	Reports    []models.TestReport `json:"reports"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// ListReports returns a filtered, paginated list of reports.
func (s *ReportService) ListReports(req ListReportsRequest) (*ListReportsResponse, error) {
	query := s.db.Model(&models.TestReport{})

	if req.ExecutionName != "" {
		query = query.Where("execution_name ILIKE ?", "%"+req.ExecutionName+"%")
	}
	if req.Status == "passed" {
		query = query.Where("failed = 0")
	} else if req.Status == "failed" {
		query = query.Where("failed > 0")
	}
	if req.From != nil {
		query = query.Where("uploaded_at >= ?", req.From)
	}
	if req.To != nil {
		query = query.Where("uploaded_at <= ?", req.To)
	}

	// Free-text search across execution_name, name, and tags.
	if req.Search != "" {
		like := "%" + req.Search + "%"
		query = query.Where(
			"(execution_name ILIKE ? OR name ILIKE ? OR EXISTS (SELECT 1 FROM jsonb_array_elements(tags) elem WHERE elem->>'key' ILIKE ? OR elem->>'value' ILIKE ?))",
			like, like, like, like,
		)
	}

	// JSONB containment query for tag filtering.
	if req.TagKey != "" && req.TagValue != "" {
		fragment, _ := json.Marshal([]models.TagPair{{Key: req.TagKey, Value: req.TagValue}})
		query = query.Where("tags @> ?", string(fragment))
	} else if req.TagKey != "" {
		// Any tag with this key – use jsonb_path_exists.
		query = query.Where("EXISTS (SELECT 1 FROM jsonb_array_elements(tags) elem WHERE elem->>'key' = ?)", req.TagKey)
	} else if req.TagValue != "" {
		query = query.Where("EXISTS (SELECT 1 FROM jsonb_array_elements(tags) elem WHERE elem->>'value' = ?)", req.TagValue)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	sortCol := "uploaded_at"
	allowed := map[string]bool{
		"uploaded_at": true, "execution_name": true,
		"total_tests": true, "failed": true, "duration_sec": true,
	}
	if allowed[req.Sort] {
		sortCol = req.Sort
	}
	order := "desc"
	if req.Order == "asc" {
		order = "asc"
	}

	offset := (req.Page - 1) * req.PageSize
	var reports []models.TestReport
	if err := query.
		Order(sortCol + " " + order).
		Offset(offset).Limit(req.PageSize).
		Find(&reports).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize != 0 {
		totalPages++
	}

	return &ListReportsResponse{
		Reports:    reports,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetReport returns a single report with all related data.
func (s *ReportService) GetReport(id uuid.UUID) (*models.TestReport, error) {
	var report models.TestReport
	err := s.db.
		Preload("Suites", func(db *gorm.DB) *gorm.DB { return db.Order("name ASC") }).
		Preload("Suites.Cases", func(db *gorm.DB) *gorm.DB { return db.Order("name ASC") }).
		First(&report, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// GetRawXML returns the stored raw XML for a report.
func (s *ReportService) GetRawXML(id uuid.UUID) (string, error) {
	var report models.TestReport
	if err := s.db.Select("raw_xml").First(&report, "id = ?", id).Error; err != nil {
		return "", err
	}
	return report.RawXML, nil
}

// --- Tags ---

// AddTags appends tags to a report (up to 20 total).
func (s *ReportService) AddTags(reportID uuid.UUID, tags []models.TagPair) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var report models.TestReport
		if err := tx.Select("id", "tags").First(&report, "id = ?", reportID).Error; err != nil {
			return err
		}

		existing := make(map[string]bool)
		for _, t := range report.Tags {
			existing[t.Key+"\x00"+t.Value] = true
		}

		merged := append(models.TagList{}, report.Tags...)
		for _, t := range tags {
			k := t.Key + "\x00" + t.Value
			if existing[k] {
				continue
			}
			if len(merged) >= maxTagsPerReport {
				break
			}
			merged = append(merged, t)
			existing[k] = true
		}

		s.registerTagKeys(tx, tags)

		return tx.Model(&models.TestReport{}).Where("id = ?", reportID).
			Update("tags", merged).Error
	})
}

// RemoveTag removes a tag by key+value from a report's JSONB array.
func (s *ReportService) RemoveTag(reportID uuid.UUID, key, value string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var report models.TestReport
		if err := tx.Select("id", "tags").First(&report, "id = ?", reportID).Error; err != nil {
			return err
		}

		filtered := make(models.TagList, 0, len(report.Tags))
		found := false
		for _, t := range report.Tags {
			if t.Key == key && t.Value == value {
				found = true
				continue
			}
			filtered = append(filtered, t)
		}
		if !found {
			return gorm.ErrRecordNotFound
		}

		return tx.Model(&models.TestReport{}).Where("id = ?", reportID).
			Update("tags", filtered).Error
	})
}

// TagInfo is a tag with its usage count.
type TagInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Count int    `json:"count"`
}

// ListTags returns distinct tags across all reports, optionally filtered by key.
func (s *ReportService) ListTags(key string) ([]TagInfo, error) {
	baseQuery := `
		SELECT elem->>'key' AS key, elem->>'value' AS value, COUNT(*) AS count
		FROM test_reports, jsonb_array_elements(tags) elem
	`
	var args []interface{}
	if key != "" {
		baseQuery += " WHERE elem->>'key' = ? "
		args = append(args, key)
	}
	baseQuery += " GROUP BY elem->>'key', elem->>'value' ORDER BY elem->>'key', elem->>'value'"

	var tags []TagInfo
	err := s.db.Raw(baseQuery, args...).Scan(&tags).Error
	return tags, err
}

// ListKnownTagKeys returns known tag keys for autocomplete.
func (s *ReportService) ListKnownTagKeys() ([]string, error) {
	var keys []string
	err := s.db.Model(&models.KnownTagKey{}).
		Order("key ASC").Pluck("key", &keys).Error
	return keys, err
}

// --- History ---

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
func (s *ReportService) GetExecutionHistory(executionName string, limit int) ([]ExecutionHistoryItem, error) {
	var reports []models.TestReport
	err := s.db.
		Where("execution_name = ?", executionName).
		Order("uploaded_at DESC").
		Limit(limit).
		Find(&reports).Error
	if err != nil {
		return nil, err
	}

	items := make([]ExecutionHistoryItem, len(reports))
	for i, r := range reports {
		items[i] = ExecutionHistoryItem{
			ReportID:    r.ID,
			UploadedAt:  r.UploadedAt,
			TotalTests:  r.TotalTests,
			Passed:      r.Passed,
			Failed:      r.Failed,
			Skipped:     r.Skipped,
			DurationSec: r.DurationSec,
			Tags:        r.Tags,
		}
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
func (s *ReportService) GetTestHistory(executionName, testName string, limit int) ([]TestHistoryItem, error) {
	var items []TestHistoryItem
	err := s.db.Raw(`
		SELECT tr.id AS report_id, tr.uploaded_at, tc.status,
		       tc.duration_sec, tc.failure_msg, tc.failure_type
		FROM test_cases tc
		JOIN test_suites ts ON ts.id = tc.suite_id
		JOIN test_reports tr ON tr.id = ts.report_id
		WHERE tr.execution_name = ? AND tc.test_key = ?
		ORDER BY tr.uploaded_at DESC
		LIMIT ?
	`, executionName, testName, limit).Scan(&items).Error
	return items, err
}

// --- Stats ---

// Stats holds quick stats for the home page.
type Stats struct {
	TotalReports     int64   `json:"total_reports"`
	ReportsLast7Days int64   `json:"reports_last_7_days"`
	OverallPassRate  float64 `json:"overall_pass_rate"`
	TotalFailed      int64   `json:"total_failed_last_7_days"`
}

// GetStats returns quick stats for the home page.
func (s *ReportService) GetStats() (*Stats, error) {
	var stats Stats
	s.db.Model(&models.TestReport{}).Count(&stats.TotalReports)

	cutoff := time.Now().UTC().AddDate(0, 0, -7)
	s.db.Model(&models.TestReport{}).Where("uploaded_at >= ?", cutoff).Count(&stats.ReportsLast7Days)

	var agg struct {
		TotalTests  int64
		TotalPassed int64
		TotalFailed int64
	}
	s.db.Model(&models.TestReport{}).
		Where("uploaded_at >= ?", cutoff).
		Select("COALESCE(SUM(total_tests),0) as total_tests, COALESCE(SUM(passed),0) as total_passed, COALESCE(SUM(failed),0) as total_failed").
		Scan(&agg)

	if agg.TotalTests > 0 {
		stats.OverallPassRate = float64(agg.TotalPassed) / float64(agg.TotalTests) * 100
	}
	stats.TotalFailed = agg.TotalFailed
	return &stats, nil
}

// GetDistinctExecutionNames returns unique execution names for dropdowns.
func (s *ReportService) GetDistinctExecutionNames() ([]string, error) {
	var names []string
	err := s.db.Model(&models.TestReport{}).
		Distinct("execution_name").Order("execution_name ASC").
		Pluck("execution_name", &names).Error
	return names, err
}

// ParseTags parses comma-separated key:value tag strings.
func ParseTags(raw string) []models.TagPair {
	if raw == "" {
		return nil
	}
	var tags []models.TagPair
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, ":", 2)
		if len(kv) == 2 {
			tags = append(tags, models.TagPair{
				Key:   strings.TrimSpace(kv[0]),
				Value: strings.TrimSpace(kv[1]),
			})
		}
	}
	return tags
}
