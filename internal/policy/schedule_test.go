package policy

import (
	"testing"
	"time"
)

func atIn(zone, value string) int64 {
	location, err := time.LoadLocation(zone)
	if err != nil {
		panic(err)
	}
	return time.Date(2026, 1, 5, 0, 0, 0, 0, location).Add(parseClock(value)).UnixMilli()
}

func parseClock(value string) time.Duration {
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		panic(err)
	}
	return time.Duration(parsed.Hour())*time.Hour + time.Duration(parsed.Minute())*time.Minute
}

func TestScheduleValidation(t *testing.T) {
	s := PolicySchedule{PolicyID: 1, Timezone: "Europe/Moscow", Weekdays: "1,3,1", StartMinute: 9 * 60, EndMinute: 17 * 60, Enabled: true}
	if err := validateSchedule(&s); err != nil {
		t.Fatal(err)
	}
	if s.Weekdays != "1,3" {
		t.Fatalf("canonical weekdays = %q", s.Weekdays)
	}
	for _, invalid := range []PolicySchedule{
		{PolicyID: 1, Timezone: "Not/IANA", Weekdays: "1", StartMinute: 1, EndMinute: 2},
		{PolicyID: 1, Timezone: "UTC", Weekdays: "1", StartMinute: 2, EndMinute: 2},
		{PolicyID: 1, Timezone: "UTC", Weekdays: "7", StartMinute: 1, EndMinute: 2},
	} {
		if err := validateSchedule(&invalid); err == nil {
			t.Fatalf("invalid schedule accepted: %+v", invalid)
		}
	}
}

func TestScheduleDailyWindow(t *testing.T) {
	s := PolicySchedule{Timezone: "UTC", Weekdays: "1", StartMinute: 9 * 60, EndMinute: 17 * 60, Enabled: true}
	for clock, want := range map[string]bool{"08:59": false, "09:00": true, "16:59": true, "17:00": false} {
		if got := scheduleActiveAt(s, atIn("UTC", clock)); got != want {
			t.Errorf("%s active = %v, want %v", clock, got, want)
		}
	}
}

func TestScheduleCrossMidnightAndWeekdayTransition(t *testing.T) {
	s := PolicySchedule{Timezone: "UTC", Weekdays: "1", StartMinute: 23 * 60, EndMinute: 2 * 60, Enabled: true}
	checks := map[string]bool{
		"2026-01-05T22:59:00Z": false,
		"2026-01-05T23:00:00Z": true,
		"2026-01-06T01:59:00Z": true,
		"2026-01-06T02:00:00Z": false,
	}
	for raw, want := range checks {
		instant, _ := time.Parse(time.RFC3339, raw)
		if got := scheduleActiveAt(s, instant.UnixMilli()); got != want {
			t.Errorf("%s active = %v, want %v", raw, got, want)
		}
	}
}

func TestScheduleDSTUsesNamedTimezone(t *testing.T) {
	s := PolicySchedule{Timezone: "America/New_York", Weekdays: "0", StartMinute: 1 * 60, EndMinute: 4 * 60, Enabled: true}
	// The spring-forward gap is absent, while both occurrences of the fall
	// repeated hour are real instants and remain inside the local window.
	spring, _ := time.Parse(time.RFC3339, "2026-03-08T07:30:00Z")
	if !scheduleActiveAt(s, spring.UnixMilli()) {
		t.Fatal("spring-forward local 03:30 should be active")
	}
	fallA, _ := time.Parse(time.RFC3339, "2026-11-01T05:30:00Z")
	fallB, _ := time.Parse(time.RFC3339, "2026-11-01T06:30:00Z")
	if !scheduleActiveAt(s, fallA.UnixMilli()) || !scheduleActiveAt(s, fallB.UnixMilli()) {
		t.Fatal("both fall-back 01:30 instants should be active")
	}
}

func TestScheduleDisabledAndNextActivation(t *testing.T) {
	s := PolicySchedule{Timezone: "UTC", Weekdays: "1", StartMinute: 9 * 60, EndMinute: 10 * 60, Enabled: true}
	now := time.Date(2026, 1, 5, 8, 30, 0, 0, time.UTC).UnixMilli()
	if got := scheduleNextActive(s, now); got != time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC).UnixMilli() {
		t.Fatalf("next activation = %d", got)
	}
	s.Enabled = false
	if scheduleActiveAt(s, now) || scheduleNextActive(s, now) != 0 {
		t.Fatal("disabled schedule must be inactive")
	}
}
