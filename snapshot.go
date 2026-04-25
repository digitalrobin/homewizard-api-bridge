package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
)

var errValueUnavailable = errors.New("value unavailable")

type snapshot struct {
	Device      deviceInfo              `json:"device"`
	Measurement measurement             `json:"measurement"`
	Utilities   map[string]utilityMeter `json:"utilities"`
}

type utilityMeter struct {
	Type      string `json:"type"`
	UniqueID  string `json:"unique_id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Value     any    `json:"value,omitempty"`
}

func (c *homeWizardClient) Snapshot(ctx context.Context) (*snapshot, error) {
	device, err := c.GetDeviceInfo(ctx)
	if err != nil {
		return nil, err
	}

	measurement, err := c.GetMeasurement(ctx)
	if err != nil {
		return nil, err
	}

	return &snapshot{
		Device:      device,
		Measurement: measurement,
		Utilities:   collectUtilityMeters(measurement),
	}, nil
}

func collectUtilityMeters(m measurement) map[string]utilityMeter {
	utilities := make(map[string]utilityMeter)

	external, ok := m["external"].([]any)
	if !ok {
		return utilities
	}

	for _, raw := range external {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}

		meter := utilityMeter{
			Type:      stringValue(entry["type"]),
			UniqueID:  stringValue(entry["unique_id"]),
			Timestamp: stringValue(entry["timestamp"]),
			Unit:      stringValue(entry["unit"]),
			Value:     entry["value"],
		}

		switch meter.Type {
		case "gas_meter":
			utilities["gas"] = meter
		case "water_meter":
			utilities["water"] = meter
		}
	}

	return utilities
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func numberValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	default:
		return 0, false
	}
}

func measurementNumber(s *snapshot, key string) (float64, error) {
	value, ok := s.Measurement[key]
	if !ok {
		return 0, fmt.Errorf("%w: measurement.%s", errValueUnavailable, key)
	}

	number, ok := numberValue(value)
	if !ok {
		return 0, fmt.Errorf("%w: measurement.%s", errValueUnavailable, key)
	}

	return number, nil
}

func measurementString(s *snapshot, key string) (string, error) {
	value, ok := s.Measurement[key]
	if !ok {
		return "", fmt.Errorf("%w: measurement.%s", errValueUnavailable, key)
	}

	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%w: measurement.%s", errValueUnavailable, key)
	}

	return text, nil
}

func utilityValue(s *snapshot, name string) (utilityMeter, error) {
	utility, ok := s.Utilities[name]
	if !ok {
		return utilityMeter{}, fmt.Errorf("%w: %s meter", errValueUnavailable, name)
	}
	return utility, nil
}

func positivePart(value float64) float64 {
	return math.Max(value, 0)
}

func negativeMagnitude(value float64) float64 {
	return math.Abs(math.Min(value, 0))
}
