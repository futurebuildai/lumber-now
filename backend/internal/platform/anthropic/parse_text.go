package anthropic

import (
	"context"

	"github.com/builderwire/lumber-now/backend/internal/domain"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

const textSystemPrompt = `You are an AI assistant for a lumber and building materials dealer. Your job is to parse unstructured material requests from contractors into structured line items.

For each item mentioned, extract:
- sku: Match to the closest SKU from the inventory catalog if possible, otherwise leave empty
- name: The item name/description
- quantity: Numeric quantity
- unit: Unit of measure (EA, BF, LF, SF, etc.)
- confidence: Your confidence in the match from 0.0 to 1.0
- matched: Whether you found a matching SKU in the catalog
- notes: Any relevant notes about ambiguity or assumptions

Respond with ONLY a JSON array of items. No other text.

Example output:
[{"sku":"2X4-08-SPF","name":"2x4x8 SPF Stud","quantity":100,"unit":"EA","confidence":0.95,"matched":true,"notes":""}]`

func (c *Client) ParseText(ctx context.Context, text string, inventory []db.Inventory) ([]domain.StructuredItem, float64, error) {
	invContext := buildInventoryContext(inventory)

	msgs := []message{
		{
			Role: "user",
			Content: []content{
				{Type: "text", Text: invContext + "\n\nContractor request:\n" + text},
			},
		},
	}

	resp, err := c.call(ctx, "claude-haiku-4-5-20251001", textSystemPrompt, msgs)
	if err != nil {
		return nil, 0, err
	}

	return parseItemsFromResponse(resp)
}
