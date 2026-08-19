package projection

import (
	"math"
	"time"

	"life/internal/models"
)

type Milestone struct {
	Age                   int
	Year                  int
	NetWorthCents         int64
	ContributionsCents    int64
	InvestmentGrowthCents int64
}

type Summary struct {
	AnnualGrossPayCents       int64
	AnnualNetPayCents         int64
	AnnualSpendingCents       int64
	AnnualCashSurplusCents    int64
	AnnualPostIRASurplusCents int64
	AnnualUninvestedCents     int64
	AnnualEmployee401KCents   int64
	AnnualEmployer401KCents   int64
	AnnualIRAContribution     int64
	AnnualTaxableContribution int64
	AnnualStockContribution   int64
	Milestones                []Milestone
}

func AgeOn(birthDate, date time.Time) int {
	age := date.Year() - birthDate.Year()
	if date.Month() < birthDate.Month() || (date.Month() == birthDate.Month() && date.Day() < birthDate.Day()) {
		age--
	}
	return age
}

func Calculate(startingInvestedCents, startingCashCents int64, settings models.ProjectionSettings, asOf time.Time, ages []int) Summary {
	base := float64(settings.BaseSalaryCents)
	gross := base * (1 + settings.BonusRate)
	spending := float64(settings.MonthlySpendingCents * 12)
	employee401K := float64(settings.Employee401KAnnualCents)
	employer401K := employee401K * settings.EmployerMatchRate
	netStock := float64(settings.MonthlyNetStockCents * 12)
	cashAvailable := math.Max(0, gross*settings.NetPayRate-spending)
	ira := math.Min(float64(settings.IRAAnnualCents), cashAvailable)
	taxable := math.Max(0, cashAvailable-ira) * settings.SurplusInvestedRate

	summary := Summary{
		AnnualGrossPayCents:       int64(math.Round(gross)),
		AnnualNetPayCents:         int64(math.Round(gross * settings.NetPayRate)),
		AnnualSpendingCents:       int64(math.Round(spending)),
		AnnualCashSurplusCents:    int64(math.Round(cashAvailable)),
		AnnualPostIRASurplusCents: int64(math.Round(math.Max(0, cashAvailable-ira))),
		AnnualUninvestedCents:     int64(math.Round(math.Max(0, cashAvailable-ira-taxable))),
		AnnualEmployee401KCents:   int64(math.Round(employee401K)),
		AnnualEmployer401KCents:   int64(math.Round(employer401K)),
		AnnualIRAContribution:     int64(math.Round(ira)),
		AnnualTaxableContribution: int64(math.Round(taxable)),
		AnnualStockContribution:   int64(math.Round(netStock)),
	}

	currentAge := AgeOn(settings.BirthDate, asOf)
	investedBalance := float64(startingInvestedCents)
	cashBalance := float64(startingCashCents)
	contributions := investedBalance + cashBalance
	targets := make(map[int]struct{}, len(ages))
	for _, age := range ages {
		if age > currentAge {
			targets[age] = struct{}{}
		}
	}
	for age := currentAge + 1; len(targets) > 0 && age <= 100; age++ {
		investedBalance *= 1 + settings.RealReturnRate
		cashBalance *= 1 + settings.CashRealReturnRate
		if age <= settings.RetirementAge {
			growthYears := age - currentAge - 1
			incomeFactor := math.Pow(1+settings.RealIncomeGrowthRate, float64(growthYears))
			yearGross := gross * incomeFactor
			yearCashAvailable := math.Max(0, yearGross*settings.NetPayRate-spending)
			yearIRA := math.Min(float64(settings.IRAAnnualCents), yearCashAvailable)
			yearTaxable := math.Max(0, yearCashAvailable-yearIRA) * settings.SurplusInvestedRate
			yearContribution := employee401K + employer401K + netStock + yearIRA + yearTaxable
			investedBalance += yearContribution
			contributions += yearContribution
		} else {
			withdrawal := math.Min(spending, investedBalance)
			investedBalance -= withdrawal
			remaining := spending - withdrawal
			if remaining > 0 {
				cashBalance = math.Max(0, cashBalance-remaining)
			}
			contributions -= spending
		}
		if _, ok := targets[age]; ok {
			netWorth := investedBalance + cashBalance
			summary.Milestones = append(summary.Milestones, Milestone{
				Age:                   age,
				Year:                  settings.BirthDate.Year() + age,
				NetWorthCents:         int64(math.Round(netWorth)),
				ContributionsCents:    int64(math.Round(contributions)),
				InvestmentGrowthCents: int64(math.Round(netWorth - contributions)),
			})
			delete(targets, age)
		}
	}
	return summary
}
