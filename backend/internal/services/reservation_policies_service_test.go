package services

import "testing"

func TestUniquePositiveSortsAndDeduplicates(t *testing.T) {
	values := uniquePositive([]int{90, 30, 90, 60})
	want := []int{30, 60, 90}
	if len(values) != len(want) {
		t.Fatalf("uniquePositive() = %v", values)
	}
	for index := range want {
		if values[index] != want[index] {
			t.Fatalf("uniquePositive() = %v", values)
		}
	}
}
