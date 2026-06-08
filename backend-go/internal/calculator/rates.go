package calculator

type TaxBracket struct {
	Min    float64
	Max    float64
	Rate   float64
	Label  string
}

func getIncomeBrackets(region string) []TaxBracket {
	switch region {
	case "US":
		return []TaxBracket{
			{Min: 0, Max: 11000, Rate: 0.10, Label: "10%"},
			{Min: 11000, Max: 44725, Rate: 0.12, Label: "12%"},
			{Min: 44725, Max: 95375, Rate: 0.22, Label: "22%"},
			{Min: 95375, Max: 182100, Rate: 0.24, Label: "24%"},
			{Min: 182100, Max: 231250, Rate: 0.32, Label: "32%"},
			{Min: 231250, Max: 578125, Rate: 0.35, Label: "35%"},
			{Min: 578125, Max: 0, Rate: 0.37, Label: "37%"},
		}
	case "CN":
		return []TaxBracket{
			{Min: 0, Max: 36000, Rate: 0.03, Label: "3%"},
			{Min: 36000, Max: 144000, Rate: 0.10, Label: "10%"},
			{Min: 144000, Max: 300000, Rate: 0.20, Label: "20%"},
			{Min: 300000, Max: 420000, Rate: 0.25, Label: "25%"},
			{Min: 420000, Max: 660000, Rate: 0.30, Label: "30%"},
			{Min: 660000, Max: 960000, Rate: 0.35, Label: "35%"},
			{Min: 960000, Max: 0, Rate: 0.45, Label: "45%"},
		}
	default:
		return []TaxBracket{
			{Min: 0, Max: 50000, Rate: 0.20, Label: "20%"},
			{Min: 50000, Max: 0, Rate: 0.40, Label: "40%"},
		}
	}
}

func getCorporateRate(region string) float64 {
	switch region {
	case "US":
		return 0.21
	case "CN":
		return 0.25
	case "IE":
		return 0.125
	default:
		return 0.25
	}
}

func getVATRate(region string) float64 {
	switch region {
	case "CN":
		return 0.13
	case "DE":
		return 0.19
	case "UK":
		return 0.20
	default:
		return 0.10
	}
}

func getSalesTaxRate(region string) float64 {
	switch region {
	case "US-CA":
		return 0.0725
	case "US-NY":
		return 0.08
	case "US-TX":
		return 0.0625
	default:
		return 0.05
	}
}

func getPropertyTaxRate(region string) float64 {
	switch region {
	case "US-NJ":
		return 0.0249
	case "US-IL":
		return 0.0227
	case "US-TX":
		return 0.0180
	default:
		return 0.01
	}
}

func getCapitalGainRate(year int, region string) float64 {
	switch region {
	case "US":
		if year >= 2024 {
			return 0.20
		}
		return 0.15
	case "CN":
		return 0.20
	default:
		return 0.15
	}
}
