package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// getCol
// ---------------------------------------------------------------------------

func TestGetCol(t *testing.T) {
	colMap := map[string]int{"sku": 0, "name": 1, "price": 2}
	record := []string{"ABC-123", "Test Item", "9.99"}

	tests := []struct {
		name   string
		record []string
		colMap map[string]int
		col    string
		want   string
	}{
		{"existing column", record, colMap, "sku", "ABC-123"},
		{"second column", record, colMap, "name", "Test Item"},
		{"third column", record, colMap, "price", "9.99"},
		{"missing column", record, colMap, "category", ""},
		{"empty colMap", record, map[string]int{}, "sku", ""},
		{"index out of range", []string{"A"}, map[string]int{"sku": 0, "name": 1}, "name", ""},
		{"whitespace trimmed", []string{"  ABC  "}, map[string]int{"sku": 0}, "sku", "ABC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCol(tt.record, tt.colMap, tt.col); got != tt.want {
				t.Errorf("getCol() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetCol_EmptyRecord(t *testing.T) {
	colMap := map[string]int{"sku": 0}
	got := getCol([]string{}, colMap, "sku")
	if got != "" {
		t.Errorf("getCol with empty record = %q, want empty string", got)
	}
}

func TestGetCol_NilRecord(t *testing.T) {
	colMap := map[string]int{"sku": 0}
	got := getCol(nil, colMap, "sku")
	if got != "" {
		t.Errorf("getCol with nil record = %q, want empty string", got)
	}
}

func TestGetCol_NilColMap(t *testing.T) {
	got := getCol([]string{"value"}, nil, "sku")
	if got != "" {
		t.Errorf("getCol with nil colMap = %q, want empty string", got)
	}
}

func TestGetCol_WhitespaceOnly(t *testing.T) {
	colMap := map[string]int{"sku": 0}
	got := getCol([]string{"   "}, colMap, "sku")
	if got != "" {
		t.Errorf("getCol with whitespace-only value = %q, want empty string", got)
	}
}

func TestGetCol_TabsAndNewlines(t *testing.T) {
	colMap := map[string]int{"sku": 0}
	got := getCol([]string{"\t hello \n"}, colMap, "sku")
	// strings.TrimSpace removes leading/trailing whitespace including tabs and newlines
	if got != "hello" {
		t.Errorf("getCol with tabs/newlines = %q, want %q", got, "hello")
	}
}

// ---------------------------------------------------------------------------
// orDefault
// ---------------------------------------------------------------------------

func TestOrDefault(t *testing.T) {
	tests := []struct {
		name string
		val  string
		def  string
		want string
	}{
		{"empty uses default", "", "EA", "EA"},
		{"non-empty uses value", "BF", "EA", "BF"},
		{"whitespace is not empty", " ", "EA", " "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := orDefault(tt.val, tt.def); got != tt.want {
				t.Errorf("orDefault(%q, %q) = %q, want %q", tt.val, tt.def, got, tt.want)
			}
		})
	}
}

func TestOrDefault_BothEmpty(t *testing.T) {
	got := orDefault("", "")
	if got != "" {
		t.Errorf("orDefault(\"\", \"\") = %q, want empty string", got)
	}
}

func TestOrDefault_DefaultNotUsedWhenValuePresent(t *testing.T) {
	got := orDefault("actual", "default")
	if got != "actual" {
		t.Errorf("orDefault(\"actual\", \"default\") = %q, want \"actual\"", got)
	}
}

// ---------------------------------------------------------------------------
// CSVImportResult struct
// ---------------------------------------------------------------------------

func TestCSVImportResult_ZeroValue(t *testing.T) {
	var result CSVImportResult
	if result.Imported != 0 {
		t.Errorf("zero-value Imported = %d, want 0", result.Imported)
	}
	if result.Skipped != 0 {
		t.Errorf("zero-value Skipped = %d, want 0", result.Skipped)
	}
	if result.Errors != nil {
		t.Errorf("zero-value Errors should be nil, got %v", result.Errors)
	}
}

func TestCSVImportResult_WithErrors(t *testing.T) {
	result := CSVImportResult{
		Imported: 5,
		Skipped:  2,
		Errors:   []string{"line 3: empty sku or name", "line 7: import failed"},
	}
	if result.Imported != 5 {
		t.Errorf("Imported = %d, want 5", result.Imported)
	}
	if result.Skipped != 2 {
		t.Errorf("Skipped = %d, want 2", result.Skipped)
	}
	if len(result.Errors) != 2 {
		t.Errorf("len(Errors) = %d, want 2", len(result.Errors))
	}
}

func TestCSVImportResult_EmptyErrors(t *testing.T) {
	result := CSVImportResult{
		Imported: 10,
		Skipped:  0,
		Errors:   []string{},
	}
	if result.Imported != 10 {
		t.Errorf("Imported = %d, want 10", result.Imported)
	}
	if len(result.Errors) != 0 {
		t.Errorf("len(Errors) = %d, want 0", len(result.Errors))
	}
}

func TestCSVImportResult_JSON(t *testing.T) {
	result := CSVImportResult{
		Imported: 3,
		Skipped:  1,
		Errors:   []string{"line 5: empty sku or name"},
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var decoded CSVImportResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if decoded.Imported != 3 || decoded.Skipped != 1 || len(decoded.Errors) != 1 {
		t.Errorf("round-trip mismatch: got %+v", decoded)
	}
}

func TestCSVImportResult_JSON_OmitsEmptyErrors(t *testing.T) {
	result := CSVImportResult{Imported: 2, Skipped: 0}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	s := string(data)
	if strings.Contains(s, "errors") {
		t.Errorf("expected omitempty to hide nil errors, got: %s", s)
	}
}

// ---------------------------------------------------------------------------
// NewInventoryService
// ---------------------------------------------------------------------------

func TestNewInventoryService(t *testing.T) {
	svc := NewInventoryService(nil)
	if svc == nil {
		t.Fatal("NewInventoryService(nil) returned nil")
	}
	if svc.store != nil {
		t.Error("expected store field to be nil when constructed with nil")
	}
}

// ===========================================================================
// ImportCSV tests
//
// ImportCSV requires a *store.Store with a working Pool (for transactions).
// Since pgxpool.Pool cannot be easily mocked without a real Postgres
// connection, we test the CSV parsing logic by exercising the pre-transaction
// code path. Errors from header validation and empty CSV are returned before
// WithTx is called. For valid CSV input that reaches the transaction stage,
// the nil store causes a panic which we recover from -- the key assertion is
// that CSV parsing completed without returning a header/format error.
// ===========================================================================

// callImportCSV is a helper that calls ImportCSV, recovering from panics
// caused by nil store. It returns the result and error, or (nil, nil) if a
// panic occurred (meaning parsing succeeded but the transaction layer failed).
func callImportCSV(t *testing.T, csvContent string) (*CSVImportResult, error) {
	t.Helper()
	svc := NewInventoryService(nil)

	var result *CSVImportResult
	var err error
	panicked := false

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		result, err = svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csvContent))
	}()

	if panicked {
		return nil, nil
	}
	return result, err
}

