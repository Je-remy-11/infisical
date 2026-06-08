package calculator

type PayrollStrategy struct{}

func NewPayrollStrategy() *PayrollStrategy {
	return &PayrollStrategy{}
}

func (s *PayrollStrategy) CanHandle(taxType TaxType) bool {
	return taxType == TaxTypePayroll
}

func (s *PayrollStrategy) Calculate(input TaxCalculationInput) (TaxCalculationResult, error) {
	taxableAmount := input.Amount - input.Exemptions
	if taxableAmount < 0 {
		taxableAmount = 0
	}

	taxRate := input.TaxRate
	if taxRate <= 0 {
		taxRate = 0.0765
	}

	taxAmount := taxableAmount * taxRate

	return TaxCalculationResult{
		TaxAmount:     taxAmount,
		TaxRate:       taxRate,
		TaxableAmount: taxableAmount,
		Breakdown: []TaxBreakdownItem{
			{Description: "Gross Payroll", Amount: input.Amount},
			{Description: "Exemptions", Amount: -input.Exemptions},
			{Description: "Taxable Payroll", Amount: taxableAmount},
			{Description: "Payroll Tax", Amount: taxAmount},
		},
	}, nil
}
