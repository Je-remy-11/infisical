package calculator

import (
	"context"
	"fmt"
)

type rateBracket struct {
	lowerBound float64
	upperBound float64
	rate       float64
	deduction  float64
}

var incomeTaxBrackets = []rateBracket{
	{0, 36000, 0.03, 0},
	{36000, 144000, 0.10, 2520},
	{144000, 300000, 0.20, 16920},
	{300000, 420000, 0.25, 31920},
	{420000, 660000, 0.30, 52920},
	{660000, 960000, 0.35, 85920},
	{960000, 1e18, 0.45, 181920},
}

type IncomeTaxStrategy struct{}

func (s *IncomeTaxStrategy) TaxType() TaxType { return TaxTypeIncome }

func (s *IncomeTaxStrategy) Validate(input *TaxInput) error {
	if input.Amount == 0 && input.Deductions == 0 {
		return fmt.Errorf("%w: income tax requires amount or deductions", ErrInvalidInput)
	}
	return nil
}

func (s *IncomeTaxStrategy) Calculate(_ context.Context, input *TaxInput) (*TaxResult, error) {
	taxable := input.Amount - input.Deductions - 60000
	if taxable <= 0 {
		return &TaxResult{
			TaxAmount: 0,
			Taxable:   taxable,
			Rate:      0,
			Breakdown: fmt.Sprintf("个人所得税: 应纳税所得额 %.2f <= 0, 无需缴税", taxable),
		}, nil
	}

	var taxAmount float64
	var appliedRate float64
	for _, b := range incomeTaxBrackets {
		if taxable > b.lowerBound {
			taxAmount = taxable*b.rate - b.deduction
			appliedRate = b.rate
		}
	}

	return &TaxResult{
		TaxAmount: taxAmount,
		Taxable:   taxable,
		Rate:      appliedRate,
		Breakdown: fmt.Sprintf(
			"个人所得税: 应纳税所得额 %.2f = 收入 %.2f - 扣除 %.2f - 起征点 60000, 税额 %.2f (税率 %.0f%%)",
			taxable, input.Amount, input.Deductions, taxAmount, appliedRate*100,
		),
	}, nil
}

type VATStrategy struct{}

func (s *VATStrategy) TaxType() TaxType { return TaxTypeVAT }

func (s *VATStrategy) Validate(input *TaxInput) error {
	if input.Amount <= 0 {
		return fmt.Errorf("%w: VAT requires positive amount", ErrInvalidInput)
	}
	if input.Rate <= 0 {
		return fmt.Errorf("%w: VAT requires positive rate", ErrInvalidInput)
	}
	return nil
}

func (s *VATStrategy) Calculate(_ context.Context, input *TaxInput) (*TaxResult, error) {
	outputVAT := input.Amount * input.Rate
	inputVAT := input.Deductions * input.Rate
	taxAmount := outputVAT - inputVAT

	return &TaxResult{
		TaxAmount: taxAmount,
		Taxable:   input.Amount - input.Deductions,
		Rate:      input.Rate,
		Breakdown: fmt.Sprintf(
			"增值税: 销项税额 %.2f (%.2f × %.0f%%) - 进项税额 %.2f (%.2f × %.0f%%) = %.2f",
			outputVAT, input.Amount, input.Rate*100,
			inputVAT, input.Deductions, input.Rate*100,
			taxAmount,
		),
	}, nil
}

type CorporateTaxStrategy struct{}

func (s *CorporateTaxStrategy) TaxType() TaxType { return TaxTypeCorporate }

func (s *CorporateTaxStrategy) Validate(input *TaxInput) error {
	if input.Amount < 0 {
		return fmt.Errorf("%w: corporate income cannot be negative", ErrInvalidInput)
	}
	return nil
}

func (s *CorporateTaxStrategy) Calculate(_ context.Context, input *TaxInput) (*TaxResult, error) {
	taxable := input.Amount - input.Deductions
	if taxable <= 0 {
		return &TaxResult{
			TaxAmount: 0,
			Taxable:   taxable,
			Rate:      0,
			Breakdown: fmt.Sprintf("企业所得税: 应纳税所得额 %.2f <= 0, 无需缴税", taxable),
		}, nil
	}

	rate := 0.25
	specialLabel := "标准税率"

	if input.IsHighTech != nil && *input.IsHighTech {
		rate = 0.15
		specialLabel = "高新技术企业优惠税率"
	}

	taxAmount := taxable * rate

	return &TaxResult{
		TaxAmount: taxAmount,
		Taxable:   taxable,
		Rate:      rate,
		Breakdown: fmt.Sprintf(
			"企业所得税(%s): 应纳税所得额 %.2f = 收入 %.2f - 扣除 %.2f, 税额 %.2f (税率 %.0f%%)",
			specialLabel, taxable, input.Amount, input.Deductions, taxAmount, rate*100,
		),
	}, nil
}

type PropertyTaxStrategy struct{}

func (s *PropertyTaxStrategy) TaxType() TaxType { return TaxTypeProperty }

func (s *PropertyTaxStrategy) Validate(input *TaxInput) error {
	if input.Amount <= 0 {
		return fmt.Errorf("%w: property value must be positive", ErrInvalidInput)
	}
	return nil
}

func (s *PropertyTaxStrategy) Calculate(_ context.Context, input *TaxInput) (*TaxResult, error) {
	taxableValue := input.Amount * 0.7
	rate := 0.012

	if input.Rate > 0 {
		rate = input.Rate
	}

	taxAmount := taxableValue * rate
	propInfo := ""
	if input.PropertyID != nil {
		propInfo = fmt.Sprintf(" [房产编号: %s]", *input.PropertyID)
	}

	return &TaxResult{
		TaxAmount: taxAmount,
		Taxable:   taxableValue,
		Rate:      rate,
		Breakdown: fmt.Sprintf(
			"房产税%s: 计税余值 %.2f = 原值 %.2f × 70%%, 税额 %.2f (税率 %.1f%%)",
			propInfo, taxableValue, input.Amount, taxAmount, rate*100,
		),
	}, nil
}

var stampDutyRates = map[string]float64{
	"purchase":      0.0005,
	"loan":          0.00005,
	"lease":         0.001,
	"transport":     0.0005,
	"warehouse":     0.001,
	"property":      0.0005,
	"tech_contract": 0.0003,
}

type StampDutyStrategy struct{}

func (s *StampDutyStrategy) TaxType() TaxType { return TaxTypeStampDuty }

func (s *StampDutyStrategy) Validate(input *TaxInput) error {
	if input.Amount <= 0 {
		return fmt.Errorf("%w: stamp duty requires positive amount", ErrInvalidInput)
	}
	if input.DocType == nil || *input.DocType == "" {
		return fmt.Errorf("%w: stamp duty requires document type", ErrInvalidInput)
	}
	if _, ok := stampDutyRates[*input.DocType]; !ok {
		return fmt.Errorf("%w: unknown document type: %s", ErrInvalidInput, *input.DocType)
	}
	return nil
}

func (s *StampDutyStrategy) Calculate(_ context.Context, input *TaxInput) (*TaxResult, error) {
	rate := stampDutyRates[*input.DocType]
	if input.Rate > 0 {
		rate = input.Rate
	}

	taxAmount := input.Amount * rate

	return &TaxResult{
		TaxAmount: taxAmount,
		Taxable:   input.Amount,
		Rate:      rate,
		Breakdown: fmt.Sprintf(
			"印花税(%s): 金额 %.2f × %.3f%% = %.2f",
			*input.DocType, input.Amount, rate*100, taxAmount,
		),
	}, nil
}