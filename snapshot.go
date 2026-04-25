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
	APIMode      string                  `json:"api_mode"`
	AuthRequired bool                    `json:"auth_required"`
	Device       deviceInfo              `json:"device"`
	Measurement  measurement             `json:"measurement"`
	Utilities    map[string]utilityMeter `json:"utilities"`
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
		APIMode:      string(c.mode),
		AuthRequired: c.AuthRequired(),
		Device:       device,
		Measurement:  measurement,
		Utilities:    collectUtilityMeters(measurement),
	}, nil
}

func collectUtilityMeters(m measurement) map[string]utilityMeter {
	utilities := make(map[string]utilityMeter)

	external, ok := m["external"].([]any)
	if ok {
		for _, raw := range external {
			entry, ok := raw.(map[string]any)
			if !ok {
				continue
			}

			meter := utilityMeter{
				Type:      stringValue(entry["type"]),
				UniqueID:  stringValue(entry["unique_id"]),
				Timestamp: plainStringValue(entry["timestamp"]),
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
	}

	if _, ok := utilities["gas"]; !ok {
		if value, ok := m["total_gas_m3"]; ok {
			utilities["gas"] = utilityMeter{
				Type:      "gas_meter",
				UniqueID:  plainStringValue(m["gas_unique_id"]),
				Timestamp: plainStringValue(m["gas_timestamp"]),
				Unit:      "m3",
				Value:     value,
			}
		}
	}

	return utilities
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func plainStringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case nil:
		return ""
	default:
		return formatPlainValue(typed)
	}
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

func measurementNumberAny(s *snapshot, keys ...string) (float64, error) {
	value, key, ok := measurementValueAny(s, keys...)
	if !ok {
		return 0, fmt.Errorf("%w: measurement.%s", errValueUnavailable, keys[0])
	}

	number, ok := numberValue(value)
	if !ok {
		return 0, fmt.Errorf("%w: measurement.%s", errValueUnavailable, key)
	}

	return number, nil
}

func measurementValueAny(s *snapshot, keys ...string) (any, string, bool) {
	for _, key := range keys {
		value, ok := s.Measurement[key]
		if ok {
			return value, key, true
		}
	}
	return nil, "", false
}

func measurementStringAny(s *snapshot, keys ...string) (string, error) {
	value, key, ok := measurementValueAny(s, keys...)
	if !ok {
		return "", fmt.Errorf("%w: measurement.%s", errValueUnavailable, keys[0])
	}

	return plainStringValueForKey(value, key)
}

func plainStringValueForKey(value any, key string) (string, error) {
	text := plainStringValue(value)
	if text == "" {
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
