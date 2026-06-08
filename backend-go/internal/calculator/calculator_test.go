package calculator

import (
	"context"
	"testing"
)

func ptr[T any](v T) *T { return &v }

func TestCalculateTax_Income(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		input      *TaxInput
		wantTaxMin float64
		wantTaxMax float64
		wantErr    bool
	}{
		{
			name: "below threshold - no tax",
			input: &TaxInput{
				TaxType:    TaxTypeIncome,
				Amount:     50000,
				Deductions: 0,
			},
			wantTaxMin: 0,
			wantTaxMax: 0,
		},
		{
			name: "first bracket - 3%",
			input: &TaxInput{
				TaxType:    TaxTypeIncome,
				Amount:     100000,
				Deductions: 0,
			},
			wantTaxMin: 1000,
			wantTaxMax: 2000,
		},
		{
			name: "high income - higher bracket",
			input: &TaxInput{
				TaxType:    TaxTypeIncome,
				Amount:     500000,
				Deductions: 0,
			},
			wantTaxMin: 50000,
			wantTaxMax: 150000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateTax(ctx, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.TaxAmount < tt.wantTaxMin || result.TaxAmount > tt.wantTaxMax {
				t.Errorf("tax amount %.2f not in range [%.2f, %.2f]",
					result.TaxAmount, tt.wantTaxMin, tt.wantTaxMax)
			}
		})
	}
}

func TestCalculateTax_VAT(t *testing.T) {
	ctx := context.Background()

	result, err := CalculateTax(ctx, &TaxInput{
		TaxType:    TaxTypeVAT,
		Amount:     100000,
		Deductions: 30000,
		Rate:       0.13,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := (100000 - 30000) * 0.13
	if result.TaxAmount != expected {
		t.Errorf("expected VAT %.2f, got %.2f", expected, result.TaxAmount)
	}
}

func TestCalculateTax_Corporate(t *testing.T) {
	ctx := context.Background()

	t.Run("standard rate", func(t *testing.T) {
		result, err := CalculateTax(ctx, &TaxInput{
			TaxType:    TaxTypeCorporate,
			Amount:     1000000,
			Deductions: 200000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := (1000000 - 200000) * 0.25
		if result.TaxAmount != expected {
			t.Errorf("expected %.2f, got %.2f", expected, result.TaxAmount)
		}
	})

	t.Run("high-tech rate", func(t *testing.T) {
		result, err := CalculateTax(ctx, &TaxInput{
			TaxType:    TaxTypeCorporate,
			Amount:     1000000,
			Deductions: 200000,
			IsHighTech: ptr(true),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := (1000000 - 200000) * 0.15
		if result.TaxAmount != expected {
			t.Errorf("expected %.2f, got %.2f", expected, result.TaxAmount)
		}
	})

	t.Run("loss no tax", func(t *testing.T) {
		result, err := CalculateTax(ctx, &TaxInput{
			TaxType:    TaxTypeCorporate,
			Amount:     100000,
			Deductions: 200000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.TaxAmount != 0 {
			t.Errorf("expected 0 tax for loss, got %.2f", result.TaxAmount)
		}
	})
}

func TestCalculateTax_Property(t *testing.T) {
	ctx := context.Background()

	result, err := CalculateTax(ctx, &TaxInput{
		TaxType:    TaxTypeProperty,
		Amount:     5000000,
		PropertyID: ptr("PROP-001"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := 5000000 * 0.7 * 0.012
	if result.TaxAmount != expected {
		t.Errorf("expected %.2f, got %.2f", expected, result.TaxAmount)
	}
}

func TestCalculateTax_StampDuty(t *testing.T) {
	ctx := context.Background()

	t.Run("purchase contract", func(t *testing.T) {
		result, err := CalculateTax(ctx, &TaxInput{
			TaxType: TaxTypeStampDuty,
			Amount:  1000000,
			DocType: ptr("purchase"),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := 1000000 * 0.0005
		if result.TaxAmount != expected {
			t.Errorf("expected %.2f, got %.2f", expected, result.TaxAmount)
		}
	})

	t.Run("unknown doc type", func(t *testing.T) {
		_, err := CalculateTax(ctx, &TaxInput{
			TaxType: TaxTypeStampDuty,
			Amount:  1000000,
			DocType: ptr("unknown_type"),
		})
		if err == nil {
			t.Error("expected error for unknown doc type")
		}
	})
}

func TestCalculateTax_InvalidInput(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		input *TaxInput
	}{
		{"nil input", nil},
		{"negative amount", &TaxInput{TaxType: TaxTypeIncome, Amount: -100}},
		{"negative deductions", &TaxInput{TaxType: TaxTypeIncome, Amount: 100, Deductions: -50}},
		{"empty tax type", &TaxInput{Amount: 100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateTax(ctx, tt.input)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestCalculateTax_UnsupportedType(t *testing.T) {
	ctx := context.Background()

	_, err := CalculateTax(ctx, &TaxInput{
		TaxType: TaxType("unknown_tax"),
		Amount:  100,
	})
	if err == nil {
		t.Error("expected error for unsupported tax type")
	}
}

func TestStrategyRegistry(t *testing.T) {
	r := NewStrategyRegistry()

	s, err := r.Get(TaxTypeIncome)
	if err != nil {
		t.Fatalf("expected income strategy to be registered: %v", err)
	}
	if s.TaxType() != TaxTypeIncome {
		t.Errorf("expected TaxTypeIncome, got %s", s.TaxType())
	}
}

func TestStrategy_Validate(t *testing.T) {
	tests := []struct {
		name     string
		strategy TaxStrategy
		input    *TaxInput
		wantErr  bool
	}{
		{
			name:     "income tax valid",
			strategy: NewIncomeTaxStrategy(),
			input:    &TaxInput{Amount: 100000},
		},
		{
			name:     "vat missing rate",
			strategy: NewVATStrategy(),
			input:    &TaxInput{Amount: 100000},
			wantErr:  true,
		},
		{
			name:     "vat valid",
			strategy: NewVATStrategy(),
			input:    &TaxInput{Amount: 100000, Rate: 0.13},
		},
		{
			name:     "stamp duty missing doc type",
			strategy: NewStampDutyStrategy(),
			input:    &TaxInput{Amount: 100000},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.strategy.Validate(tt.input)
			if tt.wantErr && err == nil {
				t.Error("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestBuilder(t *testing.T) {
	b, err := NewTaxBuilder(TaxTypeCorporate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := b.Build()
	if s.TaxType() != TaxTypeCorporate {
		t.Errorf("expected TaxTypeCorporate, got %s", s.TaxType())
	}
}

func TestGetAllSupportedTaxTypes(t *testing.T) {
	types := GetAllSupportedTaxTypes()
	if len(types) != 5 {
		t.Errorf("expected 5 supported tax types, got %d", len(types))
	}
}

func TestIsSupported(t *testing.T) {
	if !IsSupported(TaxTypeIncome) {
		t.Error("income tax should be supported")
	}
	if IsSupported(TaxType("unknown")) {
		t.Error("unknown tax type should not be supported")
	}
}