package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	ID            uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	ExecutionName string      `gorm:"type:varchar(255);not null;index:idx_exec_uploaded,priority:1" json:"execution_name"`
	Name          *string     `gorm:"type:varchar(255)" json:"name,omitempty"`
	Source        string      `gorm:"type:varchar(50);not null" json:"source"`
	UploadedAt    time.Time   `gorm:"type:timestamptz;not null;default:now();index:idx_exec_uploaded,priority:2,sort:desc" json:"uploaded_at"`
	RawXML        string      `gorm:"type:text;not null" json:"-"`
	RawXMLSHA256  *string     `gorm:"type:varchar(64);index" json:"-"`
	TotalTests    int         `gorm:"not null" json:"total_tests"`
	Passed        int         `gorm:"not null" json:"passed"`
	Failed        int         `gorm:"not null" json:"failed"`
	Skipped       int         `gorm:"not null" json:"skipped"`
	DurationSec   float64     `gorm:"type:double precision;not null" json:"duration_sec"`
	Tags          TagList     `gorm:"type:jsonb;not null;default:'[]'" json:"tags"`
	Suites        []TestSuite `gorm:"foreignKey:ReportID;constraint:OnDelete:CASCADE" json:"suites,omitempty"`
}

// BeforeCreate sets defaults before inserting a TestReport.
func (r *TestReport) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.UploadedAt.IsZero() {
		r.UploadedAt = time.Now().UTC()
	}
	if r.Tags == nil {
		r.Tags = TagList{}
	}
	return nil
}

// TestSuite represents a <testsuite> element.
type TestSuite struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	ReportID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"report_id"`
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	SuiteKey    string     `gorm:"type:varchar(255);not null;index" json:"suite_key"`
	TotalTests  int        `gorm:"not null" json:"total_tests"`
	Passed      int        `gorm:"not null" json:"passed"`
	Failed      int        `gorm:"not null" json:"failed"`
	Skipped     int        `gorm:"not null" json:"skipped"`
	DurationSec float64    `gorm:"type:double precision;not null" json:"duration_sec"`
	Timestamp   *time.Time `gorm:"type:timestamptz" json:"timestamp,omitempty"`
	Cases       []TestCase `gorm:"foreignKey:SuiteID;constraint:OnDelete:CASCADE" json:"cases,omitempty"`
}

// BeforeCreate sets defaults before inserting a TestSuite.
func (s *TestSuite) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// TestCase represents a <testcase> element.
type TestCase struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SuiteID     uuid.UUID `gorm:"type:uuid;not null;index" json:"suite_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	TestKey     string    `gorm:"type:varchar(255);not null;index" json:"test_key"`
	ClassName   *string   `gorm:"type:varchar(255)" json:"classname,omitempty"`
	DurationSec float64   `gorm:"type:double precision;not null" json:"duration_sec"`
	Status      string    `gorm:"type:varchar(50);not null" json:"status"`
	FailureMsg  *string   `gorm:"type:text" json:"failure_msg,omitempty"`
	FailureType *string   `gorm:"type:varchar(255)" json:"failure_type,omitempty"`
	FailureText *string   `gorm:"type:text" json:"failure_text,omitempty"`
	SystemOut   *string   `gorm:"type:text" json:"system_out,omitempty"`
	SystemErr   *string   `gorm:"type:text" json:"system_err,omitempty"`
}

// BeforeCreate sets defaults before inserting a TestCase.
func (c *TestCase) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// KnownTagKey tracks distinct tag keys for autocomplete.
type KnownTagKey struct {
	Key        string    `gorm:"type:varchar(255);primaryKey" json:"key"`
	LastSeenAt time.Time `gorm:"type:timestamptz;not null;default:now()" json:"last_seen_at"`
}
