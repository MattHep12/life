package payroll

import "testing"

func TestParseTextKeepsOnlyApprovedPayrollFields(t *testing.T) {
	text := `
Pay Statement
Period Start Date 07/27/2026
Period End Date 08/09/2026
Pay Date 08/14/2026
Document 12345678
Sensitive Person Employee Number 999999999 SSN XXX-XX-1234
Pay Summary
Gross FIT Taxable Wages Taxes Deductions Net Pay
Current $4,777.32 $4,771.42 $1,268.88 $12.05 $3,496.39
YTD $111,944.54 $86,344.24 $25,800.86 $39,167.63 $46,976.05
`

	statement, err := ParseText(text)
	if err != nil {
		t.Fatal(err)
	}

	if statement.DocumentNumber != "12345678" || statement.GrossCents != 477732 || statement.NetCents != 349639 {
		t.Fatalf("unexpected parsed statement: %#v", statement)
	}
	if statement.SourceFilename != "" {
		t.Fatalf("ParseText should not invent a source filename: %q", statement.SourceFilename)
	}
}

func TestParseTextRejectsUnexpectedLayout(t *testing.T) {
	if _, err := ParseText("not a payroll statement"); err == nil {
		t.Fatal("expected unsupported layout error")
	}
}
