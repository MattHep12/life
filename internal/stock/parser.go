package stock

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"life/internal/models"
)

func ParseTSV(reader io.Reader) ([]models.StockVest, error) {
	csvReader := csv.NewReader(reader)
	csvReader.Comma = '\t'
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read Schwab data: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("Schwab data has no rows")
	}

	headers := make(map[string]int)
	for i, header := range records[0] {
		headers[strings.TrimSpace(header)] = i
	}
	required := []string{"Symbol", "Share Type", "Deposit Date", "Date Acquired", "Acquisition Price", "Shares", "Available to Sell"}
	for _, header := range required {
		if _, ok := headers[header]; !ok {
			return nil, fmt.Errorf("missing required column %q", header)
		}
	}

	type key struct {
		symbol, vestDate, depositDate string
		priceCents                    int64
	}
	aggregated := make(map[key]*models.StockVest)
	for rowNumber, record := range records[1:] {
		if len(record) == 0 || blankRecord(record) {
			continue
		}
		if field(record, headers["Share Type"]) != "Restricted Stock" {
			continue
		}

		symbol := field(record, headers["Symbol"])
		depositDate, err := parseDate(field(record, headers["Deposit Date"]))
		if err != nil {
			return nil, fmt.Errorf("row %d deposit date: %w", rowNumber+2, err)
		}
		vestDate, err := parseDate(field(record, headers["Date Acquired"]))
		if err != nil {
			return nil, fmt.Errorf("row %d vest date: %w", rowNumber+2, err)
		}
		priceCents, err := parseScaledDecimal(field(record, headers["Acquisition Price"]), 100)
		if err != nil {
			return nil, fmt.Errorf("row %d acquisition price: %w", rowNumber+2, err)
		}
		grossShares, err := parseScaledDecimal(field(record, headers["Shares"]), models.ShareScale)
		if err != nil {
			return nil, fmt.Errorf("row %d shares: %w", rowNumber+2, err)
		}
		netShares, err := parseScaledDecimal(field(record, headers["Available to Sell"]), models.ShareScale)
		if err != nil {
			return nil, fmt.Errorf("row %d available shares: %w", rowNumber+2, err)
		}
		if netShares > grossShares {
			return nil, fmt.Errorf("row %d available shares exceed vested shares", rowNumber+2)
		}

		groupKey := key{symbol: symbol, vestDate: vestDate.Format("2006-01-02"), depositDate: depositDate.Format("2006-01-02"), priceCents: priceCents}
		vest := aggregated[groupKey]
		if vest == nil {
			vest = &models.StockVest{Symbol: symbol, VestDate: vestDate, DepositDate: depositDate, AcquisitionPriceCents: priceCents}
			aggregated[groupKey] = vest
		}
		vest.GrossSharesMicros += grossShares
		vest.NetSharesMicros += netShares
	}

	vests := make([]models.StockVest, 0, len(aggregated))
	for _, vest := range aggregated {
		vest.WithheldSharesMicros = vest.GrossSharesMicros - vest.NetSharesMicros
		vest.GrossValueCents = roundedProduct(vest.GrossSharesMicros, vest.AcquisitionPriceCents)
		vest.NetValueCents = roundedProduct(vest.NetSharesMicros, vest.AcquisitionPriceCents)
		vests = append(vests, *vest)
	}
	sort.Slice(vests, func(i, j int) bool { return vests[i].VestDate.Before(vests[j].VestDate) })
	return vests, nil
}

func field(record []string, index int) string {
	if index >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[index])
}

func blankRecord(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func parseDate(value string) (time.Time, error) {
	date, err := time.Parse("01/02/2006", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q", value)
	}
	return date, nil
}

func parseScaledDecimal(value string, scale int64) (int64, error) {
	value = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(value), "$", ""), ",", "")
	rational, ok := new(big.Rat).SetString(value)
	if !ok {
		return 0, fmt.Errorf("invalid decimal %q", value)
	}
	rational.Mul(rational, big.NewRat(scale, 1))
	quotient := new(big.Int).Quo(rational.Num(), rational.Denom())
	if !quotient.IsInt64() {
		return 0, fmt.Errorf("decimal %q is out of range", value)
	}
	return quotient.Int64(), nil
}

func roundedProduct(sharesMicros, priceCents int64) int64 {
	product := sharesMicros * priceCents
	return (product + models.ShareScale/2) / models.ShareScale
}

func FormatShares(micros int64) string {
	whole := micros / models.ShareScale
	fraction := micros % models.ShareScale
	return strings.TrimRight(strings.TrimRight(strconv.FormatInt(whole, 10)+"."+fmt.Sprintf("%06d", fraction), "0"), ".")
}
