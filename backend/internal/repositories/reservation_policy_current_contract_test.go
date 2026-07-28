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
	normalized := strings.ToUpper(string(source))
	start := strings.Index(normalized, "FUNC GETCURRENTRESERVATIONPOLICYCOMPLETE")
	if start < 0 {
		t.Fatal("current policy repository function not found")
	}
	currentPolicySource := normalized[start:]
	if !strings.Contains(currentPolicySource, "IS_PUBLISHED = 1") {
		t.Fatal("current policy endpoint can expose an unpublished policy")
	}
}
