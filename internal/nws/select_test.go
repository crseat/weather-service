package nws_test

import (
	"testing"
	"time"

	"weather-service/internal/nws"
)

func TestSelectTodayPrefersNamed(t *testing.T) {
	now := time.Date(2025, 8, 13, 9, 0, 0, 0, time.UTC)
	periods := []nws.Period{
		{Name: "Tonight", IsDaytime: false, StartTime: now},
		{Name: "Today", IsDaytime: true, StartTime: now},
	}
	p, ok := nws.SelectToday(periods, now)
	if !ok || p.Name != "Today" {
		t.Fatalf("expected Today, got %+v ok=%v", p, ok)
	}
}

func TestSelectTodayDaytimeSameDate(t *testing.T) {
	now := time.Date(2025, 8, 13, 7, 0, 0, 0, time.FixedZone("X", -7*3600))
	periods := []nws.Period{
		{Name: "This Afternoon", IsDaytime: true, StartTime: time.Date(2025, 8, 13, 12, 0, 0, 0, time.FixedZone("X", -7*3600))},
	}
	p, ok := nws.SelectToday(periods, now)
	if !ok || p.Name != "This Afternoon" {
		t.Fatalf("expected This Afternoon, got %+v ok=%v", p, ok)
	}
}
