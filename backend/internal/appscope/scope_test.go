package appscope

import "testing"

func TestCurrentDefaultsToMVP1(t *testing.T) {
	t.Setenv("MVP_SCOPE", "")

	if got := Current(); got != MVP1 {
		t.Fatalf(
			"Current() = %q, want %q",
			got,
			MVP1,
		)
	}
}

func TestCurrentRecognizesMVP2(t *testing.T) {
	t.Setenv("MVP_SCOPE", "MVP2")

	if got := Current(); got != MVP2 {
		t.Fatalf(
			"Current() = %q, want %q",
			got,
			MVP2,
		)
	}

	if !HasMVP2() {
		t.Fatal("MVP2 must enable MVP2 features")
	}

	if IsFull() {
		t.Fatal("MVP2 must not enable full legacy scope")
	}
}

func TestCurrentRecognizesFull(t *testing.T) {
	t.Setenv("MVP_SCOPE", "FULL")

	if got := Current(); got != Full {
		t.Fatalf(
			"Current() = %q, want %q",
			got,
			Full,
		)
	}

	if !HasMVP2() {
		t.Fatal("full scope must include MVP2 features")
	}

	if !IsFull() {
		t.Fatal("expected full scope")
	}
}

func TestUnknownScopeFailsClosedToMVP1(t *testing.T) {
	t.Setenv("MVP_SCOPE", "experimental")

	if got := Current(); got != MVP1 {
		t.Fatalf(
			"unknown scope must fail closed; got %q",
			got,
		)
	}
}
