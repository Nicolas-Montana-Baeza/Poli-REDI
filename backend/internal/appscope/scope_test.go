package appscope

import "testing"

func TestCurrentDefaultsToMVP1(t *testing.T) {
	t.Setenv("MVP_SCOPE", "")
	if got := Current(); got != MVP1 {
		t.Fatalf("Current() = %q, want %q", got, MVP1)
	}
}

func TestCurrentRequiresExplicitFullScope(t *testing.T) {
	t.Setenv("MVP_SCOPE", "FULL")
	if got := Current(); got != Full {
		t.Fatalf("Current() = %q, want %q", got, Full)
	}

	t.Setenv("MVP_SCOPE", "mvp2")
	if got := Current(); got != MVP1 {
		t.Fatalf("unknown scope must fail closed; got %q", got)
	}
}
