package calculator

import (
	"errors"
	"math"
)

var (
	ErrUnsupportedTaxType = errors.New("unsupported tax type")
	ErrInvalidAmount      = errors.New("invalid amount: must be non-negative")
)

type IncomeTaxStrategy struct{}

func (s *IncomeTaxStrategy) SupportedType() TaxType {
	return TaxTypeIncome
}

func (s *IncomeTaxStrategy) Calculate(input TaxInput) (TaxResult, error) {
	if input.Amount < 0 {
		return TaxResult{}, ErrInvalidAmount
	}

	if input.IsExempt {
		return TaxResult{
			TaxAmount: 0,
			TaxRate:   0,
			TaxType:   TaxTypeIncome,
			Breakdown: map[string]float64{"exempt": 0},
		}, nil
	}

	taxableIncome := input.Amount - input.Deductions
	if taxableIncome < 0 {
		taxableIncome = 0
	}

	brackets := getIncomeBrackets(input.Region)
	taxAmount := 0.0
	breakdown := make(map[string]float64)
	remaining := taxableIncome

	for _, bracket := range brackets {
		if remaining <= 0 {
			break
		}

		taxableInBracket := math.Min(remaining, bracket.Max-bracket.Min)
		if bracket.Max == 0 {
			taxableInBracket = remaining
		}

		taxInBracket := taxableInBracket * bracket.Rate
		taxAmount += taxInBracket
		breakdown[bracket.Label] = taxInBracket
		remaining -= taxableInBracket
	}

	var effectiveRate float64
	if input.Amount > 0 {
		effectiveRate = taxAmount / input.Amount
	}

	return TaxResult{
		TaxAmount: math.Round(taxAmount*100) / 100,
		TaxRate:   math.Round(effectiveRate*10000) / 10000,
		TaxType:   TaxTypeIncome,
		Breakdown: breakdown,
	}, nil
}

type CorporateTaxStrategy struct{}

func (s *CorporateTaxStrategy) SupportedType() TaxType {
	return TaxTypeCorporate
}

func (s *CorporateTaxStrategy) Calculate(input TaxInput) (TaxResult, error) {
	if input.Amount < 0 {
		return TaxResult{}, ErrInvalidAmount
	}

	if input.IsExempt {
		return TaxResult{
			TaxAmount: 0,
			TaxRate:   0,
			TaxType:   TaxTypeCorporate,
			Breakdown: map[string]float64{"exempt": 0},
		}, nil
	}

	taxableIncome := input.Amount - input.Deductions
	if taxableIncome < 0 {
		taxableIncome = 0
	}

	rate := getCorporateRate(input.Region)
	taxAmount := taxableIncome * rate

	var effectiveRate float64
	if input.Amount > 0 {
		effectiveRate = taxAmount / input.Amount
	}

	return TaxResult{
		TaxAmount: math.Round(taxAmount*100) / 100,
		TaxRate:   math.Round(effectiveRate*10000) / 10000,
		TaxType:   TaxTypeCorporate,
		Breakdown: map[string]float64{"base_tax": taxAmount},
	}, nil
}

type ValueAddedTaxStrategy struct{}

func (s *ValueAddedTaxStrategy) SupportedType() TaxType {
	return TaxTypeValueAdded
}

func (s *ValueAddedTaxStrategy) Calculate(input TaxInput) (TaxResult, error) {
	if input.Amount < 0 {
		return TaxResult{}, ErrInvalidAmount
	}

	if input.IsExempt {
		return TaxResult{
			TaxAmount: 0,
			TaxRate:   0,
			TaxType:   TaxTypeValueAdded,
			Breakdown: map[string]float64{"exempt": 0},
		}, nil
	}

	rate := getVATRate(input.Region)
	taxAmount := input.Amount * rate

	return TaxResult{
		TaxAmount: math.Round(taxAmount*100) / 100,
		TaxRate:   math.Round(rate*10000) / 10000,
		TaxType:   TaxTypeValueAdded,
		Breakdown: map[string]float64{"vat": taxAmount},
	}, nil
}

type SalesTaxStrategy struct{}

func (s *SalesTaxStrategy) SupportedType() TaxType {
	return TaxTypeSales
}

func (s *SalesTaxStrategy) Calculate(input TaxInput) (TaxResult, error) {
	if input.Amount < 0 {
		return TaxResult{}, ErrInvalidAmount
	}

	if input.IsExempt {
		return TaxResult{
			TaxAmount: 0,
			TaxRate:   0,
			TaxType:   TaxTypeSales,
			Breakdown: map[string]float64{"exempt": 0},
		}, nil
	}

	rate := getSalesTaxRate(input.Region)
	taxAmount := input.Amount * rate

	return TaxResult{
		TaxAmount: math.Round(taxAmount*100) / 100,
		TaxRate:   math.Round(rate*10000) / 10000,
		TaxType:   TaxTypeSales,
		Breakdown: map[string]float64{"sales_tax": taxAmount},
	}, nil
}

type PropertyTaxStrategy struct{}

func (s *PropertyTaxStrategy) SupportedType() TaxType {
	return TaxTypeProperty
}

func (s *PropertyTaxStrategy) Calculate(input TaxInput) (TaxResult, error) {
	if input.Amount < 0 {
		return TaxResult{}, ErrInvalidAmount
	}

	if input.IsExempt {
		return TaxResult{
			TaxAmount: 0,
			TaxRate:   0,
			TaxType:   TaxTypeProperty,
			Breakdown: map[string]float64{"exempt": 0},
		}, nil
	}

	rate := getPropertyTaxRate(input.Region)
	assessedValue := input.Amount * 0.8
	taxAmount := assessedValue * rate

	return TaxResult{
		TaxAmount: math.Round(taxAmount*100) / 100,
		TaxRate:   math.Round(rate*10000) / 10000,
		TaxType:   TaxTypeProperty,
		Breakdown: map[string]float64{
			"assessed_value": assessedValue,
			"property_tax":   taxAmount,
		},
	}, nil
}

type CapitalGainTaxStrategy struct{}

func (s *CapitalGainTaxStrategy) SupportedType() TaxType {
	return TaxTypeCapitalGain
}

func (s *CapitalGainTaxStrategy) Calculate(input TaxInput) (TaxResult, error) {
	if input.Amount < 0 {
		return TaxResult{}, ErrInvalidAmount
	}

	if input.IsExempt {
		return TaxResult{
			TaxAmount: 0,
			TaxRate:   0,
			TaxType:   TaxTypeCapitalGain,
			Breakdown: map[string]float64{"exempt": 0},
		}, nil
	}

	taxableGain := input.Amount - input.Deductions
	if taxableGain < 0 {
		taxableGain = 0
	}

	rate := getCapitalGainRate(input.Year, input.Region)
	taxAmount := taxableGain * rate

	var effectiveRate float64
	if input.Amount > 0 {
		effectiveRate = taxAmount / input.Amount
	}

	return TaxResult{
		TaxAmount: math.Round(taxAmount*100) / 100,
		TaxRate:   math.Round(effectiveRate*10000) / 10000,
		TaxType:   TaxTypeCapitalGain,
		Breakdown: map[string]float64{"capital_gain_tax": taxAmount},
	}, nil
}
