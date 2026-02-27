package anthropic

import (
	"context"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

const imageSystemPrompt = `You are an AI assistant for a lumber and building materials dealer. Your job is to analyze images of material lists, handwritten notes, product labels, or jobsite photos and extract structured material requests.

For each item identified, extract:
- sku: Match to the closest SKU from the inventory catalog if possible, otherwise leave empty
- name: The item name/description
- quantity: Numeric quantity (estimate if not clearly specified)
- unit: Unit of measure (EA, BF, LF, SF, etc.)
- confidence: Your confidence in the identification from 0.0 to 1.0
- matched: Whether you found a matching SKU in the catalog
- notes: Any relevant notes about what you see

Respond with ONLY a JSON array of items. No other text.`

func (c *Client) ParseImage(ctx context.Context, imageURL string, inventory []db.Inventory) ([]domain.StructuredItem, float64, error) {
	invContext := buildInventoryContext(inventory)

	msgs := []message{
		{
			Role: "user",
			Content: []content{
				{
					Type: "image",
					Source: &source{
						Type: "url",
						URL:  imageURL,
					},
				},
				{Type: "text", Text: invContext + "\n\nPlease analyze this image and extract all material items."},
			},
		},
	}

	resp, err := c.call(ctx, "claude-sonnet-4-5-20250929", imageSystemPrompt, msgs)
	if err != nil {
		return nil, 0, err
	}

	return parseItemsFromResponse(resp)
}
