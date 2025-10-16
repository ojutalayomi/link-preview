package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ExampleClient demonstrates how to use the Link Preview API
type ExampleClient struct {
	baseURL string
	client  *http.Client
}

// NewExampleClient creates a new example client
func NewExampleClient(baseURL string) *ExampleClient {
	return &ExampleClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchPreview demonstrates fetching a link preview
func (ec *ExampleClient) FetchPreview(url string) (*LinkPreviewResponse, error) {
	reqBody := LinkPreviewRequest{URL: url}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := ec.client.Post(
		ec.baseURL+"/preview",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var preview LinkPreviewResponse
	if err := json.Unmarshal(body, &preview); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &preview, nil
}

// HealthCheck demonstrates checking API health
func (ec *ExampleClient) HealthCheck() (map[string]interface{}, error) {
	resp, err := ec.client.Get(ec.baseURL + "/health")
	if err != nil {
		return nil, fmt.Errorf("failed to make health check request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read health check response: %v", err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		return nil, fmt.Errorf("failed to unmarshal health check response: %v", err)
	}

	return health, nil
}

// RunExamples demonstrates the API usage
// This function should be called after starting the server
func RunExamples() {
	fmt.Println("\n🚀 Running Link Preview API Examples...")
	fmt.Println("Make sure the API server is running on http://localhost:5465")
	fmt.Println("Run 'go run main.go' in another terminal first!")
	// fmt.Println("="*60)

	client := NewExampleClient("http://localhost:5465")

	// Test URLs to demonstrate different scenarios
	testURLs := []string{
		"https://github.com",
		"https://golang.org",
		"https://stackoverflow.com",
		"https://www.example.com",
	}

	// Test health check first
	fmt.Println("\n🏥 Testing Health Check...")
	health, err := client.HealthCheck()
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Health check successful: %+v\n", health)

	// Test link previews
	fmt.Println("\n🔗 Testing Link Previews...")
	for i, testURL := range testURLs {
		fmt.Printf("\n%d. Fetching preview for: %s\n", i+1, testURL)
		// fmt.Println("-"*50)

		preview, err := client.FetchPreview(testURL)
		if err != nil {
			fmt.Printf("❌ Failed to fetch preview: %v\n", err)
			continue
		}

		if preview.Error != "" {
			fmt.Printf("⚠️  Preview returned error: %s\n", preview.Error)
		} else {
			fmt.Printf("✅ Preview fetched successfully!\n")
		}

		fmt.Printf("📍 URL: %s\n", preview.URL)
		fmt.Printf("📝 Title: %s\n", preview.Title)
		fmt.Printf("📄 Description: %s\n", truncateString(preview.Description, 100))
		fmt.Printf("🖼️  Image: %s\n", preview.Image)
		fmt.Printf("🏢 Site Name: %s\n", preview.SiteName)

		// Add a small delay between requests to be respectful
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n\n🎉 Examples completed! Check the output above for results.")
	fmt.Println("💡 You can also test the API manually using curl:")
	fmt.Println(`   curl -X POST http://localhost:5465/preview \`)
	fmt.Println(`     -H "Content-Type: application/json" \`)
	fmt.Println(`     -d '{"url": "https://github.com"}'`)
}

// Helper function to truncate long strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
