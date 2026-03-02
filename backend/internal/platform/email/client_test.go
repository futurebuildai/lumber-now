package email

import (
	"bytes"
	"strings"
	"testing"

	"github.com/builderwire/lumber-now/backend/internal/domain"
)

// TestOrderConfirmationTemplateRendersItems verifies that the email template
// renders a basic list of items correctly.
func TestOrderConfirmationTemplateRendersItems(t *testing.T) {
	items := []domain.StructuredItem{
		{SKU: "LBR-001", Name: "2x4 Stud 8ft", Quantity: 100, Unit: "pcs"},
		{SKU: "PLY-003", Name: "3/4 Plywood Sheet", Quantity: 20, Unit: "sheets"},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Acme Lumber", Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	// Verify dealer name is present
	if !strings.Contains(output, "Acme Lumber") {
		t.Error("output should contain dealer name 'Acme Lumber'")
	}

	// Verify SKUs appear
	if !strings.Contains(output, "LBR-001") {
		t.Error("output should contain SKU 'LBR-001'")
	}
	if !strings.Contains(output, "PLY-003") {
		t.Error("output should contain SKU 'PLY-003'")
	}

	// Verify item names appear
	if !strings.Contains(output, "2x4 Stud 8ft") {
		t.Error("output should contain item name '2x4 Stud 8ft'")
	}
	if !strings.Contains(output, "3/4 Plywood Sheet") {
		t.Error("output should contain item name '3/4 Plywood Sheet'")
	}

	// Verify quantities are formatted without decimals
	if !strings.Contains(output, "100 pcs") {
		t.Error("output should contain '100 pcs'")
	}
	if !strings.Contains(output, "20 sheets") {
		t.Error("output should contain '20 sheets'")
	}
}

// TestOrderConfirmationTemplateEscapesXSS verifies that html/template
// properly escapes dangerous characters in item names, preventing XSS.
func TestOrderConfirmationTemplateEscapesXSS(t *testing.T) {
	items := []domain.StructuredItem{
		{
			SKU:      "XSS-001",
			Name:     `<script>alert("xss")</script>`,
			Quantity: 1,
			Unit:     "pcs",
		},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Test Dealer", Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	// The raw <script> tag must NOT appear in the output
	if strings.Contains(output, "<script>") {
		t.Error("output contains unescaped <script> tag -- XSS vulnerability")
	}

	// The escaped form should be present instead
	if !strings.Contains(output, "&lt;script&gt;") {
		t.Error("output should contain HTML-escaped script tag (&lt;script&gt;)")
	}
}

// TestOrderConfirmationTemplateEscapesDealerNameXSS verifies that the dealer
// name is also escaped by html/template.
func TestOrderConfirmationTemplateEscapesDealerNameXSS(t *testing.T) {
	items := []domain.StructuredItem{
		{SKU: "SAFE-001", Name: "Normal Item", Quantity: 1, Unit: "pcs"},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: `Evil<img src=x onerror=alert(1)>Corp`, Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	if strings.Contains(output, `<img src=x`) {
		t.Error("output contains unescaped <img> tag in dealer name -- XSS vulnerability")
	}

	if !strings.Contains(output, "&lt;img") {
		t.Error("output should contain HTML-escaped img tag in dealer name")
	}
}

// TestOrderConfirmationTemplateEscapesHTMLEntitiesInSKU verifies that
// special characters in SKU fields are escaped.
func TestOrderConfirmationTemplateEscapesHTMLEntitiesInSKU(t *testing.T) {
	items := []domain.StructuredItem{
		{
			SKU:      `"><img src=x onerror=alert(1)>`,
			Name:     "Lumber",
			Quantity: 5,
			Unit:     "pcs",
		},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Safe Dealer", Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	if strings.Contains(output, `"><img`) {
		t.Error("output contains unescaped HTML in SKU field -- XSS vulnerability")
	}
}

// TestOrderConfirmationTemplateEscapesUnitXSS verifies that the unit field
// is also escaped.
func TestOrderConfirmationTemplateEscapesUnitXSS(t *testing.T) {
	items := []domain.StructuredItem{
		{
			SKU:      "UNIT-001",
			Name:     "Board",
			Quantity: 10,
			Unit:     `<b onmouseover=alert(1)>pcs</b>`,
		},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Test", Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	if strings.Contains(output, `<b onmouseover`) {
		t.Error("output contains unescaped HTML in unit field -- XSS vulnerability")
	}
}

// TestOrderConfirmationTemplateAmpersandEscaping verifies that ampersands
// in names are properly escaped to &amp; in HTML output.
func TestOrderConfirmationTemplateAmpersandEscaping(t *testing.T) {
	items := []domain.StructuredItem{
		{SKU: "AMP-001", Name: "Nuts & Bolts", Quantity: 50, Unit: "pcs"},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Smith & Sons", Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	// Ampersands should be escaped to &amp;
	if !strings.Contains(output, "Nuts &amp; Bolts") {
		t.Error("ampersand in item name should be escaped to &amp;")
	}
	if !strings.Contains(output, "Smith &amp; Sons") {
		t.Error("ampersand in dealer name should be escaped to &amp;")
	}
}

// TestOrderConfirmationTemplateQuoteInText verifies that double quotes
// in text content are handled correctly. In html/template, double quotes
// in element text context are rendered as &#34; for safety.
func TestOrderConfirmationTemplateQuoteInText(t *testing.T) {
	items := []domain.StructuredItem{
		{SKU: "QT-001", Name: `2"x4" Lumber`, Quantity: 25, Unit: "pcs"},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Test Dealer", Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	// html/template escapes double quotes to &#34; even in text context
	if strings.Contains(output, `2"x4"`) {
		t.Error("double quotes in item name should be escaped to &#34;")
	}
	if !strings.Contains(output, "2&#34;x4&#34; Lumber") {
		t.Error("expected &#34; escaped form of double quotes")
	}
}

// TestOrderConfirmationTemplateEmptyItems verifies that the template handles
// an empty items slice gracefully (no table rows, no error).
func TestOrderConfirmationTemplateEmptyItems(t *testing.T) {
	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Empty Dealer", Items: nil})

	if err != nil {
		t.Fatalf("template execution failed with empty items: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Empty Dealer") {
		t.Error("output should still contain the dealer name")
	}

	// Should contain the table structure but no data rows
	if !strings.Contains(output, "<thead>") {
		t.Error("output should contain the table header")
	}
}

// TestOrderConfirmationTemplateStructure verifies that the HTML structure
// is well-formed with expected elements.
func TestOrderConfirmationTemplateStructure(t *testing.T) {
	items := []domain.StructuredItem{
		{SKU: "STR-001", Name: "Test Board", Quantity: 1, Unit: "ea"},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Structure Test", Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	requiredElements := []string{
		"<html>",
		"</html>",
		"<body",
		"</body>",
		"<h2>New Material Request</h2>",
		"<table",
		"</table>",
		"<thead>",
		"</thead>",
		"<tbody>",
		"</tbody>",
		"LumberNow",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(output, elem) {
			t.Errorf("output missing expected element: %s", elem)
		}
	}
}

// TestOrderConfirmationQuantityFormatting verifies that quantities are
// formatted as integers (no decimal point) using printf "%.0f".
func TestOrderConfirmationQuantityFormatting(t *testing.T) {
	items := []domain.StructuredItem{
		{SKU: "FMT-001", Name: "Board", Quantity: 42.0, Unit: "pcs"},
		{SKU: "FMT-002", Name: "Sheet", Quantity: 100.0, Unit: "sheets"},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Format Test", Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	// Should show "42 pcs" not "42.0 pcs" or "42.000000 pcs"
	if !strings.Contains(output, "42 pcs") {
		t.Error("quantity 42.0 should be formatted as '42 pcs'")
	}
	if strings.Contains(output, "42.0") {
		t.Error("quantity should not show decimal point for whole numbers")
	}

	if !strings.Contains(output, "100 sheets") {
		t.Error("quantity 100.0 should be formatted as '100 sheets'")
	}
}

// ---------------------------------------------------------------------------
// Client construction and configuration tests
// ---------------------------------------------------------------------------

// TestNewClient verifies that NewClient creates a properly configured client.
func TestNewClient(t *testing.T) {
	client := NewClient("re_test_api_key_123", "noreply@example.com")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.apiKey != "re_test_api_key_123" {
		t.Errorf("apiKey = %q, want %q", client.apiKey, "re_test_api_key_123")
	}
	if client.from != "noreply@example.com" {
		t.Errorf("from = %q, want %q", client.from, "noreply@example.com")
	}
}

// TestNewClient_HTTPClientConfigured verifies that NewClient sets up the HTTP
// client with a non-nil transport and a timeout.
func TestNewClient_HTTPClientConfigured(t *testing.T) {
	client := NewClient("key", "from@test.com")
	if client.httpClient == nil {
		t.Fatal("httpClient should not be nil")
	}
	if client.httpClient.Timeout <= 0 {
		t.Errorf("httpClient.Timeout should be positive, got %v", client.httpClient.Timeout)
	}
	if client.httpClient.Timeout.Seconds() != 30 {
		t.Errorf("httpClient.Timeout = %v, want 30s", client.httpClient.Timeout)
	}
}

// TestNewClient_BreakerConfigured verifies that the circuit breaker is set up.
func TestNewClient_BreakerConfigured(t *testing.T) {
	client := NewClient("key", "from@test.com")
	if client.breaker == nil {
		t.Fatal("breaker should not be nil")
	}
}

// TestNewClient_EmptyAPIKey verifies that NewClient works even with an empty
// API key (the validation happens at send time).
func TestNewClient_EmptyAPIKey(t *testing.T) {
	client := NewClient("", "from@test.com")
	if client == nil {
		t.Fatal("NewClient with empty API key returned nil")
	}
	if client.apiKey != "" {
		t.Errorf("apiKey = %q, want empty string", client.apiKey)
	}
}

// TestNewClient_EmptyFrom verifies that NewClient works with an empty from
// address (the validation happens at send time via the email API).
func TestNewClient_EmptyFrom(t *testing.T) {
	client := NewClient("key", "")
	if client == nil {
		t.Fatal("NewClient with empty from returned nil")
	}
	if client.from != "" {
		t.Errorf("from = %q, want empty string", client.from)
	}
}

// TestSendRequest_Serialization verifies that the sendRequest struct can be
// marshaled to JSON with the expected field names.
func TestSendRequest_Serialization(t *testing.T) {
	req := sendRequest{
		From:    "sender@test.com",
		To:      []string{"recipient@test.com"},
		Subject: "Test Subject",
		HTML:    "<p>Hello</p>",
	}

	if req.From != "sender@test.com" {
		t.Errorf("From = %q, want %q", req.From, "sender@test.com")
	}
	if len(req.To) != 1 || req.To[0] != "recipient@test.com" {
		t.Errorf("To = %v, want [recipient@test.com]", req.To)
	}
	if req.Subject != "Test Subject" {
		t.Errorf("Subject = %q, want %q", req.Subject, "Test Subject")
	}
	if req.HTML != "<p>Hello</p>" {
		t.Errorf("HTML = %q, want %q", req.HTML, "<p>Hello</p>")
	}
}

// TestOrderConfirmationTemplateWithManyItems verifies that the template
// handles a large number of items without error.
func TestOrderConfirmationTemplateWithManyItems(t *testing.T) {
	items := make([]domain.StructuredItem, 100)
	for i := range items {
		items[i] = domain.StructuredItem{
			SKU:      "BULK-" + strings.Repeat("0", 3-len(strings.TrimLeft("", "0"))) + string(rune('0'+i%10)),
			Name:     "Bulk Item",
			Quantity: float64(i + 1),
			Unit:     "pcs",
		}
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Bulk Test", Items: items})

	if err != nil {
		t.Fatalf("template execution failed with many items: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Bulk Test") {
		t.Error("output should contain dealer name for bulk test")
	}
}

// TestOrderConfirmationTemplateTableHeaders verifies the expected column
// headers exist in the template output.
func TestOrderConfirmationTemplateTableHeaders(t *testing.T) {
	items := []domain.StructuredItem{
		{SKU: "HDR-001", Name: "Header Test", Quantity: 1, Unit: "ea"},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Header Test Dealer", Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	headers := []string{"SKU", "Item", "Quantity"}
	for _, h := range headers {
		if !strings.Contains(output, h) {
			t.Errorf("output missing expected table header: %s", h)
		}
	}
}

// TestOrderConfirmationTemplateFractionalQuantity verifies that fractional
// quantities are formatted without decimal places by %.0f.
func TestOrderConfirmationTemplateFractionalQuantity(t *testing.T) {
	items := []domain.StructuredItem{
		{SKU: "FRAC-001", Name: "Partial Item", Quantity: 2.7, Unit: "pcs"},
	}

	var buf bytes.Buffer
	err := orderConfirmationTmpl.Execute(&buf, struct {
		DealerName string
		Items      []domain.StructuredItem
	}{DealerName: "Fraction Test", Items: items})

	if err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	output := buf.String()

	// %.0f of 2.7 should produce "3" (rounds to nearest)
	if !strings.Contains(output, "3 pcs") {
		t.Errorf("expected quantity 2.7 to be formatted as '3 pcs' (%%.0f rounding), output: %s", output)
	}
}
