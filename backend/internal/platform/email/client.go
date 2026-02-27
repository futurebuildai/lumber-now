package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/builderwire/lumber-now/backend/internal/domain"
)

type Client struct {
	apiKey     string
	from       string
	httpClient *http.Client
}

func NewClient(apiKey, from string) *Client {
	return &Client{
		apiKey:     apiKey,
		from:       from,
		httpClient: &http.Client{},
	}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (c *Client) Send(ctx context.Context, to, subject, htmlBody string) error {
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody bytes.Buffer
		errBody.ReadFrom(resp.Body)
		return fmt.Errorf("resend API error (status %d): %s", resp.StatusCode, errBody.String())
	}

	slog.Info("email sent via Resend", "to", to, "subject", subject)
	return nil
}

func (c *Client) SendOrderConfirmation(ctx context.Context, toEmail, dealerName string, items []domain.StructuredItem) error {
	subject := fmt.Sprintf("New Material Request from %s", dealerName)

	var rows strings.Builder
	for _, item := range items {
		rows.WriteString(fmt.Sprintf(
			"<tr><td style=\"padding:8px;border:1px solid #ddd;\">%s</td><td style=\"padding:8px;border:1px solid #ddd;\">%s</td><td style=\"padding:8px;border:1px solid #ddd;\">%.0f %s</td></tr>",
			item.SKU, item.Name, item.Quantity, item.Unit,
		))
	}

	html := fmt.Sprintf(`<html><body style="font-family:sans-serif;">
<h2>New Material Request</h2>
<p>A new order has been submitted via <strong>LumberNow</strong> for <strong>%s</strong>.</p>
<table style="border-collapse:collapse;width:100%%;">
<thead><tr style="background:#f5f5f5;">
<th style="padding:8px;border:1px solid #ddd;text-align:left;">SKU</th>
<th style="padding:8px;border:1px solid #ddd;text-align:left;">Item</th>
<th style="padding:8px;border:1px solid #ddd;text-align:left;">Quantity</th>
</tr></thead>
<tbody>%s</tbody>
</table>
<p style="margin-top:20px;color:#666;">This email was sent by LumberNow. Please log in to review and fulfill this request.</p>
</body></html>`, dealerName, rows.String())

	return c.Send(ctx, toEmail, subject, html)
}
