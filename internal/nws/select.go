package nws

import "time"

// SelectToday chooses the best "today" period from a list of periods.
// Preference order:
//  1. Exact name match: If any period has Name == "Today", return it immediately.
//  2. Same local calendar date and daytime: Otherwise, scan daytime periods (IsDaytime == true) and pick the first
//     whose StartTime has the same local date as “now” interpreted in that period’s time zone/offset.
//  3. Fallback: If neither 1 nor 2 yields a result, return the first period in the list.
func SelectToday(periods []Period, now time.Time) (Period, bool) {
	if len(periods) == 0 {
		return Period{}, false
	}
	for _, p := range periods {
		if p.Name == "Today" {
			return p, true
		}
	}
	for _, p := range periods {
		if !p.IsDaytime {
			continue
		}
		locNow := now.In(p.StartTime.Location())
		if sameDate(p.StartTime, locNow) {
			return p, true
		}
	}
	return periods[0], true
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
