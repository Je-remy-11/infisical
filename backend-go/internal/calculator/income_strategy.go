package calculator

type IncomeStrategy struct{}

func NewIncomeStrategy() *IncomeStrategy {
	return &IncomeStrategy{}
}

func (s *IncomeStrategy) CanHandle(taxType TaxType) bool {
	return taxType == TaxTypeIncome
}

func (s *IncomeStrategy) Calculate(input TaxCalculationInput) (TaxCalculationResult, error) {
	taxableAmount := input.Amount - input.Deductions - input.Exemptions
	if taxableAmount < 0 {
		taxableAmount = 0
	}

	var totalTax float64
	breakdown := []TaxBreakdownItem{
		{Description: "Gross Income", Amount: input.Amount},
		{Description: "Deductions", Amount: -input.Deductions},
		{Description: "Exemptions", Amount: -input.Exemptions},
		{Description: "Taxable Income", Amount: taxableAmount},
	}

	if len(input.TaxBrackets) > 0 {
		for _, bracket := range input.TaxBrackets {
			if taxableAmount <= bracket.MinAmount {
				continue
			}
			bracketAmount := min(taxableAmount, bracket.MaxAmount) - bracket.MinAmount
			if bracketAmount > 0 {
				bracketTax := bracketAmount * bracket.Rate
				totalTax += bracketTax
				breakdown = append(breakdown, TaxBreakdownItem{
					Description: "Bracket Tax",
					Amount:      bracketTax,
				})
			}
		}
	} else {
		taxRate := input.TaxRate
		if taxRate <= 0 {
			taxRate = 0.25
		}
		totalTax = taxableAmount * taxRate
		breakdown = append(breakdown, TaxBreakdownItem{
			Description: "Income Tax",
			Amount:      totalTax,
		})
	}

	return TaxCalculationResult{
		TaxAmount:     totalTax,
		TaxRate:       input.TaxRate,
		TaxableAmount: taxableAmount,
		Breakdown:     breakdown,
	}, nil
}