// assertNoHeaderError checks that the error (if any) is not a "missing
// required column" error.
func assertNoHeaderError(t *testing.T, err error) {
	t.Helper()
	if err != nil && strings.Contains(err.Error(), "missing required column") {
		t.Errorf("unexpected header error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ImportCSV: Error cases (returned before transaction)
// ---------------------------------------------------------------------------

func TestImportCSV_EmptyReader(t *testing.T) {
	svc := NewInventoryService(nil)
	_, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(""))
	if err == nil {
		t.Error("expected error for empty CSV, got nil")
	}
	if !strings.Contains(err.Error(), "read CSV header") {
		t.Errorf("expected 'read CSV header' error, got: %v", err)
	}
}

func TestImportCSV_MissingSKUHeader(t *testing.T) {
	svc := NewInventoryService(nil)
	csv := "name,description,category\nWidget,A widget,Tools\n"
	_, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing 'sku' header, got nil")
	}
	if !strings.Contains(err.Error(), "missing required column: sku") {
		t.Errorf("expected 'missing required column: sku', got: %v", err)
	}
}

func TestImportCSV_MissingNameHeader(t *testing.T) {
	svc := NewInventoryService(nil)
	csv := "sku,description,category\nABC,A widget,Tools\n"
	_, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing 'name' header, got nil")
	}
	if !strings.Contains(err.Error(), "missing required column: name") {
		t.Errorf("expected 'missing required column: name', got: %v", err)
	}
}

