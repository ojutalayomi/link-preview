package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LinkPreviewRequest represents the incoming request structure
// Contains the URL for which we want to fetch the preview
type LinkPreviewRequest struct {
	URL string `json:"url" binding:"required"` // The URL to fetch preview for
}

// LinkPreviewResponse represents the response structure
// Contains all the metadata extracted from the webpage
type LinkPreviewResponse struct {
	URL         string `json:"url"`             // Original URL
	Title       string `json:"title"`           // Page title
	Description string `json:"description"`     // Page description (meta description)
	Image       string `json:"image"`           // Preview image URL
	SiteName    string `json:"site_name"`       // Site name (og:site_name)
	Error       string `json:"error,omitempty"` // Error message if any
}

// MetaExtractor handles the extraction of metadata from HTML content
type MetaExtractor struct {
	client *http.Client
}

// NewMetaExtractor creates a new instance of MetaExtractor
// with a configured HTTP client that has reasonable timeouts
func NewMetaExtractor() *MetaExtractor {
	return &MetaExtractor{
		client: &http.Client{
			Timeout: 10 * time.Second, // Set timeout to prevent hanging requests
		},
	}
}

// FetchLinkPreview fetches and extracts metadata from a given URL
// This function runs in a goroutine to handle multiple requests concurrently
func (me *MetaExtractor) FetchLinkPreview(ctx context.Context, targetURL string, resultChan chan<- LinkPreviewResponse) {
	// Defer sending result to channel to ensure we always send a response
	var result LinkPreviewResponse
	defer func() {
		select {
		case resultChan <- result:
		case <-ctx.Done():
			// Context cancelled, don't send result
		}
	}()

	// Initialize result with the original URL
	result.URL = targetURL

	// Validate URL format
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		result.Error = fmt.Sprintf("Invalid URL format: %v", err)
		return
	}

	// Ensure URL has a scheme (http/https)
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
		targetURL = parsedURL.String()
		result.URL = targetURL
	}

	// Create HTTP request with context for cancellation support
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create request: %v", err)
		return
	}

	// Set User-Agent to mimic a real browser (some sites block requests without it)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Execute the HTTP request
	resp, err := me.client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to fetch URL: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check for successful HTTP status
	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status)
		return
	}

	// Read response body with size limit to prevent memory issues
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // Limit to 1MB
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read response body: %v", err)
		return
	}

	// Extract metadata from HTML content
	me.extractMetadata(string(body), &result)
}

// extractMetadata parses HTML content and extracts relevant metadata
// Uses regular expressions to find Open Graph tags and standard HTML meta tags
func (me *MetaExtractor) extractMetadata(htmlContent string, result *LinkPreviewResponse) {
	// Convert to lowercase for case-insensitive matching
	lowerHTML := strings.ToLower(htmlContent)

	// Extract title - try <title> tag first, then og:title
	if title := me.extractTag(htmlContent, `<title[^>]*>([^<]*)</title>`); title != "" {
		result.Title = strings.TrimSpace(title)
	}
	if ogTitle := me.extractMetaContent(lowerHTML, "og:title"); ogTitle != "" {
		result.Title = strings.TrimSpace(ogTitle)
	}

	// Extract description - try meta description first, then og:description
	if desc := me.extractMetaContent(lowerHTML, "description"); desc != "" {
		result.Description = strings.TrimSpace(desc)
	}
	if ogDesc := me.extractMetaContent(lowerHTML, "og:description"); ogDesc != "" {
		result.Description = strings.TrimSpace(ogDesc)
	}

	// Extract image URL from og:image
	if ogImage := me.extractMetaContent(lowerHTML, "og:image"); ogImage != "" {
		result.Image = strings.TrimSpace(ogImage)
	}

	// Extract site name from og:site_name
	if siteName := me.extractMetaContent(lowerHTML, "og:site_name"); siteName != "" {
		result.SiteName = strings.TrimSpace(siteName)
	}
}

