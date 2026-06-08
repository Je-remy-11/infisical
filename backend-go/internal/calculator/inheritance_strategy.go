package calculator

type InheritanceStrategy struct{}

func NewInheritanceStrategy() *InheritanceStrategy {
	return &InheritanceStrategy{}
}

func (s *InheritanceStrategy) CanHandle(taxType TaxType) bool {
	return taxType == TaxTypeInheritance
}

func (s *InheritanceStrategy) Calculate(input TaxCalculationInput) (TaxCalculationResult, error) {
	taxableAmount := input.Amount - input.Exemptions
	if taxableAmount < 0 {
		taxableAmount = 0
	}

	taxRate := input.TaxRate
	if taxRate <= 0 {
		taxRate = 0.40
	}

	taxAmount := taxableAmount * taxRate

	return TaxCalculationResult{
		TaxAmount:     taxAmount,
		TaxRate:       taxRate,
		TaxableAmount: taxableAmount,
		Breakdown: []TaxBreakdownItem{
			{Description: "Inheritance Value", Amount: input.Amount},
			{Description: "Exemptions", Amount: -input.Exemptions},
			{Description: "Taxable Amount", Amount: taxableAmount},
			{Description: "Inheritance Tax", Amount: taxAmount},
		},
	}, nil
}
