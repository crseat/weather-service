package forecast

// Bands describe temperature thresholds in Fahrenheit for classification.
// If temp <= ColdMax => "cold"; if temp >= HotMin => "hot"; otherwise "moderate".
type Bands struct {
	ColdMax int // <= ColdMax => "cold"
	HotMin  int // >= HotMin  => "hot"
}

// Classify classifies a Fahrenheit temperature into hot/moderate/cold using Bands.
func Classify(temp int, b Bands) string {
	if temp >= b.HotMin {
		return "hot"
	}
	if temp <= b.ColdMax {
		return "cold"
	}
	return "moderate"
}
