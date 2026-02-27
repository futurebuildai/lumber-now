package email

import (
	"context"
	"log/slog"
)

type Client struct {
	from string
}

func NewClient(from string) *Client {
	return &Client{from: from}
}

func (c *Client) Send(ctx context.Context, to, subject, body string) error {
	// Placeholder: integrate with SES, SendGrid, etc.
	slog.Info("email sent",
		"from", c.from,
		"to", to,
		"subject", subject,
	)
	return nil
}
