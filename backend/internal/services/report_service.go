package services

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/repo-dog/reportdog/backend/internal/models"
	"github.com/repo-dog/reportdog/backend/internal/repository"
)

const maxTagsPerReport = 20

// ReportService provides report-related business logic.
type ReportService struct {
	repo *repository.ReportRepo
}

// NewReportService creates a new ReportService.
func NewReportService(repo *repository.ReportRepo) *ReportService {
	return &ReportService{repo: repo}
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

	// Derive report-level timestamp from the earliest suite timestamp.
	for _, suite := range parsed.Suites {
		if suite.Timestamp != nil {
			if report.Timestamp == nil || suite.Timestamp.Before(*report.Timestamp) {
				report.Timestamp = suite.Timestamp
			}
		}
	}

	// Set report ID on suites.
	for i := range parsed.Suites {
		parsed.Suites[i].ReportID = report.ID
	}

	if err := s.repo.CreateReport(&report, parsed.Suites); err != nil {
		return nil, err
	}

	return &report, nil
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
	result, err := s.repo.ListReports(repository.ListReportsFilter{
		ExecutionName: req.ExecutionName,
		TagKey:        req.TagKey,
		TagValue:      req.TagValue,
		Status:        req.Status,
		Search:        req.Search,
		From:          req.From,
		To:            req.To,
		Page:          req.Page,
		PageSize:      req.PageSize,
		Sort:          req.Sort,
		Order:         req.Order,
	})
	if err != nil {
		return nil, err
	}
	return &ListReportsResponse{
		Reports:    result.Reports,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
}

// GetReport returns a single report with all related data.
func (s *ReportService) GetReport(id uuid.UUID) (*models.TestReport, error) {
	return s.repo.GetReport(id)
}

// GetRawXML returns the stored raw XML for a report.
func (s *ReportService) GetRawXML(id uuid.UUID) (string, error) {
	return s.repo.GetRawXML(id)
}

// --- Tags ---

// AddTags appends tags to a report (up to 20 total).
func (s *ReportService) AddTags(reportID uuid.UUID, tags []models.TagPair) error {
	existing, err := s.repo.GetReportTags(reportID)
	if err != nil {
		return err
	}

	seen := make(map[string]bool)
	for _, t := range existing {
		seen[t.Key+"\x00"+t.Value] = true
	}

	merged := append(models.TagList{}, existing...)
	var newKeys []string
	for _, t := range tags {
		k := t.Key + "\x00" + t.Value
		if seen[k] {
			continue
		}
		if len(merged) >= maxTagsPerReport {
			break
		}
		merged = append(merged, t)
		seen[k] = true
		newKeys = append(newKeys, t.Key)
	}

	if err := s.repo.UpdateTags(reportID, merged); err != nil {
		return err
	}
	if len(newKeys) > 0 {
		_ = s.repo.RegisterTagKeys(newKeys)
	}
	return nil
}

// RemoveTag removes a tag by key+value from a report's JSONB array.
func (s *ReportService) RemoveTag(reportID uuid.UUID, key, value string) error {
	existing, err := s.repo.GetReportTags(reportID)
	if err != nil {
		return err
	}

	filtered := make(models.TagList, 0, len(existing))
	found := false
	for _, t := range existing {
		if t.Key == key && t.Value == value {
			found = true
			continue
		}
		filtered = append(filtered, t)
	}
	if !found {
		return fmt.Errorf("tag not found")
	}

	return s.repo.UpdateTags(reportID, filtered)
}

// TagInfo re-exports from repository.
type TagInfo = repository.TagInfo

// ListTags returns distinct tags across all reports, optionally filtered by key.
func (s *ReportService) ListTags(key string) ([]TagInfo, error) {
	return s.repo.ListTags(key)
}

// ListKnownTagKeys returns known tag keys for autocomplete.
func (s *ReportService) ListKnownTagKeys() ([]string, error) {
	return s.repo.ListKnownTagKeys()
}

// --- History ---

// ExecutionHistoryItem re-exports from repository.
type ExecutionHistoryItem = repository.ExecutionHistoryItem

// GetExecutionHistory returns the last N reports for an execution_name.
func (s *ReportService) GetExecutionHistory(executionName string, limit int) ([]ExecutionHistoryItem, error) {
	return s.repo.GetExecutionHistory(executionName, limit)
}

// TestHistoryItem re-exports from repository.
type TestHistoryItem = repository.TestHistoryItem

// GetTestHistory returns the last N results for a test in an execution series.
func (s *ReportService) GetTestHistory(executionName, testName string, limit int) ([]TestHistoryItem, error) {
	return s.repo.GetTestHistory(executionName, testName, limit)
}

// --- Stats ---

// Stats re-exports from repository.
type Stats = repository.Stats

// GetStats returns quick stats for the home page.
func (s *ReportService) GetStats() (*Stats, error) {
	return s.repo.GetStats()
}

// GetDistinctExecutionNames returns unique execution names for dropdowns.
func (s *ReportService) GetDistinctExecutionNames() ([]string, error) {
	return s.repo.GetDistinctExecutionNames()
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
