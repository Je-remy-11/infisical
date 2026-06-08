package calculator

import "errors"

var (
	ErrUnsupportedTaxType = errors.New("unsupported tax type")
)

type TaxStrategyFactory struct {
	strategies []TaxStrategy
}

func NewTaxStrategyFactory() *TaxStrategyFactory {
	factory := &TaxStrategyFactory{
		strategies: []TaxStrategy{
			NewVATStrategy(),
			NewIncomeStrategy(),
			NewCorporateStrategy(),
			NewSalesStrategy(),
			NewPropertyStrategy(),
			NewExciseStrategy(),
			NewCustomsStrategy(),
			NewPayrollStrategy(),
			NewCapitalGainStrategy(),
			NewInheritanceStrategy(),
		},
	}
	return factory
}

func (f *TaxStrategyFactory) GetStrategy(taxType TaxType) (TaxStrategy, error) {
	for _, strategy := range f.strategies {
		if strategy.CanHandle(taxType) {
			return strategy, nil
		}
	}
	return nil, ErrUnsupportedTaxType
}

func (f *TaxStrategyFactory) RegisterStrategy(strategy TaxStrategy) {
	f.strategies = append(f.strategies, strategy)
}
