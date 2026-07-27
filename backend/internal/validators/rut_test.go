package validators

import "testing"

func TestNormalizeAndHasRUT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		normalized string
		hasRUT     bool
	}{
		{name: "empty", input: "  ", normalized: "", hasRUT: false},
		{name: "canonicalizes dots and spaces", input: "12.345.678-5", normalized: "12345678-5", hasRUT: true},
		{name: "adds dash", input: "123456785", normalized: "12345678-5", hasRUT: true},
		{name: "legacy malformed is absent", input: "legacy-value", normalized: "LEGACY-VALUE", hasRUT: false},
		{name: "wrong verifier is absent", input: "12345678-K", normalized: "12345678-K", hasRUT: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeRUT(test.input); got != test.normalized {
				t.Fatalf("NormalizeRUT(%q) = %q; want %q", test.input, got, test.normalized)
			}
			if got := HasRUT(test.input); got != test.hasRUT {
				t.Fatalf("HasRUT(%q) = %t; want %t", test.input, got, test.hasRUT)
			}
		})
	}
}

func TestIsValidRUTUsesCanonicalNormalization(t *testing.T) {
	t.Parallel()

	if !IsValidRUT("12.345.678-5") {
		t.Fatal("expected formatted valid RUT to pass")
	}
	if IsValidRUT("12.345.678-K") {
		t.Fatal("expected invalid verifier to fail")
	}
}
