package anthropic

import (
	"fmt"
	"strings"
	"testing"

	"github.com/builderwire/lumber-now/backend/internal/store/db"
	"github.com/google/uuid"
)

func TestParseItemsFromResponse(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantItemCount   int
		wantAvgConf     float64
		wantErr         bool
		checkFirstItem  bool
		expectedFirstSKU string
	}{
		{
			name: "valid JSON array",
			input: `[
				{"sku":"ABC123","name":"2x4 Lumber","quantity":10,"unit":"ea","confidence":0.95,"matched":true},
				{"sku":"XYZ789","name":"Nails","quantity":5,"unit":"box","confidence":0.85,"matched":true}
			]`,
			wantItemCount:    2,
			wantAvgConf:      0.90,
			wantErr:          false,
			checkFirstItem:   true,
			expectedFirstSKU: "ABC123",
		},
		{
			name: "JSON array with surrounding text",
			input: `Here are the items I found:
			[
				{"sku":"SKU001","name":"Plywood","quantity":3,"unit":"sheet","confidence":0.88,"matched":true},
				{"sku":"SKU002","name":"Screws","quantity":2,"unit":"box","confidence":0.92,"matched":false}
			]
			Total items: 2`,
			wantItemCount:    2,
			wantAvgConf:      0.90,
			wantErr:          false,
			checkFirstItem:   true,
			expectedFirstSKU: "SKU001",
		},
		{
			name:          "empty JSON array",
			input:         `[]`,
			wantItemCount: 0,
			wantAvgConf:   0.0,
			wantErr:       false,
		},
		{
			name:    "invalid JSON",
			input:   `[{"sku":"ABC123","name":"Item",,}]`,
			wantErr: true,
		},
		{
			name:    "no JSON array found",
			input:   `This is just text without any JSON`,
			wantErr: true,
		},
		{
			name: "array with mixed valid/invalid items",
			input: `[
				{"sku":"VALID1","name":"Item 1","quantity":1,"unit":"ea","confidence":0.75,"matched":true},
				{"sku":"VALID2","name":"Item 2","quantity":2,"unit":"ea","confidence":0.65,"matched":false}
			]`,
			wantItemCount:    2,
			wantAvgConf:      0.70,
			wantErr:          false,
			checkFirstItem:   true,
			expectedFirstSKU: "VALID1",
		},
		{
			name: "confidence calculation with multiple items",
			input: `[
				{"sku":"A","name":"Item A","quantity":1,"unit":"ea","confidence":1.0,"matched":true},
				{"sku":"B","name":"Item B","quantity":1,"unit":"ea","confidence":0.8,"matched":true},
				{"sku":"C","name":"Item C","quantity":1,"unit":"ea","confidence":0.6,"matched":true},
				{"sku":"D","name":"Item D","quantity":1,"unit":"ea","confidence":0.4,"matched":true}
			]`,
			wantItemCount: 4,
			wantAvgConf:   0.70,
			wantErr:       false,
		},
		{
			name: "single item confidence",
			input: `[
				{"sku":"SOLO","name":"Single Item","quantity":1,"unit":"ea","confidence":0.95,"matched":true}
			]`,
			wantItemCount:    1,
			wantAvgConf:      0.95,
			wantErr:          false,
			checkFirstItem:   true,
			expectedFirstSKU: "SOLO",
		},
		{
			name:    "incomplete JSON array (no closing bracket)",
			input:   `[{"sku":"ABC","name":"Item","quantity":1,"unit":"ea","confidence":0.9,"matched":true}`,
			wantErr: true,
		},
		{
			name:    "only opening bracket",
			input:   `[`,
			wantErr: true,
		},
		{
			name:    "only closing bracket",
			input:   `]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, avgConf, err := parseItemsFromResponse(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseItemsFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(items) != tt.wantItemCount {
				t.Errorf("parseItemsFromResponse() got %d items, want %d", len(items), tt.wantItemCount)
			}

			// Allow for small floating point differences
			if diff := avgConf - tt.wantAvgConf; diff < -0.01 || diff > 0.01 {
				t.Errorf("parseItemsFromResponse() avgConf = %v, want %v", avgConf, tt.wantAvgConf)
			}

			if tt.checkFirstItem && len(items) > 0 {
				if items[0].SKU != tt.expectedFirstSKU {
					t.Errorf("parseItemsFromResponse() first item SKU = %v, want %v", items[0].SKU, tt.expectedFirstSKU)
				}
			}
		})
	}
}

func TestBuildInventoryContext(t *testing.T) {
	tests := []struct {
		name         string
		inventory    []db.Inventory
		wantContains []string
		wantNotContains []string
	}{
		{
			name:      "empty inventory",
			inventory: []db.Inventory{},
			wantContains: []string{
				"No inventory catalog available",
				"Parse items as best as possible",
			},
		},
		{
			name: "single item",
			inventory: []db.Inventory{
				{
					ID:       uuid.New(),
					DealerID: uuid.New(),
					Sku:      "ABC123",
					Name:     "2x4 Lumber",
					Category: "Lumber",
					Unit:     "ea",
					Price:    "5.99",
				},
			},
			wantContains: []string{
				"Available inventory catalog:",
				"SKU: ABC123",
				"Name: 2x4 Lumber",
				"Category: Lumber",
				"Unit: ea",
				"Price: 5.99",
			},
		},
		{
			name: "multiple items",
			inventory: []db.Inventory{
				{
					Sku:      "SKU001",
					Name:     "Plywood",
					Category: "Wood",
					Unit:     "sheet",
					Price:    "45.00",
				},
				{
					Sku:      "SKU002",
					Name:     "Nails",
					Category: "Hardware",
					Unit:     "box",
					Price:    "12.50",
				},
				{
					Sku:      "SKU003",
					Name:     "Paint",
					Category: "Finishing",
					Unit:     "gal",
					Price:    "35.99",
				},
			},
			wantContains: []string{
				"Available inventory catalog:",
				"SKU: SKU001",
				"SKU: SKU002",
				"SKU: SKU003",
				"Name: Plywood",
				"Name: Nails",
				"Name: Paint",
			},
		},
		{
			name:      "more than 500 items should be capped",
			inventory: generateLargeInventory(600),
			wantContains: []string{
				"Available inventory catalog:",
				"Showing 500 of 600 items",
				"Match against shown items or parse as best as possible",
				"SKU: SKU-0", // First item
			},
			wantNotContains: []string{
				"SKU: SKU-550", // Items beyond 500 should not appear
			},
		},
		{
			name:      "exactly 500 items should not show capping message",
			inventory: generateLargeInventory(500),
			wantContains: []string{
				"Available inventory catalog:",
				"SKU: SKU-0",
				"SKU: SKU-499",
			},
			wantNotContains: []string{
				"Showing",
			},
		},
		{
			name:      "499 items should not show capping message",
			inventory: generateLargeInventory(499),
			wantContains: []string{
				"Available inventory catalog:",
			},
			wantNotContains: []string{
				"Showing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildInventoryContext(tt.inventory)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("buildInventoryContext() result does not contain %q", want)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(result, notWant) {
					t.Errorf("buildInventoryContext() result should not contain %q", notWant)
				}
			}

			// Additional check for large inventory: verify it starts with catalog header
			if len(tt.inventory) > 0 && !strings.HasPrefix(result, "No inventory") {
				if !strings.HasPrefix(result, "Available inventory catalog:") {
					t.Errorf("buildInventoryContext() should start with catalog header")
				}
			}
		})
	}
}

// generateLargeInventory creates n inventory items for testing
func generateLargeInventory(n int) []db.Inventory {
	items := make([]db.Inventory, n)
	for i := 0; i < n; i++ {
		items[i] = db.Inventory{
			ID:       uuid.New(),
			DealerID: uuid.New(),
			Sku:      fmt.Sprintf("SKU-%d", i),
			Name:     fmt.Sprintf("Item-%d", i),
			Category: "Category",
			Unit:     "ea",
			Price:    "10.00",
		}
	}
	return items
}

// TestParseItemsFromResponse_FieldValidation tests that all fields are properly parsed
func TestParseItemsFromResponse_FieldValidation(t *testing.T) {
	input := `[
		{
			"sku": "TEST-SKU-001",
			"name": "Test Product Name",
			"quantity": 42.5,
			"unit": "pieces",
			"confidence": 0.87,
			"matched": true,
			"notes": "Some important notes"
		}
	]`

	items, avgConf, err := parseItemsFromResponse(input)
	if err != nil {
		t.Fatalf("parseItemsFromResponse() unexpected error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]

	if item.SKU != "TEST-SKU-001" {
		t.Errorf("SKU = %v, want TEST-SKU-001", item.SKU)
	}
	if item.Name != "Test Product Name" {
		t.Errorf("Name = %v, want Test Product Name", item.Name)
	}
	if item.Quantity != 42.5 {
		t.Errorf("Quantity = %v, want 42.5", item.Quantity)
	}
	if item.Unit != "pieces" {
		t.Errorf("Unit = %v, want pieces", item.Unit)
	}
	if item.Confidence != 0.87 {
		t.Errorf("Confidence = %v, want 0.87", item.Confidence)
	}
	if !item.Matched {
		t.Errorf("Matched = %v, want true", item.Matched)
	}
	if item.Notes != "Some important notes" {
		t.Errorf("Notes = %v, want 'Some important notes'", item.Notes)
	}
	if avgConf != 0.87 {
		t.Errorf("avgConf = %v, want 0.87", avgConf)
	}
}

// TestBuildInventoryContext_Format tests the exact format of the output
func TestBuildInventoryContext_Format(t *testing.T) {
	inventory := []db.Inventory{
		{
			Sku:      "TEST-001",
			Name:     "Test Item",
			Category: "TestCat",
			Unit:     "box",
			Price:    "19.99",
		},
	}

	result := buildInventoryContext(inventory)

	expectedFormat := "- SKU: TEST-001 | Name: Test Item | Category: TestCat | Unit: box | Price: 19.99"
	if !strings.Contains(result, expectedFormat) {
		t.Errorf("buildInventoryContext() format mismatch.\nGot:\n%s\n\nExpected to contain:\n%s", result, expectedFormat)
	}
}
