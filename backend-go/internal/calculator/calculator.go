package calculator

func CalculateTax(input TaxInput) (TaxResult, error) {
	strategy, err := GetStrategy(input.Type)
	if err != nil {
		return TaxResult{}, err
	}
	return strategy.Calculate(input)
}
