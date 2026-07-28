package repositories

import (
	"os"
	"strings"
	"testing"
)

func TestFreshDatabaseBootstrapAssignsAllThreeCourtsToGroupPolicy(t *testing.T) {
	seed, err := os.ReadFile("../../../database/seed.sql")
	if err != nil {
		t.Fatal(err)
	}
	source := string(seed)
	for _, expected := range []string{
		"INSERT INTO dbo.reservation_policy_group_resources",
		"WHERE r.id IN (1, 2, 7)",
	} {
		if !strings.Contains(source, expected) {
			t.Fatalf("seed bootstrap missing %q", expected)
		}
	}
}
