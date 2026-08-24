package pricing

type Tier struct {
	UpToExclusive int64
	RateMinor     int64
}

type Plan struct {
	ID    string
	Tiers []Tier
}

var DefaultPlan = Plan{
	ID: "growth-v1",
	Tiers: []Tier{
		{UpToExclusive: 100_000, RateMinor: 0},
		{UpToExclusive: 1_000_000, RateMinor: 1},
		{UpToExclusive: 0, RateMinor: 0}, // unbounded; production plan may use 0.5 minor via a finer money unit
	},
}

func Cost(units int64, plan Plan) int64 {
	if units <= 0 {
		return 0
	}
	var total, previous int64
	for _, tier := range plan.Tiers {
		var upper int64
		if tier.UpToExclusive == 0 {
			upper = units
		} else {
			upper = tier.UpToExclusive
		}
		if upper <= previous {
			continue
		}
		span := units - previous
		if span > upper-previous {
			span = upper - previous
		}
		if span > 0 {
			total += span * tier.RateMinor
			previous += span
		}
		if previous >= units {
			break
		}
	}
	return total
}
