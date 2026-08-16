package app

import (
	"fmt"
	"strings"
)

func formatMoney(amount float64) string {
	formatted := fmt.Sprintf("%.2f", amount)
	sign := ""
	if strings.HasPrefix(formatted, "-") {
		sign = "-"
		formatted = strings.TrimPrefix(formatted, "-")
	}

	parts := strings.Split(formatted, ".")
	whole := parts[0]
	decimal := parts[1]

	var result strings.Builder

	for i, digit := range whole {
		if i > 0 && (len(whole)-i)%3 == 0 {
			result.WriteString(",")
		}

		result.WriteRune(digit)
	}

	return sign + "$" + result.String() + "." + decimal
}

func formatCents(cents int64) string {
	return formatMoney(float64(cents) / 100)
}
