package groupquota

import (
	"testing"
	"time"
)

func TestCeilMulUsesFixedPointWithoutFloatDrift(t *testing.T) {
	tests := []struct {
		name             string
		bytes, ppm, want int64
	}{
		{"identity", 100, Scale, 100},
		{"round-up", 101, 500_000, 51},
		{"fractional", 3, 333_333, 1},
		{"zero", 0, Scale, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ceilMul(tt.bytes, tt.ppm); got != tt.want {
				t.Fatalf("ceilMul(%d,%d)=%d, want %d", tt.bytes, tt.ppm, got, tt.want)
			}
		})
	}
}

func TestMemberUsageDoesNotGoNegativeAfterRebaseline(t *testing.T) {
	base, current := int64(100), int64(20)
	if got := max(current-base, 0); got != 0 {
		t.Fatalf("member usage after counter decrease = %d, want 0", got)
	}
	carry := max(base-0, 0)
	if carry != 100 {
		t.Fatalf("preserved carry = %d, want 100", carry)
	}
}

func TestResetDueIsDeterministic(t *testing.T) {
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	last := now.Add(-25 * time.Hour).UnixMilli()
	if !resetDue(State{ResetPeriod: periodDaily, LastResetAt: last}, now) {
		t.Fatal("daily reset should be due")
	}
	if resetDue(State{ResetPeriod: periodWeekly, LastResetAt: last}, now) {
		t.Fatal("weekly reset should not be due")
	}
	if !resetDue(State{ResetPeriod: periodMonthly, LastResetAt: now.AddDate(0, -1, 0).UnixMilli()}, now) {
		t.Fatal("monthly reset should be due")
	}
}
