package calculator

type TaxType string

const (
	TaxTypeIncome      TaxType = "income"
	TaxTypeCorporate   TaxType = "corporate"
	TaxTypeValueAdded  TaxType = "value_added"
	TaxTypeSales       TaxType = "sales"
	TaxTypeProperty    TaxType = "property"
	TaxTypeCapitalGain TaxType = "capital_gain"
)

type TaxInput struct {
	Type       TaxType
	Amount     float64
	Region     string
	IsExempt   bool
	Deductions float64
	Year       int
}

type TaxResult struct {
	TaxAmount  float64
	TaxRate    float64
	TaxType    TaxType
	Breakdown  map[string]float64
}

type TaxStrategy interface {
	Calculate(input TaxInput) (TaxResult, error)
	SupportedType() TaxType
}
