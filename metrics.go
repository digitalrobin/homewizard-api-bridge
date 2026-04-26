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

func metricRoutes(usage *usageStore) []metricRoute {
	return []metricRoute{
		{
			Group:       "electricity",
			Path:        "/electricity/usage",
			Description: "Current net electricity usage in watts. Negative export from HomeWizard is converted to positive export on the dedicated export endpoint.",
			Unit:        "W",
			Source:      "measurement.power_w or measurement.active_power_w",
			Example:     "428",
			Resolver: func(s *snapshot) (any, error) {
				return measurementNumberAny(s, "power_w", "active_power_w")
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/import-usage",
			Description: "Current imported electricity usage in watts.",
			Unit:        "W",
			Source:      "derived from measurement.power_w or measurement.active_power_w",
			Example:     "428",
			Resolver: func(s *snapshot) (any, error) {
				value, err := measurementNumberAny(s, "power_w", "active_power_w")
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
			Source:      "derived from measurement.power_w or measurement.active_power_w",
			Example:     "678",
			Resolver: func(s *snapshot) (any, error) {
				value, err := measurementNumberAny(s, "power_w", "active_power_w")
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
			Source:      "measurement.energy_import_kwh or measurement.total_power_import_kwh",
			Example:     "13779.338",
			Resolver: func(s *snapshot) (any, error) {
				return measurementNumberAny(s, "energy_import_kwh", "total_power_import_kwh")
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/total-export",
			Description: "Total exported electricity in kWh.",
			Unit:        "kWh",
			Source:      "measurement.energy_export_kwh or measurement.total_power_export_kwh",
			Example:     "2876.514",
			Resolver: func(s *snapshot) (any, error) {
				return measurementNumberAny(s, "energy_export_kwh", "total_power_export_kwh")
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/daily-usage",
			Description: "Imported electricity usage since the start of the current day.",
			Unit:        "kWh",
			Source:      "derived from persisted total import history",
			Example:     "8.412",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "electricity_import_kwh", usageDaily)
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/weekly-usage",
			Description: "Imported electricity usage since the start of the current week.",
			Unit:        "kWh",
			Source:      "derived from persisted total import history",
			Example:     "53.928",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "electricity_import_kwh", usageWeekly)
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/monthly-usage",
			Description: "Imported electricity usage since the start of the current month.",
			Unit:        "kWh",
			Source:      "derived from persisted total import history",
			Example:     "241.003",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "electricity_import_kwh", usageMonthly)
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/daily-export",
			Description: "Exported electricity since the start of the current day.",
			Unit:        "kWh",
			Source:      "derived from persisted total export history",
			Example:     "4.331",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "electricity_export_kwh", usageDaily)
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/weekly-export",
			Description: "Exported electricity since the start of the current week.",
			Unit:        "kWh",
			Source:      "derived from persisted total export history",
			Example:     "27.842",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "electricity_export_kwh", usageWeekly)
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/monthly-export",
			Description: "Exported electricity since the start of the current month.",
			Unit:        "kWh",
			Source:      "derived from persisted total export history",
			Example:     "118.443",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "electricity_export_kwh", usageMonthly)
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/tariff",
			Description: "Currently active tariff.",
			Source:      "measurement.tariff or measurement.active_tariff",
			Example:     "2",
			Resolver: func(s *snapshot) (any, error) {
				return measurementNumberAny(s, "tariff", "active_tariff")
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/usage-l1",
			Description: "Current electricity usage for phase L1 in watts.",
			Unit:        "W",
			Source:      "measurement.power_l1_w or measurement.active_power_l1_w",
			Example:     "-676",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "power_l1_w", "active_power_l1_w") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/usage-l2",
			Description: "Current electricity usage for phase L2 in watts.",
			Unit:        "W",
			Source:      "measurement.power_l2_w or measurement.active_power_l2_w",
			Example:     "133",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "power_l2_w", "active_power_l2_w") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/usage-l3",
			Description: "Current electricity usage for phase L3 in watts.",
			Unit:        "W",
			Source:      "measurement.power_l3_w or measurement.active_power_l3_w",
			Example:     "0",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "power_l3_w", "active_power_l3_w") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/current-total",
			Description: "Total current in amperes.",
			Unit:        "A",
			Source:      "measurement.current_a or measurement.active_current_a",
			Example:     "6",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "current_a", "active_current_a") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/current-l1",
			Description: "Current on phase L1 in amperes.",
			Unit:        "A",
			Source:      "measurement.current_l1_a or measurement.active_current_l1_a",
			Example:     "-4",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "current_l1_a", "active_current_l1_a") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/current-l2",
			Description: "Current on phase L2 in amperes.",
			Unit:        "A",
			Source:      "measurement.current_l2_a or measurement.active_current_l2_a",
			Example:     "2",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "current_l2_a", "active_current_l2_a") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/current-l3",
			Description: "Current on phase L3 in amperes.",
			Unit:        "A",
			Source:      "measurement.current_l3_a or measurement.active_current_l3_a",
			Example:     "0",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "current_l3_a", "active_current_l3_a") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/voltage",
			Description: "Voltage in volts when the meter exposes a single combined voltage value.",
			Unit:        "V",
			Source:      "measurement.voltage_v or measurement.active_voltage_v",
			Example:     "236.0",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "voltage_v", "active_voltage_v") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/voltage-l1",
			Description: "Voltage on phase L1 in volts.",
			Unit:        "V",
			Source:      "measurement.voltage_l1_v or measurement.active_voltage_l1_v",
			Example:     "236.0",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "voltage_l1_v", "active_voltage_l1_v") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/voltage-l2",
			Description: "Voltage on phase L2 in volts.",
			Unit:        "V",
			Source:      "measurement.voltage_l2_v or measurement.active_voltage_l2_v",
			Example:     "232.6",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "voltage_l2_v", "active_voltage_l2_v") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/voltage-l3",
			Description: "Voltage on phase L3 in volts.",
			Unit:        "V",
			Source:      "measurement.voltage_l3_v or measurement.active_voltage_l3_v",
			Example:     "235.1",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "voltage_l3_v", "active_voltage_l3_v") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/frequency",
			Description: "Grid frequency in hertz.",
			Unit:        "Hz",
			Source:      "measurement.frequency_hz or measurement.active_frequency_hz",
			Example:     "50.0",
			Resolver:    func(s *snapshot) (any, error) { return measurementNumberAny(s, "frequency_hz", "active_frequency_hz") },
		},
		{
			Group:       "electricity",
			Path:        "/electricity/average-demand-15m",
			Description: "15-minute average demand, useful for Belgian capacity tariff monitoring.",
			Unit:        "W",
			Source:      "measurement.average_power_15m_w or measurement.active_power_average_w",
			Example:     "123",
			Resolver: func(s *snapshot) (any, error) {
				return measurementNumberAny(s, "average_power_15m_w", "active_power_average_w")
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/monthly-peak",
			Description: "Current monthly peak demand.",
			Unit:        "W",
			Source:      "measurement.monthly_power_peak_w or measurement.montly_power_peak_w",
			Example:     "1111",
			Resolver: func(s *snapshot) (any, error) {
				return measurementNumberAny(s, "monthly_power_peak_w", "montly_power_peak_w")
			},
		},
		{
			Group:       "electricity",
			Path:        "/electricity/monthly-peak-timestamp",
			Description: "Timestamp of the monthly peak demand.",
			Source:      "measurement.monthly_power_peak_timestamp or measurement.montly_power_peak_timestamp",
			Example:     "2024-06-04T10:11:22",
			Resolver: func(s *snapshot) (any, error) {
				return measurementStringAny(s, "monthly_power_peak_timestamp", "montly_power_peak_timestamp")
			},
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
			Path:        "/gas/current-usage",
			Description: "Estimated current gas usage as a rolling rate derived from recent total meter history.",
			Unit:        "m3/h",
			Source:      "derived from persisted gas total history over the recent sample window",
			Example:     "0.842",
			Resolver: func(s *snapshot) (any, error) {
				return usage.RatePerHour(s, "gas_m3", gasRateWindow, gasRateMinAge)
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
			Group:       "gas",
			Path:        "/gas/daily-usage",
			Description: "Gas usage since the start of the current day.",
			Unit:        "m3",
			Source:      "derived from persisted gas total history",
			Example:     "0.736",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "gas_m3", usageDaily)
			},
		},
		{
			Group:       "gas",
			Path:        "/gas/weekly-usage",
			Description: "Gas usage since the start of the current week.",
			Unit:        "m3",
			Source:      "derived from persisted gas total history",
			Example:     "4.112",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "gas_m3", usageWeekly)
			},
		},
		{
			Group:       "gas",
			Path:        "/gas/monthly-usage",
			Description: "Gas usage since the start of the current month.",
			Unit:        "m3",
			Source:      "derived from persisted gas total history",
			Example:     "18.903",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "gas_m3", usageMonthly)
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
		{
			Group:       "water",
			Path:        "/water/daily-usage",
			Description: "Water usage since the start of the current day.",
			Unit:        "m3",
			Source:      "derived from persisted water total history",
			Example:     "0.182",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "water_m3", usageDaily)
			},
		},
		{
			Group:       "water",
			Path:        "/water/weekly-usage",
			Description: "Water usage since the start of the current week.",
			Unit:        "m3",
			Source:      "derived from persisted water total history",
			Example:     "1.044",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "water_m3", usageWeekly)
			},
		},
		{
			Group:       "water",
			Path:        "/water/monthly-usage",
			Description: "Water usage since the start of the current month.",
			Unit:        "m3",
			Source:      "derived from persisted water total history",
			Example:     "4.327",
			Resolver: func(s *snapshot) (any, error) {
				return usage.Usage(s, "water_m3", usageMonthly)
			},
		},
	}
}

func metricRouteMap(usage *usageStore) map[string]metricRoute {
	routes := metricRoutes(usage)
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
