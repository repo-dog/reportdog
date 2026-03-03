package router

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/repo-dog/reportdog/backend/internal/handlers"
)

// Setup creates and configures the Gin engine.
func Setup(h *handlers.Handler, allowOrigin string, disableManualUpload bool) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{allowOrigin, "http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-Execution-Name", "X-Report-Name", "X-Tags"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.MaxMultipartMemory = 50 << 20

	api := r.Group("/api/v1")
	{
		if !disableManualUpload {
			api.POST("/reports/upload", h.UploadReport)
			api.POST("/reports/ingest", h.IngestRawXML)
		}
		api.GET("/reports", h.ListReports)
		api.GET("/reports/:id", h.GetReport)
		api.GET("/reports/:id/raw", h.GetRawXML)
		api.POST("/reports/:id/tags", h.AddTags)
		api.DELETE("/reports/:id/tags", h.RemoveTag)
		api.GET("/tags", h.ListTags)
		api.GET("/tags/keys", h.ListKnownTagKeys)
		api.GET("/executions/:executionName/reports", h.ExecutionHistory)
		api.GET("/executions/:executionName/tests/:testName/history", h.TestHistory)
		api.GET("/stats", h.GetStats)
		api.GET("/execution-names", h.GetExecutionNames)
		api.GET("/config", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"manual_upload_enabled": !disableManualUpload,
			})
		})
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Serve frontend static files if the public directory exists.
	publicDir := getEnvOrDefault("PUBLIC_DIR", "./public")
	if info, err := os.Stat(publicDir); err == nil && info.IsDir() {
		absDir, _ := filepath.Abs(publicDir)
		fileServer := http.FileServer(http.Dir(absDir))

		r.NoRoute(func(c *gin.Context) {
			// If the request looks like a file (has extension), try to serve it directly.
			requestPath := c.Request.URL.Path
			fullPath := filepath.Join(absDir, requestPath)

			// Check if file exists on disk
			if _, err := fs.Stat(os.DirFS(absDir), strings.TrimPrefix(requestPath, "/")); err == nil {
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}

			// SPA fallback: serve index.html for all other routes
			_ = fullPath
			c.File(filepath.Join(absDir, "index.html"))
		})
	}

	return r
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
