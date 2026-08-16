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
Bonus 401K Pre $1,000.00 Yes $250.00 $4,000.00 $125.00 $2,000.00
401K Pretax $3,000.00 Yes $450.00 $8,000.00 $225.00 $4,000.00
`

	statement, err := ParseText(text)
	if err != nil {
		t.Fatal(err)
	}

	if statement.DocumentNumber != "12345678" || statement.GrossCents != 477732 || statement.NetCents != 349639 {
		t.Fatalf("unexpected parsed statement: %#v", statement)
	}
	if statement.Employee401KCents != 70000 || statement.Employer401KCents != 35000 {
		t.Fatalf("unexpected 401(k) amounts: employee=%d employer=%d", statement.Employee401KCents, statement.Employer401KCents)
	}
	if statement.Employee401KYTDCents != 1200000 || statement.Employer401KYTDCents != 600000 {
		t.Fatalf("unexpected 401(k) YTD amounts: employee=%d employer=%d", statement.Employee401KYTDCents, statement.Employer401KYTDCents)
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

func TestParseTextSupportsCheckHistoryRetirementLayout(t *testing.T) {
	text := `
Period Start Date 07/27/2026
Period End Date 08/09/2026
Pay Date 08/14/2026
Document 12345678
Pay Summary Current $4,777.32 $4,771.42 $1,268.88 $12.05 $3,496.39
Bonus 401K Pre
$0.00
$14,537.51
$0.00
$7,268.76
401K Pretax
$0.00
$9,962.49
$0.00
$4,981.23
`
	statement, err := ParseText(text)
	if err != nil {
		t.Fatal(err)
	}
	if statement.Employee401KYTDCents != 2450000 || statement.Employer401KYTDCents != 1224999 {
		t.Fatalf("unexpected history YTD amounts: employee=%d employer=%d", statement.Employee401KYTDCents, statement.Employer401KYTDCents)
	}
}
