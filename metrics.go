package main

import (
	"fmt"
	"sort"
)

type metricResolver func(*snapshot) (any, error)

type metricRoute struct {
	Group       string `json:"group"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Unit        string `json:"unit,omitempty"`
	Source      string `json:"source"`
	Example     string `json:"example,omitempty"`
	Resolver    metricResolver
}

func metricRoutes() []metricRoute {
	return []metricRoute{
		{
			Group:       "electricity",
			Path:        "/electricity/usage",
			Description: "Current net electricity usage in watts. Negative export from HomeWizard is converted to positive export on the dedicated export endpoint.",
			Unit:        "W",
			Source:      "measurement.power_w",
			Example:     "428",
			Resolver: func(s *snapshot) (any, error) {
				return measurementNumber(s, "power_w")
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/import-usage",
			Description: "Current imported electricity usage in watts.",
			Unit:        "W",
			Source:      "derived from measurement.power_w",
			Example:     "428",
			Resolver: func(s *snapshot) (any, error) {
				value, err := measurementNumber(s, "power_w")
				if err != nil {
					return nil, err
				}
				return positivePart(value), nil
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/export-usage",
			Description: "Current exported electricity usage in watts.",
			Unit:        "W",
			Source:      "derived from measurement.power_w",
			Example:     "678",
			Resolver: func(s *snapshot) (any, error) {
				value, err := measurementNumber(s, "power_w")
				if err != nil {
					return nil, err
				}
				return negativeMagnitude(value), nil
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/total-usage",
			Description: "Total imported electricity in kWh.",
			Unit:        "kWh",
			Source:      "measurement.energy_import_kwh",
			Example:     "13779.338",
			Resolver: func(s *snapshot) (any, error) {
				return measurementNumber(s, "energy_import_kwh")
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/total-export",
			Description: "Total exported electricity in kWh.",
			Unit:        "kWh",
			Source:      "measurement.energy_export_kwh",
			Example:     "2876.514",
			Resolver: func(s *snapshot) (any, error) {
				return measurementNumber(s, "energy_export_kwh")
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/tariff",
			Description: "Currently active tariff.",
			Source:      "measurement.tariff",
			Example:     "2",
			Resolver: func(s *snapshot) (any, error) {
				return measurementNumber(s, "tariff")
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/usage-l1",
			Description: "Current electricity usage for phase L1 in watts.",
			Unit:        "W",
			Source:      "measurement.power_l1_w",
			Example:     "-676",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "power_l1_w") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/usage-l2",
			Description: "Current electricity usage for phase L2 in watts.",
			Unit:        "W",
			Source:      "measurement.power_l2_w",
			Example:     "133",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "power_l2_w") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/usage-l3",
			Description: "Current electricity usage for phase L3 in watts.",
			Unit:        "W",
			Source:      "measurement.power_l3_w",
			Example:     "0",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "power_l3_w") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/current-total",
			Description: "Total current in amperes.",
			Unit:        "A",
			Source:      "measurement.current_a",
			Example:     "6",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "current_a") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/current-l1",
			Description: "Current on phase L1 in amperes.",
			Unit:        "A",
			Source:      "measurement.current_l1_a",
			Example:     "-4",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "current_l1_a") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/current-l2",
			Description: "Current on phase L2 in amperes.",
			Unit:        "A",
			Source:      "measurement.current_l2_a",
			Example:     "2",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "current_l2_a") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/current-l3",
			Description: "Current on phase L3 in amperes.",
			Unit:        "A",
			Source:      "measurement.current_l3_a",
			Example:     "0",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "current_l3_a") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/voltage",
			Description: "Voltage in volts when the meter exposes a single combined voltage value.",
			Unit:        "V",
			Source:      "measurement.voltage_v",
			Example:     "236.0",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "voltage_v") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/voltage-l1",
			Description: "Voltage on phase L1 in volts.",
			Unit:        "V",
			Source:      "measurement.voltage_l1_v",
			Example:     "236.0",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "voltage_l1_v") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/voltage-l2",
			Description: "Voltage on phase L2 in volts.",
			Unit:        "V",
			Source:      "measurement.voltage_l2_v",
			Example:     "232.6",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "voltage_l2_v") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/voltage-l3",
			Description: "Voltage on phase L3 in volts.",
			Unit:        "V",
			Source:      "measurement.voltage_l3_v",
			Example:     "235.1",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "voltage_l3_v") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/frequency",
			Description: "Grid frequency in hertz.",
			Unit:        "Hz",
			Source:      "measurement.frequency_hz",
			Example:     "50.0",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "frequency_hz") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/average-demand-15m",
			Description: "15-minute average demand, useful for Belgian capacity tariff monitoring.",
			Unit:        "W",
			Source:      "measurement.average_power_15m_w",
			Example:     "123",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "average_power_15m_w") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/monthly-peak",
			Description: "Current monthly peak demand.",
			Unit:        "W",
			Source:      "measurement.monthly_power_peak_w",
			Example:     "1111",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumber(s, "monthly_power_peak_w") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/monthly-peak-timestamp",
			Description: "Timestamp of the monthly peak demand.",
			Source:      "measurement.monthly_power_peak_timestamp",
			Example:     "2024-06-04T10:11:22",
			Resolver:    func(s *snapshot) (any, error) { return measurementString(s, "monthly_power_peak_timestamp") },
		},
		{
			Group:       "gas",
			Path:        "/gas/total-usage",
			Description: "Total gas meter reading.",
			Source:      "measurement.external[type=gas_meter].value",
			Example:     "2569.646",
			Resolver: func(s *snapshot) (any, error) {
				utility, err := utilityValue(s, "gas")
				if err != nil {
					return nil, err
				}
				return utility.Value, nil
			},
		},
		{
			Group:       "gas",
			Path:        "/gas/timestamp",
			Description: "Timestamp of the latest gas reading.",
			Source:      "measurement.external[type=gas_meter].timestamp",
			Example:     "2021-06-06T14:00:10",
			Resolver: func(s *snapshot) (any, error) {
				utility, err := utilityValue(s, "gas")
				if err != nil {
					return nil, err
				}
				if utility.Timestamp == "" {
					return nil, fmt.Errorf("%w: gas timestamp", errValueUnavailable)
				}
				return utility.Timestamp, nil
			},
		},
		{
			Group:       "gas",
			Path:        "/gas/unit",
			Description: "Unit used by the gas meter, usually m3.",
			Source:      "measurement.external[type=gas_meter].unit",
			Example:     "m3",
			Resolver: func(s *snapshot) (any, error) {
				utility, err := utilityValue(s, "gas")
				if err != nil {
					return nil, err
				}
				if utility.Unit == "" {
					return nil, fmt.Errorf("%w: gas unit", errValueUnavailable)
				}
				return utility.Unit, nil
			},
		},
		{
			Group:       "water",
			Path:        "/water/total-usage",
			Description: "Total water meter reading.",
			Source:      "measurement.external[type=water_meter].value",
			Example:     "123.456",
			Resolver: func(s *snapshot) (any, error) {
				utility, err := utilityValue(s, "water")
				if err != nil {
					return nil, err
				}
				return utility.Value, nil
			},
		},
		{
			Group:       "water",
			Path:        "/water/timestamp",
			Description: "Timestamp of the latest water reading.",
			Source:      "measurement.external[type=water_meter].timestamp",
			Example:     "2024-06-28T14:12:34",
			Resolver: func(s *snapshot) (any, error) {
				utility, err := utilityValue(s, "water")
				if err != nil {
					return nil, err
				}
				if utility.Timestamp == "" {
					return nil, fmt.Errorf("%w: water timestamp", errValueUnavailable)
				}
				return utility.Timestamp, nil
			},
		},
		{
			Group:       "water",
			Path:        "/water/unit",
			Description: "Unit used by the water meter.",
			Source:      "measurement.external[type=water_meter].unit",
			Example:     "m3",
			Resolver: func(s *snapshot) (any, error) {
				utility, err := utilityValue(s, "water")
				if err != nil {
					return nil, err
				}
				if utility.Unit == "" {
					return nil, fmt.Errorf("%w: water unit", errValueUnavailable)
				}
				return utility.Unit, nil
			},
		},
	}
}

func metricRouteMap() map[string]metricRoute {
	routes := metricRoutes()
	index := make(map[string]metricRoute, len(routes))
	for _, route := range routes {
		index[route.Path] = route
	}
	return index
}

func metricPaths(routes map[string]metricRoute) []string {
	paths := make([]string, 0, len(routes))
	for path := range routes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func metricGroups(routes map[string]metricRoute) map[string][]metricRoute {
	grouped := make(map[string][]metricRoute)
	for _, route := range routes {
		grouped[route.Group] = append(grouped[route.Group], route)
	}

	for group := range grouped {
		sort.Slice(grouped[group], func(i, j int) bool {
			return grouped[group][i].Path < grouped[group][j].Path
		})
	}

	return grouped
}
