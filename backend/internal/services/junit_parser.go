package services

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/akhilbojedla/reportdog/backend/internal/models"
	"github.com/google/uuid"
)

// --- JUnit XML structures ---

type junitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	TestSuites []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	XMLName    xml.Name         `xml:"testsuite"`
	Name       string           `xml:"name,attr"`
	Tests      string           `xml:"tests,attr"`
	Failures   string           `xml:"failures,attr"`
	Errors     string           `xml:"errors,attr"`
	Skipped    string           `xml:"skipped,attr"`
	Time       string           `xml:"time,attr"`
	Timestamp  string           `xml:"timestamp,attr"`
	TestCases  []junitTestCase  `xml:"testcase"`
	TestSuites []junitTestSuite `xml:"testsuite"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure"`
	Error     *junitError   `xml:"error"`
	Skipped   *junitSkipped `xml:"skipped"`
	SystemOut string        `xml:"system-out"`
	SystemErr string        `xml:"system-err"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

type junitError struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

type junitSkipped struct {
	Message string `xml:"message,attr"`
}

// ParseResult holds parsed data ready for persistence.
type ParseResult struct {
	TotalTests  int
	Passed      int
	Failed      int
	Skipped     int
	DurationSec float64
	Suites      []models.TestSuite
}

// ParseJUnitXML parses raw JUnit XML bytes into a ParseResult.
func ParseJUnitXML(data []byte) (*ParseResult, error) {
	trimmed := strings.TrimSpace(string(data))
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty XML data")
	}
	raw := []byte(trimmed)

	// Try <testsuites> wrapper
	var suites junitTestSuites
	if err := xml.Unmarshal(raw, &suites); err == nil && len(suites.TestSuites) > 0 {
		return buildResult(suites.TestSuites)
	}

	// Try single <testsuite>
	var suite junitTestSuite
	if err := xml.Unmarshal(raw, &suite); err == nil && (suite.Name != "" || len(suite.TestCases) > 0) {
		return buildResult([]junitTestSuite{suite})
	}

	return nil, fmt.Errorf("unable to parse XML as JUnit format")
}

func buildResult(jsuites []junitTestSuite) (*ParseResult, error) {
	result := &ParseResult{}
	allSuites := flattenSuites(jsuites)

	for _, js := range allSuites {
		suite := convertSuite(js)
		result.TotalTests += suite.TotalTests
		result.Passed += suite.Passed
		result.Failed += suite.Failed
		result.Skipped += suite.Skipped
		result.DurationSec += suite.DurationSec
		result.Suites = append(result.Suites, suite)
	}

	return result, nil
}

func flattenSuites(suites []junitTestSuite) []junitTestSuite {
	var flat []junitTestSuite
	for _, s := range suites {
		if len(s.TestCases) > 0 {
			flat = append(flat, s)
		}
		if len(s.TestSuites) > 0 {
			flat = append(flat, flattenSuites(s.TestSuites)...)
		}
		if len(s.TestCases) == 0 && len(s.TestSuites) == 0 && s.Name != "" {
			flat = append(flat, s)
		}
	}
	return flat
}

func convertSuite(js junitTestSuite) models.TestSuite {
	suite := models.TestSuite{
		ID:       uuid.New(),
		Name:     js.Name,
		SuiteKey: js.Name,
	}

	dur, _ := strconv.ParseFloat(js.Time, 64)
	suite.DurationSec = dur

	if js.Timestamp != "" {
		layouts := []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000",
			"2006-01-02 15:04:05",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, js.Timestamp); err == nil {
				utc := t.UTC()
				suite.Timestamp = &utc
				break
			}
		}
	}

	var passed, failed, skipped int
	for _, jc := range js.TestCases {
		tc := convertCase(jc)
		tc.SuiteID = suite.ID
		suite.Cases = append(suite.Cases, tc)

		switch tc.Status {
		case "passed":
			passed++
		case "failed", "error":
			failed++
		case "skipped":
			skipped++
		}
	}

	suite.TotalTests = len(js.TestCases)
	suite.Passed = passed
	suite.Failed = failed
	suite.Skipped = skipped

	return suite
}

func convertCase(jc junitTestCase) models.TestCase {
	tc := models.TestCase{
		ID:      uuid.New(),
		Name:    jc.Name,
		TestKey: jc.Name,
		Status:  "passed",
	}

	if jc.ClassName != "" {
		tc.ClassName = strPtr(jc.ClassName)
	}

	dur, _ := strconv.ParseFloat(jc.Time, 64)
	tc.DurationSec = dur

	switch {
	case jc.Failure != nil:
		tc.Status = "failed"
		tc.FailureMsg = strPtrNonEmpty(jc.Failure.Message)
		tc.FailureType = strPtrNonEmpty(jc.Failure.Type)
		tc.FailureText = strPtrNonEmpty(jc.Failure.Text)
	case jc.Error != nil:
		tc.Status = "error"
		tc.FailureMsg = strPtrNonEmpty(jc.Error.Message)
		tc.FailureType = strPtrNonEmpty(jc.Error.Type)
		tc.FailureText = strPtrNonEmpty(jc.Error.Text)
	case jc.Skipped != nil:
		tc.Status = "skipped"
		tc.FailureMsg = strPtrNonEmpty(jc.Skipped.Message)
	}

	tc.SystemOut = strPtrNonEmpty(jc.SystemOut)
	tc.SystemErr = strPtrNonEmpty(jc.SystemErr)

	return tc
}

func strPtr(s string) *string {
	return &s
}

func strPtrNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
