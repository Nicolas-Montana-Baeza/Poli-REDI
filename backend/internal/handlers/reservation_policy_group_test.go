package handlers

import "testing"

func TestExplicitNullGroupResources(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{"omitted", `{}`, false},
		{"empty", `{"groupResources":[]}`, false},
		{"valid", `{"groupResources":[{"resourceId":1,"minimumParticipants":10}]}`, false},
		{"null", `{"groupResources":null}`, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := explicitNullGroupResources([]byte(tc.body))
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
