package projection

import (
	"testing"
	"time"

	"life/internal/models"
)

func TestAgeOn(t *testing.T) {
	birth := time.Date(1997, time.July, 13, 0, 0, 0, 0, time.UTC)
	if got := AgeOn(birth, time.Date(2027, time.July, 12, 0, 0, 0, 0, time.UTC)); got != 29 {
		t.Fatalf("age before birthday = %d", got)
	}
	if got := AgeOn(birth, time.Date(2027, time.July, 13, 0, 0, 0, 0, time.UTC)); got != 30 {
		t.Fatalf("age on birthday = %d", got)
	}
}

func TestCalculateUsesSpendingAndAvoidsDoubleCountingIRA(t *testing.T) {
	s := models.DefaultProjectionSettings()
	s.BirthDate = time.Date(1997, time.July, 13, 0, 0, 0, 0, time.UTC)
	s.BaseSalaryCents = 10000000
	s.BonusRate = 0
	s.NetPayRate = 0.60
	s.MonthlySpendingCents = 400000
	s.IRAAnnualCents = 700000
	s.Employee401KAnnualCents = 0
	s.EmployerMatchRate = 0
	s.MonthlyNetStockCents = 0
	s.RealReturnRate = 0
	s.RealIncomeGrowthRate = 0

	result := Calculate(0, 0, s, time.Date(2026, time.August, 18, 0, 0, 0, 0, time.UTC), []int{30})
	if result.AnnualNetPayCents != 6000000 {
		t.Fatalf("annual net pay = %d, want 6000000", result.AnnualNetPayCents)
	}
	if result.AnnualIRAContribution != 700000 || result.AnnualTaxableContribution != 500000 {
		t.Fatalf("IRA/taxable = %d/%d, want 700000/500000", result.AnnualIRAContribution, result.AnnualTaxableContribution)
	}
	if len(result.Milestones) != 1 || result.Milestones[0].NetWorthCents != 1200000 {
		t.Fatalf("first milestone = %+v", result.Milestones)
	}
}

func TestCalculateAppliesSurplusInvestmentRate(t *testing.T) {
	s := models.DefaultProjectionSettings()
	s.BirthDate = time.Date(1997, time.July, 13, 0, 0, 0, 0, time.UTC)
	s.BaseSalaryCents = 10000000
	s.BonusRate = 0
	s.NetPayRate = 0.60
	s.MonthlySpendingCents = 400000
	s.IRAAnnualCents = 700000
	s.SurplusInvestedRate = 0.50
	s.Employee401KAnnualCents = 0
	s.EmployerMatchRate = 0
	s.MonthlyNetStockCents = 0
	s.RealReturnRate = 0
	s.RealIncomeGrowthRate = 0

	result := Calculate(0, 0, s, time.Date(2026, time.August, 18, 0, 0, 0, 0, time.UTC), []int{30})
	if result.AnnualTaxableContribution != 250000 {
		t.Fatalf("taxable contribution = %d, want 250000", result.AnnualTaxableContribution)
	}
}

func TestCalculateUsesSeparateCashReturn(t *testing.T) {
	s := models.DefaultProjectionSettings()
	s.BirthDate = time.Date(1997, time.July, 13, 0, 0, 0, 0, time.UTC)
	s.BaseSalaryCents = 0
	s.BonusRate = 0
	s.MonthlySpendingCents = 0
	s.Employee401KAnnualCents = 0
	s.EmployerMatchRate = 0
	s.MonthlyNetStockCents = 0
	s.IRAAnnualCents = 0
	s.RealReturnRate = 0.07
	s.CashRealReturnRate = 0.015

	result := Calculate(100000, 100000, s, time.Date(2026, time.August, 18, 0, 0, 0, 0, time.UTC), []int{30})
	if got := result.Milestones[0].NetWorthCents; got != 208500 {
		t.Fatalf("net worth = %d, want 208500", got)
	}
}

func TestCalculateStopsContributionsAndWithdrawsAfterRetirement(t *testing.T) {
	s := models.DefaultProjectionSettings()
	s.BirthDate = time.Date(1997, time.July, 13, 0, 0, 0, 0, time.UTC)
	s.BaseSalaryCents = 0
	s.BonusRate = 0
	s.MonthlySpendingCents = 10000
	s.Employee401KAnnualCents = 0
	s.EmployerMatchRate = 0
	s.MonthlyNetStockCents = 0
	s.IRAAnnualCents = 0
	s.RealReturnRate = 0
	s.CashRealReturnRate = 0
	s.RetirementAge = 29

	result := Calculate(1000000, 0, s, time.Date(2026, time.August, 18, 0, 0, 0, 0, time.UTC), []int{30})
	if got := result.Milestones[0].NetWorthCents; got != 880000 {
		t.Fatalf("net worth after first retired year = %d, want 880000", got)
	}
}
