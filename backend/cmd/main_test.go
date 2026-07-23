package main

import "testing"

func TestParseJoinCodeKeyVersion(t *testing.T) {
	if got, err := parseJoinCodeKeyVersion(" 1 "); err != nil || got != 1 {
		t.Fatalf("valid version = %d, %v", got, err)
	}
	for _, invalid := range []string{"", "0", "-1", `"1"`, "1:key", "one"} {
		if _, err := parseJoinCodeKeyVersion(invalid); err == nil {
			t.Fatalf("accepted invalid version %q", invalid)
		}
	}
}
