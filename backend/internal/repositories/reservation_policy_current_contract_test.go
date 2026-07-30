package repositories

import (
	"os"
	"strings"
	"testing"
)

func TestCurrentPolicyUsesOnlyPublishedPolicies(t *testing.T) {
	source, err := os.ReadFile("reservation_policies_repository.go")
	if err != nil {
		t.Fatal(err)
	}
	currentPolicySource := repositoryFunctionSource(
		t,
		string(source),
		"GetCurrentReservationPolicyComplete",
	)
	if !strings.Contains(currentPolicySource, "IS_PUBLISHED = 1") {
		t.Fatal("current policy endpoint can expose an unpublished policy")
	}
}

func TestCurrentPolicyPublishedContractDoesNotMatchLaterFunctions(t *testing.T) {
	source := `
func GetCurrentReservationPolicyComplete() {
	query("SELECT * FROM policies")
}

func GetReservationPolicyHistory() {
	query("SELECT * FROM policies WHERE is_published = 1")
}`
	current := repositoryFunctionSource(t, source, "GetCurrentReservationPolicyComplete")
	if strings.Contains(current, "IS_PUBLISHED = 1") {
		t.Fatal("function extraction included a later repository function")
	}
}

func repositoryFunctionSource(t *testing.T, source, name string) string {
	t.Helper()
	normalized := strings.ToUpper(strings.ReplaceAll(source, "\r\n", "\n"))
	signature := "FUNC " + strings.ToUpper(name)
	start := strings.Index(normalized, signature)
	if start < 0 {
		t.Fatalf("repository function %s not found", name)
	}
	current := normalized[start:]
	if next := strings.Index(current[len(signature):], "\nFUNC "); next >= 0 {
		current = current[:len(signature)+next]
	}
	return current
}
