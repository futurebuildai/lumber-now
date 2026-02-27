package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

const baseURL = "https://api.anthropic.com/v1/messages"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type message struct {
	Role    string    `json:"role"`
	Content []content `json:"content"`
}

type content struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	Source    *source `json:"source,omitempty"`
}

type source struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

type apiRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []message `json:"messages"`
}

type apiResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func (c *Client) call(ctx context.Context, model, systemPrompt string, msgs []message) (string, error) {
	body := apiRequest{
		Model:     model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages:  msgs,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return apiResp.Content[0].Text, nil
}

func buildInventoryContext(inventory []db.Inventory) string {
	if len(inventory) == 0 {
		return "No inventory catalog available. Parse items as best as possible."
	}

	var buf bytes.Buffer
	buf.WriteString("Available inventory catalog:\n")
	for _, item := range inventory {
		buf.WriteString(fmt.Sprintf("- SKU: %s | Name: %s | Category: %s | Unit: %s | Price: %s\n",
			item.Sku, item.Name, item.Category, item.Unit, item.Price))
	}
	return buf.String()
}

func parseItemsFromResponse(text string) ([]domain.StructuredItem, float64, error) {
	// Try to extract JSON from the response
	start := -1
	end := -1
	for i, ch := range text {
		if ch == '[' && start == -1 {
			start = i
		}
		if ch == ']' {
			end = i + 1
		}
	}

	if start == -1 || end == -1 {
		return nil, 0, fmt.Errorf("no JSON array found in response")
	}

	var items []domain.StructuredItem
	if err := json.Unmarshal([]byte(text[start:end]), &items); err != nil {
		return nil, 0, fmt.Errorf("parse items JSON: %w", err)
	}

	// Calculate average confidence
	var totalConf float64
	for _, item := range items {
		totalConf += item.Confidence
	}
	avgConf := 0.0
	if len(items) > 0 {
		avgConf = totalConf / float64(len(items))
	}

	return items, avgConf, nil
}
