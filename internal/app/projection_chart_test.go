package app

import (
	"strings"
	"testing"

	"life/internal/projection"
)

func TestRenderNetWorthLineChart(t *testing.T) {
	chart := renderNetWorthLineChart([]projection.Milestone{
		{Age: 29, NetWorthCents: 28486400, ContributionsCents: 28486400},
		{Age: 35, NetWorthCents: 115676456, ContributionsCents: 85000000},
		{Age: 45, NetWorthCents: 350000000, ContributionsCents: 180000000},
		{Age: 65, NetWorthCents: 1350314211, ContributionsCents: 100000000},
	}, 45, 140)
	for _, expected := range []string{"NET WORTH TRAJECTORY", "net worth", "net contributions", "retirement at age 45"} {
		if !strings.Contains(chart, expected) {
			t.Fatalf("chart does not contain %q", expected)
		}
	}
}