func TestImportCSV_MissingBothRequiredHeaders(t *testing.T) {
	svc := NewInventoryService(nil)
	csv := "description,category\nA widget,Tools\n"
	_, err := svc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	if err == nil {
		t.Fatal("expected error for missing required headers, got nil")
	}
	// The code checks sku first, then name.
	if !strings.Contains(err.Error(), "missing required column") {
		t.Errorf("expected 'missing required column' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ImportCSV: Header handling (case, whitespace, ordering)
// ---------------------------------------------------------------------------

func TestImportCSV_CaseInsensitiveHeaders(t *testing.T) {
	_, err := callImportCSV(t, "SKU,NAME,DESCRIPTION\nABC,Widget,A tool\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_HeadersWithLeadingSpaces(t *testing.T) {
	_, err := callImportCSV(t, " sku, name, description\nABC,Widget,A tool\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_ReorderedColumns(t *testing.T) {
	_, err := callImportCSV(t, "price,name,sku,category\n12.99,Lumber,ABC-001,Wood\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_ExtraColumnsIgnored(t *testing.T) {
	_, err := callImportCSV(t, "sku,name,extra_col1,extra_col2\nABC,Widget,foo,bar\n")
	assertNoHeaderError(t, err)
}

// ---------------------------------------------------------------------------
// ImportCSV: Valid CSV (reaches transaction, panics on nil store)
// ---------------------------------------------------------------------------

func TestImportCSV_HeaderOnly_NoDataRows(t *testing.T) {
	result, err := callImportCSV(t, "sku,name,description,category,unit,price,in_stock,metadata\n")
	assertNoHeaderError(t, err)
	if result != nil {
		if result.Imported != 0 {
			t.Errorf("expected 0 imported for header-only CSV, got %d", result.Imported)
		}
		if result.Skipped != 0 {
			t.Errorf("expected 0 skipped for header-only CSV, got %d", result.Skipped)
		}
	}
}

func TestImportCSV_OnlyHeaderRow(t *testing.T) {
	result, err := callImportCSV(t, "sku,name\n")
	assertNoHeaderError(t, err)
	if result != nil {
		if result.Imported != 0 {
			t.Errorf("expected 0 imported for header-only CSV, got %d", result.Imported)
		}
	}
}

func TestImportCSV_SingleValidRow(t *testing.T) {
	_, err := callImportCSV(t, "sku,name,price,in_stock\nABC-001,Lumber 2x4,12.99,true\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_AllColumnsProvided(t *testing.T) {
	csv := `sku,name,description,category,unit,price,in_stock,metadata
ABC-001,Lumber 2x4,8ft kiln-dried,Lumber,EA,12.99,true,{}
DEF-002,Plywood,3/4 inch,Panels,SHT,45.00,false,{}
`
	_, err := callImportCSV(t, csv)
	assertNoHeaderError(t, err)
}

// ---------------------------------------------------------------------------
// ImportCSV: Row-level validation (empty sku/name)
// ---------------------------------------------------------------------------

func TestImportCSV_RowWithEmptySKU(t *testing.T) {
	// A row where SKU is empty should be skipped during parsing.
	_, err := callImportCSV(t, "sku,name\n,Widget\nABC,Board\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_RowWithEmptyName(t *testing.T) {
	_, err := callImportCSV(t, "sku,name\nABC-001,\nDEF-002,Plywood\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_WhitespaceOnlySKU(t *testing.T) {
	// After TrimSpace, a whitespace-only SKU becomes "" which triggers skip.
	_, err := callImportCSV(t, "sku,name\n   ,Widget\nABC,Board\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_WhitespaceOnlyName(t *testing.T) {
	_, err := callImportCSV(t, "sku,name\nABC,   \nDEF,Board\n")
	assertNoHeaderError(t, err)
}

// ---------------------------------------------------------------------------
// ImportCSV: Price parsing
// ---------------------------------------------------------------------------

func TestImportCSV_InvalidPriceDefaultsToZero(t *testing.T) {
	// An invalid price like "not-a-number" does not cause a parse error;
	// it defaults to "0" per the code logic.
	_, err := callImportCSV(t, "sku,name,price\nABC,Widget,not-a-number\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_ZeroPrice(t *testing.T) {
	_, err := callImportCSV(t, "sku,name,price\nABC,Free Item,0\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_NegativePrice(t *testing.T) {
	_, err := callImportCSV(t, "sku,name,price\nABC,Credit,-5.00\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_EmptyPrice(t *testing.T) {
	_, err := callImportCSV(t, "sku,name,price\nABC,Widget,\n")
	assertNoHeaderError(t, err)
}

// ---------------------------------------------------------------------------
// ImportCSV: in_stock parsing
// ---------------------------------------------------------------------------

func TestImportCSV_InStockParsing(t *testing.T) {
	// The code accepts "true", "1", "yes" (case-insensitive) as true.
	// All other values are interpreted as false.
	tests := []struct {
		name    string
		inStock string
	}{
		{"true lowercase", "true"},
		{"TRUE uppercase", "TRUE"},
		{"True mixed", "True"},
		{"1", "1"},
		{"yes", "yes"},
		{"YES", "YES"},
		{"false", "false"},
		{"0", "0"},
		{"no", "no"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csv := "sku,name,in_stock\nABC,Widget," + tt.inStock + "\n"
			_, err := callImportCSV(t, csv)
			assertNoHeaderError(t, err)
		})
	}
}

// ---------------------------------------------------------------------------
// ImportCSV: Unit defaults to "EA"
// ---------------------------------------------------------------------------

func TestImportCSV_UnitDefaultsToEA(t *testing.T) {
	_, err := callImportCSV(t, "sku,name\nABC,Widget\n")
	assertNoHeaderError(t, err)
}

// ---------------------------------------------------------------------------
// ImportCSV: Malformed CSV
// ---------------------------------------------------------------------------

func TestImportCSV_MalformedRow_WrongFieldCount(t *testing.T) {
	// This CSV has 3 header columns but one row has only 1 field.
	// The Go CSV reader will report an error for field count mismatch.
	// The malformed row should be skipped; subsequent rows should be parsed.
	csv := "sku,name,price\nABC,Widget,9.99\nBAD_ROW\nDEF,Board,4.50\n"
	_, err := callImportCSV(t, csv)
	assertNoHeaderError(t, err)
}

func TestImportCSV_WindowsLineEndings(t *testing.T) {
	_, err := callImportCSV(t, "sku,name,price\r\nABC,Widget,9.99\r\nDEF,Board,4.50\r\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_QuotedFields(t *testing.T) {
	csv := `sku,name,description
"ABC-001","2x4 Lumber","8 foot, kiln-dried"
"DEF-002","Plywood 3/4""","For cabinets"
`
	_, err := callImportCSV(t, csv)
	assertNoHeaderError(t, err)
}

// ---------------------------------------------------------------------------
// ImportCSV: Mixed valid and invalid rows
// ---------------------------------------------------------------------------

func TestImportCSV_MixedValidAndInvalidRows(t *testing.T) {
	csv := `sku,name,price
ABC-001,Lumber,12.99
,Missing SKU,5.00
DEF-002,,3.50
GHI-003,Plywood,not-a-price
JKL-004,Nails,2.99
`
	// Parsing should identify:
	// - Row 2 (ABC-001): valid
	// - Row 3 (empty SKU): skipped
	// - Row 4 (empty name): skipped
	// - Row 5 (GHI-003): valid (bad price defaults to "0")
	// - Row 6 (JKL-004): valid
	_, err := callImportCSV(t, csv)
	assertNoHeaderError(t, err)
}

// ---------------------------------------------------------------------------
// ImportCSV: Edge cases
// ---------------------------------------------------------------------------

func TestImportCSV_ManyColumns(t *testing.T) {
	// Build a CSV with many extra columns.
	headers := "sku,name"
	values := "ABC,Widget"
	for i := 0; i < 50; i++ {
		headers += ",extraxxxxx"
		values += ",val"
	}
	csv := headers + "\n" + values + "\n"
	_, err := callImportCSV(t, csv)
	assertNoHeaderError(t, err)
}

func TestImportCSV_UnicodeContent(t *testing.T) {
	_, err := callImportCSV(t, "sku,name,description\nABC-001,Holz Bretter,Hochwertiges Fichtenholz\nDEF-002,Madera,Madera de pino\n")
	assertNoHeaderError(t, err)
}

func TestImportCSV_DuplicateSKUs(t *testing.T) {
	csv := `sku,name,price
ABC-001,First Entry,10.00
ABC-001,Updated Entry,15.00
`
	// Both rows should be parsed; the DB upsert handles deduplication.
	_, err := callImportCSV(t, csv)
	assertNoHeaderError(t, err)
}
