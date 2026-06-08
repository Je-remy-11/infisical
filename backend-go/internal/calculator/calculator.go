package calculator

import (
	"fmt"
	"math"
)

type TaxType string

const (
	TaxTypeVAT     TaxType = "vat"
	TaxTypeSales   TaxType = "sales"
	TaxTypeService TaxType = "service"
	TaxTypeImport  TaxType = "import"
	TaxTypeLuxury  TaxType = "luxury"
)

type CustomerType string

const (
	CustomerTypeIndividual CustomerType = "individual"
	CustomerTypeBusiness   CustomerType = "business"
	CustomerTypeNonProfit  CustomerType = "non_profit"
)

type TaxRequest struct {
	TaxType        TaxType
	Amount         float64
	Quantity       int
	TaxRate        float64
	FixedFee       float64
	Threshold      float64
	AdditionalRate float64
	Region         string
	CustomerType   CustomerType
	IsImported     bool
	IsEssential    bool
	TaxExempt      bool
}

type TaxResult struct {
	Subtotal      float64
	TaxableAmount float64
	TaxAmount     float64
	Total         float64
}

type TaxStrategy interface {
	Calculate(TaxRequest) (TaxResult, error)
}

type VATStrategy struct{}

type SalesTaxStrategy struct{}

type ServiceTaxStrategy struct{}

type ImportTaxStrategy struct{}

type LuxuryTaxStrategy struct{}

func CalculateTax(req TaxRequest) (TaxResult, error) {
	if err := validateRequest(req); err != nil {
		return TaxResult{}, err
	}

	strategy, err := NewTaxStrategy(req.TaxType)
	if err != nil {
		return TaxResult{}, err
	}

	return strategy.Calculate(req)
}

func NewTaxStrategy(taxType TaxType) (TaxStrategy, error) {
	switch taxType {
	case TaxTypeVAT:
		return VATStrategy{}, nil
	case TaxTypeSales:
		return SalesTaxStrategy{}, nil
	case TaxTypeService:
		return ServiceTaxStrategy{}, nil
	case TaxTypeImport:
		return ImportTaxStrategy{}, nil
	case TaxTypeLuxury:
		return LuxuryTaxStrategy{}, nil
	default:
		return nil, fmt.Errorf("unsupported tax type: %s", taxType)
	}
}

func (VATStrategy) Calculate(req TaxRequest) (TaxResult, error) {
	subtotal := calculateSubtotal(req)
	if req.TaxExempt {
		return newTaxResult(subtotal, 0, 0), nil
	}

	taxableAmount := subtotal
	rate := req.TaxRate
	if req.IsEssential {
		rate = rate / 2
	}
	if req.CustomerType == CustomerTypeBusiness && req.Region == "export" {
		rate = 0
	}

	taxAmount := taxableAmount * rate
	return newTaxResult(subtotal, taxableAmount, taxAmount), nil
}

func (SalesTaxStrategy) Calculate(req TaxRequest) (TaxResult, error) {
	subtotal := calculateSubtotal(req)
	if req.TaxExempt || req.Region == "tax_free" {
		return newTaxResult(subtotal, 0, 0), nil
	}

	taxableAmount := subtotal - req.Threshold
	if taxableAmount < 0 {
		taxableAmount = 0
	}
	if req.IsEssential {
		taxableAmount *= 0.5
	}

	taxAmount := taxableAmount * req.TaxRate
	return newTaxResult(subtotal, taxableAmount, taxAmount), nil
}

func (ServiceTaxStrategy) Calculate(req TaxRequest) (TaxResult, error) {
	subtotal := calculateSubtotal(req)
	if req.TaxExempt || req.CustomerType == CustomerTypeNonProfit {
		return newTaxResult(subtotal, 0, 0), nil
	}

	taxableAmount := subtotal
	rate := req.TaxRate
	if req.CustomerType == CustomerTypeBusiness {
		rate *= 0.9
	}

	taxAmount := (taxableAmount * rate) + req.FixedFee
	return newTaxResult(subtotal, taxableAmount, taxAmount), nil
}

func (ImportTaxStrategy) Calculate(req TaxRequest) (TaxResult, error) {
	subtotal := calculateSubtotal(req)
	if req.TaxExempt {
		return newTaxResult(subtotal, 0, 0), nil
	}

	taxableAmount := subtotal
	if req.IsImported {
		taxableAmount += req.FixedFee
	}

	taxAmount := taxableAmount * req.TaxRate
	if req.IsImported {
		taxAmount += subtotal * req.AdditionalRate
	}

	return newTaxResult(subtotal, taxableAmount, taxAmount), nil
}

func (LuxuryTaxStrategy) Calculate(req TaxRequest) (TaxResult, error) {
	subtotal := calculateSubtotal(req)
	if req.TaxExempt || subtotal <= req.Threshold {
		return newTaxResult(subtotal, 0, 0), nil
	}

	taxableAmount := subtotal - req.Threshold
	taxAmount := taxableAmount * req.TaxRate
	if req.IsImported {
		taxAmount += subtotal * req.AdditionalRate
	}

	return newTaxResult(subtotal, taxableAmount, taxAmount), nil
}

func validateRequest(req TaxRequest) error {
	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	if req.Amount < 0 {
		return fmt.Errorf("amount must be non-negative")
	}
	if err := validateRate("tax rate", req.TaxRate); err != nil {
		return err
	}
	if err := validateRate("additional rate", req.AdditionalRate); err != nil {
		return err
	}
	if req.FixedFee < 0 {
		return fmt.Errorf("fixed fee must be non-negative")
	}
	if req.Threshold < 0 {
		return fmt.Errorf("threshold must be non-negative")
	}

	return nil
}

func validateRate(name string, value float64) error {
	if value < 0 || value > 1 {
		return fmt.Errorf("%s must be between 0 and 1", name)
	}
	return nil
}

func calculateSubtotal(req TaxRequest) float64 {
	return roundCurrency(req.Amount * float64(req.Quantity))
}

func newTaxResult(subtotal, taxableAmount, taxAmount float64) TaxResult {
	subtotal = roundCurrency(subtotal)
	taxableAmount = roundCurrency(taxableAmount)
	taxAmount = roundCurrency(taxAmount)

	return TaxResult{
		Subtotal:      subtotal,
		TaxableAmount: taxableAmount,
		TaxAmount:     taxAmount,
		Total:         roundCurrency(subtotal + taxAmount),
	}
}

func roundCurrency(value float64) float64 {
	return math.Round(value*100) / 100
}
