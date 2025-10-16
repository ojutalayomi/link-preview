# Link Preview API

A high-performance Golang REST API built with Gin framework that fetches website metadata and generates link previews. The API uses goroutines to handle concurrent requests efficiently and extracts Open Graph metadata, page titles, descriptions, and preview images from web pages.

## Features

- ✨ **Fast & Concurrent**: Uses goroutines to handle multiple preview requests simultaneously
- 🛡️ **Robust Error Handling**: Comprehensive error handling with timeouts and validation
- 🌐 **CORS Support**: Cross-origin resource sharing enabled for web applications
- 📱 **Mobile-Friendly**: Proper User-Agent headers to ensure compatibility with all websites
- 🔍 **Rich Metadata**: Extracts Open Graph tags, meta descriptions, titles, and images
- 📚 **Self-Documenting**: Built-in API documentation endpoint
- 🏥 **Health Checks**: Health check endpoint for monitoring

## Installation

### Prerequisites
- Go 1.19 or higher
- Git

### Setup

1. **Clone or navigate to the project directory:**
   ```bash
   cd link-preview-api
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the application:**
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:5465`

## API Endpoints

### 1. Get Link Preview
**POST** `/preview`

Fetches metadata and generates a preview for the provided URL.

#### Request Body
```json
{
  "url": "https://example.com"
}
```

#### Response
```json
{
  "url": "https://example.com",
  "title": "Example Domain",
  "description": "This domain is for use in illustrative examples in documents.",
  "image": "https://example.com/image.jpg",
  "site_name": "Example",
  "error": ""
}
```

#### Error Response
```json
{
  "url": "https://invalid-url.com",
  "title": "",
  "description": "",
  "image": "",
  "site_name": "",
  "error": "Failed to fetch URL: no such host"
}
```

### 2. Health Check
**GET** `/health`

Returns the health status of the API.

#### Response
```json
{
  "status": "healthy",
  "service": "link-preview-api",
  "timestamp": "2024-06-14T10:36:27Z"
}
```

### 3. API Documentation
**GET** `/`

Returns comprehensive API documentation with examples.

## Usage Examples

### Using cURL

```bash
# Basic link preview
curl -X POST http://localhost:5465/preview \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com"}'

# Health check
curl http://localhost:5465/health

# API documentation
curl http://localhost:5465/
```

### Using JavaScript (Fetch)

```javascript
const fetchLinkPreview = async (url) => {
  try {
    const response = await fetch('http://localhost:5465/preview', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ url: url }),
    });
    
    const preview = await response.json();
    console.log('Link preview:', preview);
    return preview;
  } catch (error) {
    console.error('Error fetching preview:', error);
  }
};

// Usage
fetchLinkPreview('https://www.example.com');
```

### Using Go Client

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

type PreviewRequest struct {
    URL string `json:"url"`
}

type PreviewResponse struct {
    URL         string `json:"url"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Image       string `json:"image"`
    SiteName    string `json:"site_name"`
    Error       string `json:"error,omitempty"`
}

func fetchPreview(url string) (*PreviewResponse, error) {
    reqBody := PreviewRequest{URL: url}
    jsonData, _ := json.Marshal(reqBody)
    
    resp, err := http.Post(
        "http://localhost:5465/preview",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    var preview PreviewResponse
    json.Unmarshal(body, &preview)
    
    return &preview, nil
}

func main() {
    preview, err := fetchPreview("https://golang.org")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Title: %s\n", preview.Title)
    fmt.Printf("Description: %s\n", preview.Description)
}
```

## Architecture Overview

### Goroutine Implementation

The API leverages Go's powerful concurrency model:

1. **Request Handler**: Each incoming request is handled by Gin's goroutine pool
2. **Preview Fetcher**: A separate goroutine is spawned for each URL fetch operation
3. **Channel Communication**: Results are passed back via buffered channels
4. **Context Cancellation**: Timeout contexts ensure requests don't hang indefinitely

```go
// Goroutine launch example from the code
go extractor.FetchLinkPreview(ctx, strings.TrimSpace(req.URL), resultChan)

// Channel-based result handling
select {
case result := <-resultChan:
    // Process successful result
case <-ctx.Done():
    // Handle timeout
}
```

### Key Components

- **MetaExtractor**: Core component responsible for fetching and parsing HTML
- **LinkPreviewRequest/Response**: Data structures for API communication
- **Context Management**: Timeout and cancellation handling
- **Regex Parsing**: Efficient metadata extraction from HTML content

## Configuration

### Environment Variables

- `GIN_MODE`: Set to `release` for production (default: `debug`)
- `PORT`: Server port (default: `5465`)

### Timeouts

- **HTTP Client Timeout**: 10 seconds
- **Request Context Timeout**: 15 seconds
- **Response Size Limit**: 1MB

## Error Handling

The API handles various error scenarios:

- Invalid URL format
- Network timeouts
- HTTP errors (4xx, 5xx)
- Malformed HTML
- Request timeouts
- Context cancellation

## Performance Considerations

- **Concurrent Processing**: Multiple preview requests are processed simultaneously
- **Memory Limits**: Response body reading is limited to 1MB to prevent memory issues
- **Timeout Management**: Prevents hanging requests with configurable timeouts
- **Efficient Parsing**: Regex-based HTML parsing for optimal performance

## Testing

### Test the API with sample URLs:

```bash
# Test with GitHub
curl -X POST http://localhost:5465/preview \
  -H "Content-Type: application/json" \
  -d '{"url": "https://github.com"}'

# Test with Stack Overflow
curl -X POST http://localhost:5465/preview \
  -H "Content-Type: application/json" \
  -d '{"url": "https://stackoverflow.com"}'

# Test with invalid URL
curl -X POST http://localhost:5465/preview \
  -H "Content-Type: application/json" \
  -d '{"url": "not-a-valid-url"}'
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Support

For issues and questions, please open an issue in the repository.

