package calculator

type CustomsStrategy struct{}

func NewCustomsStrategy() *CustomsStrategy {
	return &CustomsStrategy{}
}

func (s *CustomsStrategy) CanHandle(taxType TaxType) bool {
	return taxType == TaxTypeCustoms
}

func (s *CustomsStrategy) Calculate(input TaxCalculationInput) (TaxCalculationResult, error) {
	taxableAmount := input.Amount
	taxRate := input.TaxRate
	if taxRate <= 0 {
		taxRate = 0.15
	}

	taxAmount := taxableAmount * taxRate

	return TaxCalculationResult{
		TaxAmount:     taxAmount,
		TaxRate:       taxRate,
		TaxableAmount: taxableAmount,
		Breakdown: []TaxBreakdownItem{
			{Description: "Customs Value", Amount: input.Amount},
			{Description: "Customs Duty", Amount: taxAmount},
		},
	}, nil
}
