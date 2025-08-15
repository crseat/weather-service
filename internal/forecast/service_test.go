package forecast

import "testing"

func TestClassify(t *testing.T) {
	b := Bands{ColdMax: 45, HotMin: 85}
	cases := []struct {
		t int
		e string
	}{
		{30, "cold"},
		{45, "cold"},
		{46, "moderate"},
		{84, "moderate"},
		{85, "hot"},
		{100, "hot"},
	}
	for _, c := range cases {
		if got := Classify(c.t, b); got != c.e {
			t.Fatalf("temp %d => %s, want %s", c.t, got, c.e)
		}
	}
}
