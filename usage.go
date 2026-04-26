package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type usagePeriod string

const (
	usageDaily    usagePeriod = "daily"
	usageWeekly   usagePeriod = "weekly"
	usageMonthly  usagePeriod = "monthly"
	gasRateWindow             = time.Hour
	gasRateMinAge             = 5 * time.Minute
)

type usageStore struct {
	path string
	mu   sync.Mutex
	data usageHistory
	now  func() time.Time
}

type usageHistory struct {
	Samples []usageSample `json:"samples"`
}

type usageSample struct {
	ObservedAt time.Time          `json:"observed_at"`
	Totals     map[string]float64 `json:"totals"`
}

func newUsageStore(dir string) (*usageStore, error) {
	store := &usageStore{
		path: filepath.Join(dir, "usage-history.json"),
		now:  time.Now,
	}

	raw, err := os.ReadFile(store.path)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, fmt.Errorf("read usage history: %w", err)
	}

	if len(raw) == 0 {
		return store, nil
	}

	if err := json.Unmarshal(raw, &store.data); err != nil {
		return nil, fmt.Errorf("parse usage history: %w", err)
	}

	sort.Slice(store.data.Samples, func(i, j int) bool {
		return store.data.Samples[i].ObservedAt.Before(store.data.Samples[j].ObservedAt)
	})

	return store, nil
}

func (s *usageStore) Record(snapshot *snapshot) error {
	totals := snapshotUsageTotals(snapshot)
	if len(totals) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now().UTC()
	if len(s.data.Samples) > 0 {
		last := s.data.Samples[len(s.data.Samples)-1]
		if sameTotals(last.Totals, totals) && now.Sub(last.ObservedAt) < time.Minute {
			return nil
		}
	}

	s.data.Samples = append(s.data.Samples, usageSample{
		ObservedAt: now,
		Totals:     totals,
	})
	s.prune(now)

	if err := s.save(); err != nil {
		return err
	}

	return nil
}

func (s *usageStore) Usage(snapshot *snapshot, metric string, period usagePeriod) (float64, error) {
	currentTotals := snapshotUsageTotals(snapshot)
	current, ok := currentTotals[metric]
	if !ok {
		return 0, fmt.Errorf("%w: %s total", errValueUnavailable, metric)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	start := periodStart(s.now(), period)
	baseline, ok := s.baseline(metric, start)
	if !ok {
		return 0, fmt.Errorf("%w: %s %s usage history", errValueUnavailable, period, metric)
	}

	usage := current - baseline
	if usage < 0 {
		return 0, nil
	}
	return usage, nil
}

func (s *usageStore) RatePerHour(snapshot *snapshot, metric string, window, minAge time.Duration) (float64, error) {
	currentTotals := snapshotUsageTotals(snapshot)
	current, ok := currentTotals[metric]
	if !ok {
		return 0, fmt.Errorf("%w: %s total", errValueUnavailable, metric)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now().UTC()
	reference, ok := s.referenceSample(metric, now, window, minAge)
	if !ok {
		return 0, fmt.Errorf("%w: %s rate history", errValueUnavailable, metric)
	}

	baseline, ok := reference.Totals[metric]
	if !ok {
		return 0, fmt.Errorf("%w: %s rate history", errValueUnavailable, metric)
	}

	elapsedHours := now.Sub(reference.ObservedAt).Hours()
	if elapsedHours <= 0 {
		return 0, fmt.Errorf("%w: %s rate history", errValueUnavailable, metric)
	}

	rate := (current - baseline) / elapsedHours
	if rate < 0 {
		return 0, nil
	}
	return rate, nil
}

func (s *usageStore) baseline(metric string, start time.Time) (float64, bool) {
	var candidate *usageSample
	for i := len(s.data.Samples) - 1; i >= 0; i-- {
		sample := s.data.Samples[i]
		if !sample.ObservedAt.After(start) {
			candidate = &sample
			break
		}
	}

	if candidate != nil {
		value, ok := candidate.Totals[metric]
		return value, ok
	}

	for _, sample := range s.data.Samples {
		if sample.ObservedAt.Before(start) {
			continue
		}
		value, ok := sample.Totals[metric]
		if ok {
			return value, true
		}
	}

	return 0, false
}

func (s *usageStore) referenceSample(metric string, now time.Time, window, minAge time.Duration) (usageSample, bool) {
	cutoff := now.Add(-window)
	var candidate usageSample
	found := false

	for _, sample := range s.data.Samples {
		if sample.ObservedAt.After(now.Add(-minAge)) {
			continue
		}
		if _, ok := sample.Totals[metric]; !ok {
			continue
		}
		if sample.ObservedAt.Before(cutoff) {
			continue
		}
		candidate = sample
		found = true
		break
	}

	if found {
		return candidate, true
	}

	for i := len(s.data.Samples) - 1; i >= 0; i-- {
		sample := s.data.Samples[i]
		if sample.ObservedAt.After(now.Add(-minAge)) {
			continue
		}
		if _, ok := sample.Totals[metric]; ok {
			return sample, true
		}
	}

	return usageSample{}, false
}

func (s *usageStore) prune(now time.Time) {
	cutoff := now.AddDate(0, -2, 0)
	kept := s.data.Samples[:0]
	for _, sample := range s.data.Samples {
		if sample.ObservedAt.Before(cutoff) {
			continue
		}
		kept = append(kept, sample)
	}
	s.data.Samples = kept
}

func (s *usageStore) save() error {
	payload, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal usage history: %w", err)
	}

	if err := os.WriteFile(s.path, payload, 0o600); err != nil {
		return fmt.Errorf("write usage history: %w", err)
	}

	return nil
}

func periodStart(now time.Time, period usagePeriod) time.Time {
	local := now.Local()
	switch period {
	case usageDaily:
		return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location()).UTC()
	case usageWeekly:
		weekday := int(local.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
		return start.AddDate(0, 0, -(weekday - 1)).UTC()
	case usageMonthly:
		return time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, local.Location()).UTC()
	default:
		return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location()).UTC()
	}
}

func sameTotals(left, right map[string]float64) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftValue := range left {
		rightValue, ok := right[key]
		if !ok || rightValue != leftValue {
			return false
		}
	}
	return true
}

func snapshotUsageTotals(s *snapshot) map[string]float64 {
	totals := make(map[string]float64)

	if value, err := measurementNumberAny(s, "energy_import_kwh", "total_power_import_kwh"); err == nil {
		totals["electricity_import_kwh"] = value
	}
	if value, err := measurementNumberAny(s, "energy_export_kwh", "total_power_export_kwh"); err == nil {
		totals["electricity_export_kwh"] = value
	}

	if gas, ok := s.Utilities["gas"]; ok {
		if value, ok := numberValue(gas.Value); ok {
			totals["gas_m3"] = value
		}
	}
	if water, ok := s.Utilities["water"]; ok {
		if value, ok := numberValue(water.Value); ok {
			totals["water_m3"] = value
		}
	}

	return totals
}
