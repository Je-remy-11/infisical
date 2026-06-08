package calculator

import "testing"

func TestCalculateTax(t *testing.T) {
	calc := NewCalculator()

	input := TaxCalculationInput{
		Amount:     100000,
		TaxRate:    0.20,
		Deductions: 10000,
		Exemptions: 5000,
	}

	result, err := calc.CalculateTax(TaxTypeVAT, input)
	if err != nil {
		t.Fatalf("CalculateTax failed: %v", err)
	}

	t.Logf("VAT Tax: %.2f", result.TaxAmount)

	result, err = calc.CalculateTax(TaxTypeIncome, input)
	if err != nil {
		t.Fatalf("CalculateTax failed: %v", err)
	}

	t.Logf("Income Tax: %.2f", result.TaxAmount)
}
