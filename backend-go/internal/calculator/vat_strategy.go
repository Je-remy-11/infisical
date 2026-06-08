package calculator

type VATStrategy struct{}

func NewVATStrategy() *VATStrategy {
	return &VATStrategy{}
}

func (s *VATStrategy) CanHandle(taxType TaxType) bool {
	return taxType == TaxTypeVAT
}

func (s *VATStrategy) Calculate(input TaxCalculationInput) (TaxCalculationResult, error) {
	taxableAmount := input.Amount - input.Exemptions
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
			{Description: "Gross Amount", Amount: input.Amount},
			{Description: "Exemptions", Amount: -input.Exemptions},
			{Description: "Taxable Amount", Amount: taxableAmount},
			{Description: "VAT Tax", Amount: taxAmount},
		},
	}, nil
}
