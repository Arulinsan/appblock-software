package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	geminiAPIEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"
	defaultTimeout    = 8 * time.Second
)

// Client handles Gemini API interactions
type Client struct {
	apiKey      string
	model       string
	personality string
	httpClient  *http.Client
	lastMessage string
	mu          sync.RWMutex
}

// GeminiRequest represents the request structure
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

// GeminiResponse represents the response structure
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

// NewClient creates a new Gemini API client
func NewClient(model, personality string) (*Client, error) {
	apiKey := getAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not found. Set environment variable or create .env file")
	}

	return &Client{
		apiKey:      apiKey,
		model:       model,
		personality: personality,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		lastMessage: "",
	}, nil
}

// getAPIKey tries to get API key from multiple sources
func getAPIKey() string {
	// 1. Try environment variable first
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey != "" {
		return apiKey
	}

	// 2. Try .env file in executable directory
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	exeDir := filepath.Dir(exePath)
	envPath := filepath.Join(exeDir, ".env")

	data, err := os.ReadFile(envPath)
	if err != nil {
		return ""
	}

	// Parse .env file (simple format: GEMINI_API_KEY=your-key-here)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "GEMINI_API_KEY=") {
			key := strings.TrimPrefix(line, "GEMINI_API_KEY=")
			key = strings.Trim(key, "\"'") // Remove quotes if present
			return strings.TrimSpace(key)
		}
	}

	return ""
}

// GetMotivationalMessage gets a motivational message from Gemini AI
func (c *Client) GetMotivationalMessage(blockedApp string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Try to get message from API
	message, err := c.fetchMessage(blockedApp)
	if err != nil {
		// Return last successful message or default
		return c.lastMessage
	}

	// Update last message cache
	c.lastMessage = message
	return message
}

// fetchMessage fetches a new message from Gemini API
func (c *Client) fetchMessage(blockedApp string) (string, error) {
	currentTime := time.Now().Format("15:04")
	
	prompt := fmt.Sprintf(`Kamu adalah asisten produktivitas yang %s.

Aplikasi "%s" baru saja ditutup pada jam %s karena sedang waktu produktif untuk belajar.

Berikan pesan motivasi singkat (maksimal 2-3 kalimat) yang:
1. Mengingatkan pentingnya fokus belajar
2. Memberikan saran konkret yang bisa dilakukan sekarang
3. Dalam bahasa Indonesia

Langsung berikan pesannya tanpa tambahan format atau penjelasan lain.

konteks lainnya pengguna adalah seorang programmer yang ingin belajar lebih giat tentang ngoding sebagai software developer
mau itu mobile dev, web dev, dan backend`, 
		c.personality, blockedApp, currentTime)

	reqBody := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf(geminiAPIEndpoint, c.model, c.apiKey)
	
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// GetLastMessage returns the last successful message
func (c *Client) GetLastMessage() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastMessage
}
