package calculator

type ExciseStrategy struct{}

func NewExciseStrategy() *ExciseStrategy {
	return &ExciseStrategy{}
}

func (s *ExciseStrategy) CanHandle(taxType TaxType) bool {
	return taxType == TaxTypeExcise
}

func (s *ExciseStrategy) Calculate(input TaxCalculationInput) (TaxCalculationResult, error) {
	taxableAmount := input.Amount
	taxRate := input.TaxRate
	if taxRate <= 0 {
		taxRate = 0.10
	}

	taxAmount := taxableAmount * taxRate

	return TaxCalculationResult{
		TaxAmount:     taxAmount,
		TaxRate:       taxRate,
		TaxableAmount: taxableAmount,
		Breakdown: []TaxBreakdownItem{
			{Description: "Product Value", Amount: input.Amount},
			{Description: "Excise Tax", Amount: taxAmount},
		},
	}, nil
}
