package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TagPair is a single key-value tag stored inside JSONB.
type TagPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// TagList is a slice of TagPair that maps to JSONB in Postgres.
type TagList []TagPair

// Scan implements sql.Scanner for reading JSONB from Postgres.
func (t *TagList) Scan(value interface{}) error {
	if value == nil {
		*t = TagList{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("TagList.Scan: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, t)
}

// Value implements driver.Valuer for writing JSONB to Postgres.
func (t TagList) Value() (driver.Value, error) {
	if t == nil {
		return "[]", nil
	}
	b, err := json.Marshal(t)
	return string(b), err
}

// TestReport represents a single ingested test run.
type TestReport struct {
	ID            uuid.UUID   `json:"id"`
	ExecutionName string      `json:"execution_name"`
	Name          *string     `json:"name,omitempty"`
	Source        string      `json:"source"`
	UploadedAt    time.Time   `json:"uploaded_at"`
	Timestamp     *time.Time  `json:"timestamp,omitempty"`
	RawXML        string      `json:"-"`
	RawXMLSHA256  *string     `json:"-"`
	TotalTests    int         `json:"total_tests"`
	Passed        int         `json:"passed"`
	Failed        int         `json:"failed"`
	Skipped       int         `json:"skipped"`
	DurationSec   float64     `json:"duration_sec"`
	Tags          TagList     `json:"tags"`
	Suites        []TestSuite `json:"suites,omitempty"`
}

// TestSuite represents a <testsuite> element.
type TestSuite struct {
	ID          uuid.UUID  `json:"id"`
	ReportID    uuid.UUID  `json:"report_id"`
	Name        string     `json:"name"`
	SuiteKey    string     `json:"suite_key"`
	TotalTests  int        `json:"total_tests"`
	Passed      int        `json:"passed"`
	Failed      int        `json:"failed"`
	Skipped     int        `json:"skipped"`
	DurationSec float64    `json:"duration_sec"`
	Timestamp   *time.Time `json:"timestamp,omitempty"`
	Cases       []TestCase `json:"cases,omitempty"`
}

// TestCase represents a <testcase> element.
type TestCase struct {
	ID          uuid.UUID `json:"id"`
	SuiteID     uuid.UUID `json:"suite_id"`
	Name        string    `json:"name"`
	TestKey     string    `json:"test_key"`
	ClassName   *string   `json:"classname,omitempty"`
	DurationSec float64   `json:"duration_sec"`
	Status      string    `json:"status"`
	FailureMsg  *string   `json:"failure_msg,omitempty"`
	FailureType *string   `json:"failure_type,omitempty"`
	FailureText *string   `json:"failure_text,omitempty"`
	SystemOut   *string   `json:"system_out,omitempty"`
	SystemErr   *string   `json:"system_err,omitempty"`
}

// KnownTagKey tracks distinct tag keys for autocomplete.
type KnownTagKey struct {
	Key        string    `json:"key"`
	LastSeenAt time.Time `json:"last_seen_at"`
}
