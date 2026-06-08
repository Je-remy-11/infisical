package calculator

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrUnsupportedTaxType = errors.New("unsupported tax type")
	ErrInvalidInput       = errors.New("invalid tax input")
)

type TaxType string

const (
	TaxTypeIncome    TaxType = "income"
	TaxTypeVAT       TaxType = "vat"
	TaxTypeCorporate TaxType = "corporate"
	TaxTypeProperty  TaxType = "property"
	TaxTypeStampDuty TaxType = "stamp_duty"
)

type TaxInput struct {
	TaxType    TaxType  `json:"tax_type"`
	Amount     float64  `json:"amount"`
	Deductions float64  `json:"deductions"`
	Rate       float64  `json:"rate"`
	IsHighTech *bool    `json:"is_high_tech,omitempty"`
	PropertyID *string  `json:"property_id,omitempty"`
	DocType    *string  `json:"doc_type,omitempty"`
}

type TaxResult struct {
	TaxAmount  float64  `json:"tax_amount"`
	Taxable    float64  `json:"taxable"`
	Rate       float64  `json:"rate"`
	Breakdown  string   `json:"breakdown"`
	Err        error    `json:"-"`
}

type TaxStrategy interface {
	Calculate(ctx context.Context, input *TaxInput) (*TaxResult, error)
	TaxType() TaxType
	Validate(input *TaxInput) error
}

var ErrStrategyRegistry = errors.New("strategy registry error")

type StrategyRegistry struct {
	strategies map[TaxType]TaxStrategy
}

func NewStrategyRegistry() *StrategyRegistry {
	r := &StrategyRegistry{
		strategies: make(map[TaxType]TaxStrategy),
	}
	r.registerDefaults()
	return r
}

func (r *StrategyRegistry) Register(s TaxStrategy) {
	r.strategies[s.TaxType()] = s
}

func (r *StrategyRegistry) Get(taxType TaxType) (TaxStrategy, error) {
	s, ok := r.strategies[taxType]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedTaxType, taxType)
	}
	return s, nil
}

func (r *StrategyRegistry) registerDefaults() {
	r.Register(&IncomeTaxStrategy{})
	r.Register(&VATStrategy{})
	r.Register(&CorporateTaxStrategy{})
	r.Register(&PropertyTaxStrategy{})
	r.Register(&StampDutyStrategy{})
}

func ValidateInput(input *TaxInput) error {
	if input == nil {
		return fmt.Errorf("%w: input is nil", ErrInvalidInput)
	}
	if input.Amount < 0 {
		return fmt.Errorf("%w: amount cannot be negative", ErrInvalidInput)
	}
	if input.Deductions < 0 {
		return fmt.Errorf("%w: deductions cannot be negative", ErrInvalidInput)
	}
	if input.TaxType == "" {
		return fmt.Errorf("%w: tax type is required", ErrInvalidInput)
	}
	return nil
}

func CalculateTax(ctx context.Context, input *TaxInput) (*TaxResult, error) {
	if err := ValidateInput(input); err != nil {
		return nil, err
	}

	registry := NewStrategyRegistry()
	strategy, err := registry.Get(input.TaxType)
	if err != nil {
		return nil, err
	}

	if err := strategy.Validate(input); err != nil {
		return nil, err
	}

	return strategy.Calculate(ctx, input)
}