// extractTag extracts content from HTML tags using regex
func (me *MetaExtractor) extractTag(html, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractMetaContent extracts content from meta tags (both name and property attributes)
func (me *MetaExtractor) extractMetaContent(html, metaName string) string {
	// Try meta tag with name attribute
	pattern1 := fmt.Sprintf(`<meta[^>]*name=["']%s["'][^>]*content=["']([^"']*)["']`, regexp.QuoteMeta(metaName))
	if content := me.extractTag(html, pattern1); content != "" {
		return content
	}

	// Try meta tag with property attribute (for Open Graph tags)
	pattern2 := fmt.Sprintf(`<meta[^>]*property=["']%s["'][^>]*content=["']([^"']*)["']`, regexp.QuoteMeta(metaName))
	if content := me.extractTag(html, pattern2); content != "" {
		return content
	}

	// Try reverse order (content before name/property)
	pattern3 := fmt.Sprintf(`<meta[^>]*content=["']([^"']*)["'][^>]*name=["']%s["']`, regexp.QuoteMeta(metaName))
	if content := me.extractTag(html, pattern3); content != "" {
		return content
	}

	pattern4 := fmt.Sprintf(`<meta[^>]*content=["']([^"']*)["'][^>]*property=["']%s["']`, regexp.QuoteMeta(metaName))
	return me.extractTag(html, pattern4)
}

// handleLinkPreview is the main HTTP handler for the /preview endpoint
// It processes the request, validates input, and coordinates the goroutine-based preview fetching
func handleLinkPreview(extractor *MetaExtractor) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse JSON request body
		var req LinkPreviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format. Expected JSON with 'url' field.",
				"details": err.Error(),
			})
			return
		}

		// Validate that URL is not empty
		if strings.TrimSpace(req.URL) == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "URL cannot be empty",
			})
			return
		}

		// Create context with timeout for the goroutine
		// This ensures that long-running requests don't hang indefinitely
		ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
		defer cancel()

		// Create channel to receive the result from the goroutine
		// Buffered channel ensures the goroutine doesn't block when sending result
		resultChan := make(chan LinkPreviewResponse, 1)

		// Launch goroutine to fetch link preview concurrently
		// This allows the server to handle multiple requests simultaneously
		go extractor.FetchLinkPreview(ctx, strings.TrimSpace(req.URL), resultChan)

		// Wait for either the result or context timeout
		select {
		case result := <-resultChan:
			// Successfully received result from goroutine
			if result.Error != "" {
				// Return error response but with 200 status as we successfully processed the request
				c.JSON(http.StatusOK, result)
			} else {
				// Return successful preview data
				c.Header("Cache-Control", "public, max-age=3600, s-maxage=3600, stale-while-revalidate=86400")
				c.JSON(http.StatusOK, result)
			}
		case <-ctx.Done():
			// Request timed out or was cancelled
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "Request timed out while fetching link preview",
				"url":   req.URL,
			})
		}
	}
}

// Config holds server configuration
type Config struct {
	AllowedOrigins []string
	Port           string
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	// Get allowed origins from environment variable
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string

	if allowedOrigins != "" {
		// Split by comma and trim spaces
		for _, origin := range strings.Split(allowedOrigins, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				origins = append(origins, origin)
			}
		}
	}

	// Default to allowing common development origins if none specified
	if len(origins) == 0 {
		origins = []string{"https://localhost:3000", "http://localhost:3000", "http://localhost:5173"}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":5465"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return &Config{
		AllowedOrigins: origins,
		Port:           port,
	}
}

// isOriginAllowed checks if the given origin is in the allowed list
func (c *Config) isOriginAllowed(origin string) bool {
	for _, allowed := range c.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// setupRoutes configures all the API routes
func setupRoutes(extractor *MetaExtractor, config *Config) *gin.Engine {
	// Create Gin router with default middleware (logger and recovery)
	router := gin.Default()
	fmt.Printf("\nGIN_MODE is %s\n", os.Getenv("ALLOWED_ORIGINS"))
	gin.SetMode(os.Getenv("GIN_MODE"))

	// Add CORS middleware with configurable allowed origins
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Set CORS headers based on configuration
		if origin != "" {
			if config.isOriginAllowed(origin) {
				// Allow specific origin (required when credentials are used)
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
			} else if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
				// Only use wildcard if no specific origin is provided and wildcard is allowed
				c.Header("Access-Control-Allow-Origin", "*")
			}
		} else if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
			// No origin header, use wildcard if configured
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "link-preview-api",
			"timestamp": time.Now().UTC(),
		})
	})

	// Main endpoint for fetching link previews
	router.POST("/preview", handleLinkPreview(extractor))

	// API documentation endpoint
	router.GET("/", func(c *gin.Context) {
		docs := map[string]interface{}{
			"service":     "Link Preview API",
			"version":     "1.0.0",
			"description": "API for fetching website metadata and link previews",
			"endpoints": map[string]interface{}{
				"POST /preview": map[string]interface{}{
					"description": "Fetch link preview for a given URL",
					"body": map[string]string{
						"url": "The URL to fetch preview for (required)",
					},
					"response": map[string]string{
						"url":         "Original URL",
						"title":       "Page title",
						"description": "Page description",
						"image":       "Preview image URL",
						"site_name":   "Site name",
						"error":       "Error message (if any)",
					},
				},
				"GET /health": "Health check endpoint",
			},
			"examples": map[string]interface{}{
				"request": map[string]string{
					"url": "https://github.com",
				},
			},
		}

		c.JSON(http.StatusOK, docs)
	})

	return router
}

func main() {
	// Create configuration
	config := NewConfig()

	// Create meta extractor instance
	extractor := NewMetaExtractor()

	// Setup routes with configuration
	router := setupRoutes(extractor, config)

	fmt.Printf("🚀 Link Preview API server starting on port %s\n", config.Port)
	fmt.Printf("🌐 Allowed origins: %v\n", config.AllowedOrigins)
	fmt.Println("📝 API Documentation available at: /")
	fmt.Println("🏥 Health check available at: /health")
	fmt.Println("🔗 Preview endpoint: POST /preview")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("  ALLOWED_ORIGINS: Comma-separated list of allowed origins (default: *)")
	fmt.Println("  PORT: Server port (default: 5465)")
	fmt.Println("  GIN_MODE: Gin mode (debug, release, test)")

	// Start server
	if err := router.Run(config.Port); err != nil {
		fmt.Printf("❌ Failed to start server: %v\n", err)
	}
}
