package pricing

type Tier struct {
	UpTo int64
	Rate int64 // minor currency units per request
}

type Plan struct {
	ID    string
	Tiers []Tier
}

var Default = Plan{
	ID: "growth",
	Tiers: []Tier{
		{UpTo: 100_000, Rate: 0},
		{UpTo: 1_000_000, Rate: 1},
		{UpTo: 0, Rate: 0}, // 0 means unbounded
	},
}

func Cost(units int64, plan Plan) int64 {
	var cost int64
	var consumed int64
	for _, tier := range plan.Tiers {
		if units <= consumed {
			break
		}
		var span int64
		if tier.UpTo == 0 {
			span = units - consumed
		} else {
			span = tier.UpTo - consumed
			if span < 0 {
				span = 0
			}
			if span > units-consumed {
				span = units - consumed
			}
		}
		cost += span * tier.Rate
		consumed += span
	}
	return cost
}
