package calculator

type CapitalGainStrategy struct{}

func NewCapitalGainStrategy() *CapitalGainStrategy {
	return &CapitalGainStrategy{}
}

func (s *CapitalGainStrategy) CanHandle(taxType TaxType) bool {
	return taxType == TaxTypeCapitalGain
}

func (s *CapitalGainStrategy) Calculate(input TaxCalculationInput) (TaxCalculationResult, error) {
	taxableAmount := input.Amount - input.Deductions
	if taxableAmount < 0 {
		taxableAmount = 0
	}

	taxRate := input.TaxRate
	if taxRate <= 0 {
		taxRate = 0.20
	}

	taxAmount := taxableAmount * taxRate

	return TaxCalculationResult{
		TaxAmount:     taxAmount,
		TaxRate:       taxRate,
		TaxableAmount: taxableAmount,
		Breakdown: []TaxBreakdownItem{
			{Description: "Capital Gain", Amount: input.Amount},
			{Description: "Deductions", Amount: -input.Deductions},
			{Description: "Taxable Gain", Amount: taxableAmount},
			{Description: "Capital Gain Tax", Amount: taxAmount},
		},
	}, nil
}
