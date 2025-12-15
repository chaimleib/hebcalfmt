package templating

var CastFuncs = map[string]any{
	// type conversions
	"itof": Itof,
	"ftoi": Ftoi,
}

// Itof casts an int to a float64.
func Itof(i int) float64 { return float64(i) }

// Ftoi casts a float64 to an int.
func Ftoi(f float64) int { return int(f) }
