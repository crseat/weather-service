package nws

import (
	"testing"
	"time"
)

func TestSelectTodayPrefersNamed(t *testing.T) {
	now := time.Date(2025, 8, 13, 9, 0, 0, 0, time.UTC)
	periods := []Period{
		{Name: "Tonight", IsDaytime: false, StartTime: now},
		{Name: "Today", IsDaytime: true, StartTime: now},
	}
	p, ok := SelectToday(periods, now)
	if !ok || p.Name != "Today" {
		t.Fatalf("expected Today, got %+v ok=%v", p, ok)
	}
}

func TestSelectTodayDaytimeSameDate(t *testing.T) {
	now := time.Date(2025, 8, 13, 7, 0, 0, 0, time.FixedZone("X", -7*3600))
	periods := []Period{
		{Name: "This Afternoon", IsDaytime: true, StartTime: time.Date(2025, 8, 13, 12, 0, 0, 0, time.FixedZone("X", -7*3600))},
	}
	p, ok := SelectToday(periods, now)
	if !ok || p.Name != "This Afternoon" {
		t.Fatalf("expected This Afternoon, got %+v ok=%v", p, ok)
	}
}
