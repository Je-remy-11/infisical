package calculator

type CorporateStrategy struct{}

func NewCorporateStrategy() *CorporateStrategy {
	return &CorporateStrategy{}
}

func (s *CorporateStrategy) CanHandle(taxType TaxType) bool {
	return taxType == TaxTypeCorporate
}

func (s *CorporateStrategy) Calculate(input TaxCalculationInput) (TaxCalculationResult, error) {
	taxableAmount := input.Amount - input.Deductions
	if taxableAmount < 0 {
		taxableAmount = 0
	}

	taxRate := input.TaxRate
	if taxRate <= 0 {
		taxRate = 0.21
	}

	taxAmount := taxableAmount * taxRate

	return TaxCalculationResult{
		TaxAmount:     taxAmount,
		TaxRate:       taxRate,
		TaxableAmount: taxableAmount,
		Breakdown: []TaxBreakdownItem{
			{Description: "Corporate Income", Amount: input.Amount},
			{Description: "Business Deductions", Amount: -input.Deductions},
			{Description: "Taxable Income", Amount: taxableAmount},
			{Description: "Corporate Tax", Amount: taxAmount},
		},
	}, nil
}
