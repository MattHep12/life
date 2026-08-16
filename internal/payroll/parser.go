package payroll

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ledongthuc/pdf"

	"life/internal/models"
)

var (
	periodStartPattern = regexp.MustCompile(`(?i)Period\s+Start\s+Date\s+(\d{2}/\d{2}/\d{4})`)
	periodEndPattern   = regexp.MustCompile(`(?i)Period\s+End\s+Date\s+(\d{2}/\d{2}/\d{4})`)
	payDatePattern     = regexp.MustCompile(`(?i)Pay\s+Date\s+(\d{2}/\d{2}/\d{4})`)
	documentPattern    = regexp.MustCompile(`(?i)Document\s+(\d+)`)
	paySummaryPattern  = regexp.MustCompile(`(?is)Pay\s+Summary.*?Current\s+\$?([\d,]+\.\d{2})\s+\$?([\d,]+\.\d{2})\s+\$?([\d,]+\.\d{2})\s+\$?([\d,]+\.\d{2})\s+\$?([\d,]+\.\d{2})`)
)

func ParsePDF(path string) (models.PayrollStatement, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return models.PayrollStatement{}, fmt.Errorf("open PDF: %w", err)
	}
	defer file.Close()

	plainText, err := reader.GetPlainText()
	if err != nil {
		return models.PayrollStatement{}, fmt.Errorf("extract PDF text: %w", err)
	}

	content, err := io.ReadAll(plainText)
	if err != nil {
		return models.PayrollStatement{}, fmt.Errorf("read PDF text: %w", err)
	}

	statement, err := ParseText(string(content))
	if err != nil {
		return models.PayrollStatement{}, err
	}
	statement.SourceFilename = filepath.Base(path)

	return statement, nil
}

func ParseText(text string) (models.PayrollStatement, error) {
	periodStart, err := extractDate(text, periodStartPattern, "period start date")
	if err != nil {
		return models.PayrollStatement{}, err
	}
	periodEnd, err := extractDate(text, periodEndPattern, "period end date")
	if err != nil {
		return models.PayrollStatement{}, err
	}
	payDate, err := extractDate(text, payDatePattern, "pay date")
	if err != nil {
		return models.PayrollStatement{}, err
	}

	documentMatch := documentPattern.FindStringSubmatch(text)
	if len(documentMatch) != 2 {
		return models.PayrollStatement{}, fmt.Errorf("document number not found; unsupported payroll PDF layout")
	}

	summaryMatch := paySummaryPattern.FindStringSubmatch(text)
	if len(summaryMatch) != 6 {
		return models.PayrollStatement{}, fmt.Errorf("current pay summary not found; unsupported payroll PDF layout")
	}

	amounts := make([]int64, 5)
	for i, value := range summaryMatch[1:] {
		amounts[i], err = parseMoney(value)
		if err != nil {
			return models.PayrollStatement{}, fmt.Errorf("parse pay summary: %w", err)
		}
	}

	return models.PayrollStatement{
		DocumentNumber:  documentMatch[1],
		PeriodStart:     periodStart,
		PeriodEnd:       periodEnd,
		PayDate:         payDate,
		GrossCents:      amounts[0],
		TaxableCents:    amounts[1],
		TaxesCents:      amounts[2],
		DeductionsCents: amounts[3],
		NetCents:        amounts[4],
	}, nil
}

func extractDate(text string, pattern *regexp.Regexp, label string) (time.Time, error) {
	match := pattern.FindStringSubmatch(text)
	if len(match) != 2 {
		return time.Time{}, fmt.Errorf("%s not found; unsupported payroll PDF layout", label)
	}

	value, err := time.Parse("01/02/2006", match[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("parse %s: %w", label, err)
	}
	return value, nil
}

func parseMoney(value string) (int64, error) {
	value = strings.ReplaceAll(value, ",", "")
	parts := strings.Split(value, ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		return 0, fmt.Errorf("invalid money value %q", value)
	}

	dollars, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}
	cents, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return dollars*100 + cents, nil
}
