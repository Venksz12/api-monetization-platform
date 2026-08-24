package pricing

import "testing"

func TestCost(t *testing.T) {
	if got := Cost(100_000, Default); got != 0 {
		t.Fatalf("got %d", got)
	}
	if got := Cost(100_001, Default); got != 1 {
		t.Fatalf("got %d", got)
	}
}
