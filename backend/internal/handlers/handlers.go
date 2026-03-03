package handlers

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/akhilbojedla/reportdog/backend/internal/models"
	"github.com/akhilbojedla/reportdog/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds all HTTP handler methods.
type Handler struct {
	svc *services.ReportService
}

// New creates a new Handler.
func New(svc *services.ReportService) *Handler {
	return &Handler{svc: svc}
}

// UploadReport handles multipart file upload of JUnit XML.
func (h *Handler) UploadReport(c *gin.Context) {
	execName := c.PostForm("execution_name")
	if execName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "execution_name is required"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required", "details": err.Error()})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file", "details": err.Error()})
		return
	}

	var name *string
	if n := c.PostForm("name"); n != "" {
		name = &n
	}

	tags := services.ParseTags(c.PostForm("tags"))

	report, err := h.svc.IngestReport(services.IngestRequest{
		RawXML:        string(data),
		ExecutionName: execName,
		Name:          name,
		Source:        "upload",
		Tags:          tags,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to ingest report", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"report_id":      report.ID,
		"execution_name": report.ExecutionName,
		"total_tests":    report.TotalTests,
		"passed":         report.Passed,
		"failed":         report.Failed,
		"skipped":        report.Skipped,
		"duration_sec":   report.DurationSec,
		"uploaded_at":    report.UploadedAt,
		"tags":           report.Tags,
	})
}

// IngestRawXML handles raw XML body ingest.
func (h *Handler) IngestRawXML(c *gin.Context) {
	execName := c.GetHeader("X-Execution-Name")
	if execName == "" {
		execName = c.Query("execution_name")
	}
	if execName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "execution_name is required (header X-Execution-Name or query param)"})
		return
	}

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body", "details": err.Error()})
		return
	}

	var name *string
	if n := c.GetHeader("X-Report-Name"); n != "" {
		name = &n
	} else if n := c.Query("name"); n != "" {
		name = &n
	}

	tagsRaw := c.GetHeader("X-Tags")
	if tagsRaw == "" {
		tagsRaw = c.Query("tags")
	}

	report, err := h.svc.IngestReport(services.IngestRequest{
		RawXML:        string(data),
		ExecutionName: execName,
		Name:          name,
		Source:        "api",
		Tags:          services.ParseTags(tagsRaw),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to ingest report", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"report_id":      report.ID,
		"execution_name": report.ExecutionName,
		"total_tests":    report.TotalTests,
		"passed":         report.Passed,
		"failed":         report.Failed,
		"skipped":        report.Skipped,
		"duration_sec":   report.DurationSec,
		"uploaded_at":    report.UploadedAt,
	})
}

// ListReports returns paginated & filtered reports.
func (h *Handler) ListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	req := services.ListReportsRequest{
		ExecutionName: c.Query("execution_name"),
		TagKey:        c.Query("tag_key"),
		TagValue:      c.Query("tag_value"),
		Status:        c.DefaultQuery("status", "all"),
		Search:        c.Query("q"),
		Page:          page,
		PageSize:      pageSize,
		Sort:          c.DefaultQuery("sort", "uploaded_at"),
		Order:         c.DefaultQuery("order", "desc"),
	}

	if from := c.Query("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			req.From = &t
		}
	}
	if to := c.Query("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			req.To = &t
		}
	}

	resp, err := h.svc.ListReports(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reports", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetReport returns a single report with suites, cases, and tags.
func (h *Handler) GetReport(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}
	report, err := h.svc.GetReport(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}
	c.JSON(http.StatusOK, report)
}

// GetRawXML returns the raw XML for a report.
func (h *Handler) GetRawXML(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}
	raw, err := h.svc.GetRawXML(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}
	c.Data(http.StatusOK, "application/xml", []byte(raw))
}

// AddTags adds tags to a report.
func (h *Handler) AddTags(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}
	var body struct {
		Tags []models.TagPair `json:"tags" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body", "details": err.Error()})
		return
	}
	if err := h.svc.AddTags(id, body.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add tags", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tags added"})
}

// RemoveTag removes a tag (by key+value) from a report.
func (h *Handler) RemoveTag(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}
	var body struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key and value required", "details": err.Error()})
		return
	}
	if err := h.svc.RemoveTag(reportID, body.Key, body.Value); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found on report"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tag removed"})
}

// ListTags returns distinct tags across all reports.
func (h *Handler) ListTags(c *gin.Context) {
	tags, err := h.svc.ListTags(c.Query("key"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tags"})
		return
	}
	c.JSON(http.StatusOK, tags)
}

// ListKnownTagKeys returns known tag keys for autocomplete.
func (h *Handler) ListKnownTagKeys(c *gin.Context) {
	keys, err := h.svc.ListKnownTagKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tag keys"})
		return
	}
	c.JSON(http.StatusOK, keys)
}

// ExecutionHistory returns the last N reports for an execution name.
func (h *Handler) ExecutionHistory(c *gin.Context) {
	execName := c.Param("executionName")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 500 {
		limit = 50
	}
	items, err := h.svc.GetExecutionHistory(execName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get execution history"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// TestHistory returns the last N results for a specific test in an execution.
func (h *Handler) TestHistory(c *gin.Context) {
	execName := c.Param("executionName")
	testName := c.Param("testName")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit < 1 || limit > 500 {
		limit = 100
	}
	items, err := h.svc.GetTestHistory(execName, testName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get test history"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetStats returns quick stats for the home page.
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetExecutionNames returns distinct execution names.
func (h *Handler) GetExecutionNames(c *gin.Context) {
	names, err := h.svc.GetDistinctExecutionNames()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get execution names"})
		return
	}
	c.JSON(http.StatusOK, names)
}
