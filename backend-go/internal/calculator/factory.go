package calculator

import (
	"fmt"
)

type TaxStrategyFactory interface {
	Create(taxType TaxType) (TaxStrategy, error)
}

type SimpleTaxFactory struct {
	registry *StrategyRegistry
}

func NewSimpleTaxFactory() *SimpleTaxFactory {
	return &SimpleTaxFactory{
		registry: NewStrategyRegistry(),
	}
}

func (f *SimpleTaxFactory) Create(taxType TaxType) (TaxStrategy, error) {
	return f.registry.Get(taxType)
}

func NewIncomeTaxStrategy() TaxStrategy {
	return &IncomeTaxStrategy{}
}

func NewVATStrategy() TaxStrategy {
	return &VATStrategy{}
}

func NewCorporateTaxStrategy() TaxStrategy {
	return &CorporateTaxStrategy{}
}

func NewPropertyTaxStrategy() TaxStrategy {
	return &PropertyTaxStrategy{}
}

func NewStampDutyStrategy() TaxStrategy {
	return &StampDutyStrategy{}
}

type Builder struct {
	strategy TaxStrategy
}

func NewTaxBuilder(taxType TaxType) (*Builder, error) {
	switch taxType {
	case TaxTypeIncome:
		return &Builder{strategy: NewIncomeTaxStrategy()}, nil
	case TaxTypeVAT:
		return &Builder{strategy: NewVATStrategy()}, nil
	case TaxTypeCorporate:
		return &Builder{strategy: NewCorporateTaxStrategy()}, nil
	case TaxTypeProperty:
		return &Builder{strategy: NewPropertyTaxStrategy()}, nil
	case TaxTypeStampDuty:
		return &Builder{strategy: NewStampDutyStrategy()}, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedTaxType, taxType)
	}
}

func (b *Builder) Build() TaxStrategy {
	return b.strategy
}

func (b *Builder) WithRate(rate float64) *Builder {
	return b
}

func (b *Builder) WithDeduction(deduction float64) *Builder {
	return b
}

func GetAllSupportedTaxTypes() []TaxType {
	registry := NewStrategyRegistry()
	types := make([]TaxType, 0, len(registry.strategies))
	for k := range registry.strategies {
		types = append(types, k)
	}
	return types
}

func IsSupported(taxType TaxType) bool {
	registry := NewStrategyRegistry()
	_, err := registry.Get(taxType)
	return err == nil
}