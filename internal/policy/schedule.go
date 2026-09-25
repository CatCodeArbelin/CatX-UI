package policy

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PolicySchedule stores a weekly local-time activation window. Weekdays are
// canonical Go weekday numbers (Sunday=0 through Saturday=6), comma separated.
// The policy remains immutable when schedule state changes.
type PolicySchedule struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	PolicyID    uint   `json:"policyId" gorm:"not null;uniqueIndex"`
	Timezone    string `json:"timezone" gorm:"not null"`
	Weekdays    string `json:"weekdays" gorm:"not null"`
	StartMinute int    `json:"startMinute" gorm:"not null"`
	EndMinute   int    `json:"endMinute" gorm:"not null"`
	Enabled     bool   `json:"enabled" gorm:"not null;index"`
	CreatedAt   int64  `json:"createdAt" gorm:"autoCreateTime:milli"`
	UpdatedAt   int64  `json:"updatedAt" gorm:"autoUpdateTime:milli"`
}

func (PolicySchedule) TableName() string { return "fork_policy_schedules" }

func validateSchedule(s *PolicySchedule) error {
	if s == nil || s.PolicyID == 0 {
		return fmt.Errorf("policy is required")
	}
	if strings.TrimSpace(s.Timezone) == "" {
		return fmt.Errorf("timezone is required")
	}
	if _, err := time.LoadLocation(s.Timezone); err != nil {
		return fmt.Errorf("invalid timezone")
	}
	if s.StartMinute < 0 || s.StartMinute >= 24*60 || s.EndMinute < 0 || s.EndMinute >= 24*60 || s.StartMinute == s.EndMinute {
		return fmt.Errorf("schedule times must be different minutes in the same day")
	}
	days, err := parseWeekdays(s.Weekdays)
	if err != nil || len(days) == 0 {
		return fmt.Errorf("weekdays must contain one or more values from 0 to 6")
	}
	s.Weekdays = formatWeekdays(days)
	return nil
}

func parseWeekdays(raw string) ([]int, error) {
	seen := map[int]bool{}
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		n, err := strconv.Atoi(item)
		if err != nil || n < 0 || n > 6 {
			return nil, fmt.Errorf("invalid weekday")
		}
		seen[n] = true
	}
	days := make([]int, 0, len(seen))
	for day := range seen {
		days = append(days, day)
	}
	sort.Ints(days)
	return days, nil
}

func formatWeekdays(days []int) string {
	parts := make([]string, len(days))
	for i, day := range days {
		parts[i] = strconv.Itoa(day)
	}
	return strings.Join(parts, ",")
}

func scheduleActiveAt(s PolicySchedule, at int64) bool {
	if !s.Enabled || at <= 0 {
		return false
	}
	loc, err := time.LoadLocation(s.Timezone)
	if err != nil {
		return false
	}
	local := time.UnixMilli(at).In(loc)
	minute := local.Hour()*60 + local.Minute()
	days, err := parseWeekdays(s.Weekdays)
	if err != nil {
		return false
	}
	selected := func(day time.Weekday) bool {
		for _, candidate := range days {
			if int(day) == candidate {
				return true
			}
		}
		return false
	}
	if s.StartMinute < s.EndMinute {
		return selected(local.Weekday()) && minute >= s.StartMinute && minute < s.EndMinute
	}
	if minute >= s.StartMinute {
		return selected(local.Weekday())
	}
	previous := (int(local.Weekday()) + 6) % 7
	return minute < s.EndMinute && selected(time.Weekday(previous))
}

func scheduleNextActive(s PolicySchedule, at int64) int64 {
	if !s.Enabled || at <= 0 {
		return 0
	}
	loc, err := time.LoadLocation(s.Timezone)
	if err != nil {
		return 0
	}
	start := time.UnixMilli(at).In(loc).Truncate(time.Minute).Add(time.Minute)
	for i := 0; i < 8*24*60; i++ {
		candidate := start.Add(time.Duration(i) * time.Minute)
		if scheduleActiveAt(s, candidate.UnixMilli()) && !scheduleActiveAt(s, candidate.Add(-time.Minute).UnixMilli()) {
			return candidate.UnixMilli()
		}
	}
	return 0
}

func (r *Repository) ListSchedules(ctx context.Context) ([]PolicySchedule, error) {
	var rows []PolicySchedule
	err := r.db.WithContext(ctx).Order("policy_id ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) GetSchedule(ctx context.Context, id uint) (PolicySchedule, error) {
	var row PolicySchedule
	err := r.db.WithContext(ctx).First(&row, id).Error
	return row, err
}

func (r *Repository) CreateSchedule(ctx context.Context, s *PolicySchedule) error {
	if err := validateSchedule(s); err != nil {
		return err
	}
	if err := r.policyExists(ctx, s.PolicyID); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *Repository) UpdateSchedule(ctx context.Context, s *PolicySchedule) error {
	if err := validateSchedule(s); err != nil {
		return err
	}
	if err := r.policyExists(ctx, s.PolicyID); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&PolicySchedule{}).Where("id = ?", s.ID).Updates(map[string]any{"policy_id": s.PolicyID, "timezone": s.Timezone, "weekdays": s.Weekdays, "start_minute": s.StartMinute, "end_minute": s.EndMinute, "enabled": s.Enabled, "updated_at": time.Now().UnixMilli()}).Error
}

func (r *Repository) DeleteSchedule(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&PolicySchedule{}, id).Error
}

func (r *Repository) ScheduleState(s PolicySchedule, at int64) (bool, int64) {
	return scheduleActiveAt(s, at), scheduleNextActive(s, at)
}
