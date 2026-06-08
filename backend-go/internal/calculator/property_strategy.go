package calculator

type PropertyStrategy struct{}

func NewPropertyStrategy() *PropertyStrategy {
	return &PropertyStrategy{}
}

func (s *PropertyStrategy) CanHandle(taxType TaxType) bool {
	return taxType == TaxTypeProperty
}

func (s *PropertyStrategy) Calculate(input TaxCalculationInput) (TaxCalculationResult, error) {
	assessedValue := input.Amount * 0.8
	taxableAmount := assessedValue - input.Exemptions
	if taxableAmount < 0 {
		taxableAmount = 0
	}

	taxRate := input.TaxRate
	if taxRate <= 0 {
		taxRate = 0.015
	}

	taxAmount := taxableAmount * taxRate

	return TaxCalculationResult{
		TaxAmount:     taxAmount,
		TaxRate:       taxRate,
		TaxableAmount: taxableAmount,
		Breakdown: []TaxBreakdownItem{
			{Description: "Market Value", Amount: input.Amount},
			{Description: "Assessed Value (80%)", Amount: assessedValue},
			{Description: "Exemptions", Amount: -input.Exemptions},
			{Description: "Taxable Value", Amount: taxableAmount},
			{Description: "Property Tax", Amount: taxAmount},
		},
	}, nil
}
