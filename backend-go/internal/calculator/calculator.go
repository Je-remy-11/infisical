package calculator

type Calculator struct {
	factory *TaxStrategyFactory
}

func NewCalculator() *Calculator {
	return &Calculator{
		factory: NewTaxStrategyFactory(),
	}
}

func (c *Calculator) CalculateTax(taxType TaxType, input TaxCalculationInput) (TaxCalculationResult, error) {
	strategy, err := c.factory.GetStrategy(taxType)
	if err != nil {
		return TaxCalculationResult{}, err
	}
	return strategy.Calculate(input)
}
