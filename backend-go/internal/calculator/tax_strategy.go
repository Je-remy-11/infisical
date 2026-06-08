package calculator

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

type TaxType string

const (
	TaxTypeVAT         TaxType = "vat"
	TaxTypeIncome      TaxType = "income"
	TaxTypeCorporate   TaxType = "corporate"
	TaxTypeSales       TaxType = "sales"
	TaxTypeProperty    TaxType = "property"
	TaxTypeExcise      TaxType = "excise"
	TaxTypeCustoms     TaxType = "customs"
	TaxTypePayroll     TaxType = "payroll"
	TaxTypeCapitalGain TaxType = "capital_gain"
	TaxTypeInheritance TaxType = "inheritance"
)

type TaxCalculationInput struct {
	Amount         float64
	Region         string
	TaxRate        float64
	Deductions     float64
	Exemptions     float64
	TaxBrackets    []TaxBracket
	AdditionalInfo map[string]interface{}
}

type TaxBracket struct {
	MinAmount float64
	MaxAmount float64
	Rate      float64
}

type TaxCalculationResult struct {
	TaxAmount     float64
	TaxRate       float64
	TaxableAmount float64
	Breakdown     []TaxBreakdownItem
}

type TaxBreakdownItem struct {
	Description string
	Amount      float64
}

type TaxStrategy interface {
	Calculate(input TaxCalculationInput) (TaxCalculationResult, error)
	CanHandle(taxType TaxType) bool
}
