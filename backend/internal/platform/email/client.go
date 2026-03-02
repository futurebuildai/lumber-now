package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/platform/circuitbreaker"
)

type Client struct {
	apiKey     string
	from       string
	httpClient *http.Client
	breaker    *circuitbreaker.Breaker
}

func NewClient(apiKey, from string) *Client {
	return &Client{
		apiKey: apiKey,
		from:   from,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		breaker: circuitbreaker.New(5, 60*time.Second),
	}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (c *Client) Send(ctx context.Context, to, subject, htmlBody string) error {
	return c.breaker.Execute(func() error {
		return c.doSend(ctx, to, subject, htmlBody)
	})
}

func (c *Client) doSend(ctx context.Context, to, subject, htmlBody string) error {
	payload := sendRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		HTML:    htmlBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal email payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create email request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	if rid := domain.RequestIDFromContext(ctx); rid != "" {
		req.Header.Set("X-Request-ID", rid)
	}

	maxRetries := 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			req.Body = io.NopCloser(bytes.NewReader(body))
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("send email: %w", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("resend API error (status %d)", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			return fmt.Errorf("resend API error (status %d)", resp.StatusCode)
		}

		slog.Info("email sent via Resend", "to", to, "subject", subject)
		return nil
	}

	return lastErr
}

var orderConfirmationTmpl = template.Must(template.New("order").Parse(`<html><body style="font-family:sans-serif;">
<h2>New Material Request</h2>
<p>A new order has been submitted via <strong>LumberNow</strong> for <strong>{{.DealerName}}</strong>.</p>
<table style="border-collapse:collapse;width:100%;">
<thead><tr style="background:#f5f5f5;">
<th style="padding:8px;border:1px solid #ddd;text-align:left;">SKU</th>
<th style="padding:8px;border:1px solid #ddd;text-align:left;">Item</th>
<th style="padding:8px;border:1px solid #ddd;text-align:left;">Quantity</th>
</tr></thead>
<tbody>
{{range .Items}}<tr><td style="padding:8px;border:1px solid #ddd;">{{.SKU}}</td><td style="padding:8px;border:1px solid #ddd;">{{.Name}}</td><td style="padding:8px;border:1px solid #ddd;">{{printf "%.0f" .Quantity}} {{.Unit}}</td></tr>
{{end}}</tbody>
</table>
<p style="margin-top:20px;color:#666;">This email was sent by LumberNow. Please log in to review and fulfill this request.</p>
</body></html>`))

func (c *Client) SendOrderConfirmation(ctx context.Context, toEmail, dealerName string, items []domain.StructuredItem) error {
	subject := fmt.Sprintf("New Material Request from %s", dealerName)

	var buf bytes.Buffer
	if err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: dealerName, Items: items}); err != nil {
		return fmt.Errorf("render email template: %w", err)
	}

	return c.Send(ctx, toEmail, subject, buf.String())
}
