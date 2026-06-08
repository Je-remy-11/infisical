package calculator

import "fmt"

var strategyRegistry = map[TaxType]TaxStrategy{
	TaxTypeIncome:      &IncomeTaxStrategy{},
	TaxTypeCorporate:   &CorporateTaxStrategy{},
	TaxTypeValueAdded:  &ValueAddedTaxStrategy{},
	TaxTypeSales:       &SalesTaxStrategy{},
	TaxTypeProperty:    &PropertyTaxStrategy{},
	TaxTypeCapitalGain: &CapitalGainTaxStrategy{},
}

func GetStrategy(taxType TaxType) (TaxStrategy, error) {
	strategy, exists := strategyRegistry[taxType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedTaxType, taxType)
	}
	return strategy, nil
}

func RegisterStrategy(taxType TaxType, strategy TaxStrategy) {
	strategyRegistry[taxType] = strategy
}

func ListSupportedTypes() []TaxType {
	types := make([]TaxType, 0, len(strategyRegistry))
	for t := range strategyRegistry {
		types = append(types, t)
	}
	return types
}
