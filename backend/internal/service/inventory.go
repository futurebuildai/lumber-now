package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/builderwire/lumber-now/backend/internal/store"
	"github.com/builderwire/lumber-now/backend/internal/store/db"
)

type InventoryService struct {
	store *store.Store
}

func NewInventoryService(s *store.Store) *InventoryService {
	return &InventoryService{store: s}
}

type CSVImportResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

func (s *InventoryService) ImportCSV(ctx context.Context, dealerID uuid.UUID, reader io.Reader) (*CSVImportResult, error) {
	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true

	// Read header
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	requiredCols := []string{"sku", "name"}
	for _, col := range requiredCols {
		if _, ok := colMap[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	// Parse all records first, then import in a transaction
	type parsedRow struct {
		lineNum int
		params  db.UpsertInventoryItemParams
	}

	var rows []parsedRow
	result := &CSVImportResult{}
	lineNum := 1

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		lineNum++
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: %v", lineNum, err))
			result.Skipped++
			continue
		}

		sku := getCol(record, colMap, "sku")
		name := getCol(record, colMap, "name")
		if sku == "" || name == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: empty sku or name", lineNum))
			result.Skipped++
			continue
		}

		price := "0"
		if p := getCol(record, colMap, "price"); p != "" {
			if _, err := strconv.ParseFloat(p, 64); err == nil {
				price = p
			}
		}

		inStock := true
		if sv := getCol(record, colMap, "in_stock"); sv != "" {
			inStock = strings.ToLower(sv) == "true" || sv == "1" || strings.ToLower(sv) == "yes"
		}

		rows = append(rows, parsedRow{
			lineNum: lineNum,
			params: db.UpsertInventoryItemParams{
				DealerID:    dealerID,
				Sku:         sku,
				Name:        name,
				Description: getCol(record, colMap, "description"),
				Category:    getCol(record, colMap, "category"),
				Unit:        orDefault(getCol(record, colMap, "unit"), "EA"),
				Price:       price,
				InStock:     inStock,
				Metadata:    json.RawMessage("{}"),
			},
		})
	}

	// Import all rows within a single transaction
	if err := s.store.WithTx(ctx, func(qtx *db.Queries) error {
		for _, row := range rows {
			if _, err := qtx.UpsertInventoryItem(ctx, row.params); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("line %d: import failed", row.lineNum))
				result.Skipped++
				continue
			}
			result.Imported++
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("import transaction: %w", err)
	}

	return result, nil
}

func getCol(record []string, colMap map[string]int, col string) string {
	idx, ok := colMap[col]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
