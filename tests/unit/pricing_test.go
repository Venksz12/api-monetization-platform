package unit

import (
	"testing"

	"github.com/yourusername/api-monetization-platform/internal/pricing"
)

func TestTieredPricing(t *testing.T) {
	cases := []struct {
		units int64
		want  int64
	}{
		{0, 0},
		{100_000, 0},
		{100_001, 1},
		{200_000, 100_000},
	}
	for _, tc := range cases {
		if got := pricing.Cost(tc.units, pricing.DefaultPlan); got != tc.want {
			t.Fatalf("units=%d got=%d want=%d", tc.units, got, tc.want)
		}
	}
}
