package calculator

type SalesStrategy struct{}

func NewSalesStrategy() *SalesStrategy {
	return &SalesStrategy{}
}

func (s *SalesStrategy) CanHandle(taxType TaxType) bool {
	return taxType == TaxTypeSales
}

func (s *SalesStrategy) Calculate(input TaxCalculationInput) (TaxCalculationResult, error) {
	taxableAmount := input.Amount
	taxRate := input.TaxRate
	if taxRate <= 0 {
		taxRate = 0.08
	}

	taxAmount := taxableAmount * taxRate

	return TaxCalculationResult{
		TaxAmount:     taxAmount,
		TaxRate:       taxRate,
		TaxableAmount: taxableAmount,
		Breakdown: []TaxBreakdownItem{
			{Description: "Sales Amount", Amount: input.Amount},
			{Description: "Sales Tax", Amount: taxAmount},
		},
	}, nil
}
