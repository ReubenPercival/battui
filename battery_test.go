package main

import (
	"testing"
	"time"
)

func testSupply(typ string, props map[string]string) PowerSupply {
	if props == nil {
		props = map[string]string{}
	}
	return PowerSupply{Name: "test", Type: typ, Properties: props, UpdatedAt: time.Now()}
}

func TestIsBattery(t *testing.T) {
	if !testSupply("Battery", nil).IsBattery() {
		t.Error("expected IsBattery() = true for Battery type")
	}
	if testSupply("Mains", nil).IsBattery() {
		t.Error("expected IsBattery() = false for Mains type")
	}
	if testSupply("UPS", nil).IsBattery() {
		t.Error("expected IsBattery() = false for UPS type")
	}
}

func TestIsCharging(t *testing.T) {
	if !testSupply("Battery", map[string]string{"POWER_SUPPLY_STATUS": "Charging"}).IsCharging() {
		t.Error("expected IsCharging() = true for Charging status")
	}
	if testSupply("Battery", map[string]string{"POWER_SUPPLY_STATUS": "Discharging"}).IsCharging() {
		t.Error("expected IsCharging() = false for Discharging status")
	}
}

func TestIsDischarging(t *testing.T) {
	if !testSupply("Battery", map[string]string{"POWER_SUPPLY_STATUS": "Discharging"}).IsDischarging() {
		t.Error("expected IsDischarging() = true")
	}
}

func TestIsFull(t *testing.T) {
	if !testSupply("Battery", map[string]string{"POWER_SUPPLY_STATUS": "Full"}).IsFull() {
		t.Error("expected IsFull() = true")
	}
}

func TestIsOnline(t *testing.T) {
	if !testSupply("Mains", map[string]string{"POWER_SUPPLY_ONLINE": "1"}).IsOnline() {
		t.Error("expected IsOnline() = true for ONLINE=1")
	}
	if testSupply("Mains", map[string]string{"POWER_SUPPLY_ONLINE": "0"}).IsOnline() {
		t.Error("expected IsOnline() = false for ONLINE=0")
	}
	if testSupply("Mains", nil).IsOnline() {
		t.Error("expected IsOnline() = false for missing key")
	}
}

func TestHasCapacity(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_CAPACITY": "75"})
	if !ps.HasCapacity() {
		t.Error("expected HasCapacity() = true for Battery with CAPACITY")
	}
	ps = testSupply("Battery", map[string]string{})
	if ps.HasCapacity() {
		t.Error("expected HasCapacity() = false for Battery without CAPACITY")
	}
	ps = testSupply("Mains", map[string]string{"POWER_SUPPLY_CAPACITY": "100"})
	if ps.HasCapacity() {
		t.Error("expected HasCapacity() = false for Mains with CAPACITY")
	}
}

func TestCapacity(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_CAPACITY": "75.5"})
	if c := ps.Capacity(); c != 75.5 {
		t.Errorf("expected Capacity() = 75.5, got %f", c)
	}
	ps = testSupply("Battery", map[string]string{})
	if c := ps.Capacity(); c != 0 {
		t.Errorf("expected Capacity() = 0 for missing key, got %f", c)
	}
	ps = testSupply("Mains", map[string]string{"POWER_SUPPLY_CAPACITY": "100"})
	if c := ps.Capacity(); c != 0 {
		t.Errorf("expected Capacity() = 0 for non-battery, got %f", c)
	}
}

func TestWearLevel(t *testing.T) {
	tests := []struct {
		full   string
		design string
		want   float64
	}{
		{"50000000", "50000000", 0},
		{"45000000", "50000000", 10},
		{"40000000", "50000000", 20},
		{"25000000", "50000000", 50},
		{"0", "50000000", 0},
		{"50000000", "0", 0},
		{"55000000", "50000000", 0},
	}
	for _, tt := range tests {
		ps := testSupply("Battery", map[string]string{
			"POWER_SUPPLY_ENERGY_FULL":        tt.full,
			"POWER_SUPPLY_ENERGY_FULL_DESIGN": tt.design,
		})
		got := ps.WearLevel()
		if got != tt.want {
			t.Errorf("WearLevel(full=%s, design=%s) = %f, want %f", tt.full, tt.design, got, tt.want)
		}
	}
}

func TestProp(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_STATUS": "Charging"})
	if ps.Prop("POWER_SUPPLY_STATUS") != "Charging" {
		t.Error("expected Prop to return 'Charging'")
	}
	if ps.Prop("POWER_SUPPLY_MISSING") != "" {
		t.Error("expected Prop to return '' for missing key")
	}
}

func TestIntProp(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_CYCLE_COUNT": "150"})
	v, ok := ps.IntProp("POWER_SUPPLY_CYCLE_COUNT")
	if !ok || v != 150 {
		t.Errorf("expected IntProp = (150, true), got (%d, %v)", v, ok)
	}
	_, ok = ps.IntProp("POWER_SUPPLY_MISSING")
	if ok {
		t.Error("expected IntProp ok = false for missing key")
	}
	v, ok = ps.IntProp("POWER_SUPPLY_STATUS")
	if ok {
		t.Error("expected IntProp ok = false for non-integer value")
	}
}

func TestFloatProp(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_VOLTAGE_NOW": "12000000"})
	v, ok := ps.FloatProp("POWER_SUPPLY_VOLTAGE_NOW")
	if !ok || v != 12000000 {
		t.Errorf("expected FloatProp = (12000000, true), got (%f, %v)", v, ok)
	}
	_, ok = ps.FloatProp("POWER_SUPPLY_MISSING")
	if ok {
		t.Error("expected FloatProp ok = false for missing key")
	}
}

func TestEnergyNowWh(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_ENERGY_NOW": "30000000"})
	if v := ps.EnergyNowWh(); v != 30 {
		t.Errorf("expected EnergyNowWh = 30, got %f", v)
	}
	ps = testSupply("Battery", map[string]string{})
	if v := ps.EnergyNowWh(); v != 0 {
		t.Errorf("expected EnergyNowWh = 0 for missing key, got %f", v)
	}
}

func TestEnergyFullWh(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_ENERGY_FULL": "50000000"})
	if v := ps.EnergyFullWh(); v != 50 {
		t.Errorf("expected EnergyFullWh = 50, got %f", v)
	}
}

func TestEnergyFullDesignWh(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_ENERGY_FULL_DESIGN": "55000000"})
	if v := ps.EnergyFullDesignWh(); v != 55 {
		t.Errorf("expected EnergyFullDesignWh = 55, got %f", v)
	}
}

func TestPowerNowW(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_POWER_NOW": "15000000"})
	if v := ps.PowerNowW(); v != 15 {
		t.Errorf("expected PowerNowW = 15, got %f", v)
	}
}

func TestVoltageNowV(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_VOLTAGE_NOW": "12000000"})
	if v := ps.VoltageNowV(); v != 12 {
		t.Errorf("expected VoltageNowV = 12, got %f", v)
	}
}

func TestCycleCount(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_CYCLE_COUNT": "300"})
	if v := ps.CycleCount(); v != 300 {
		t.Errorf("expected CycleCount = 300, got %d", v)
	}
}

func TestTemperature(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_TEMP": "350"})
	if v := ps.Temperature(); v != "35.0°C" {
		t.Errorf("expected Temperature = '35.0°C', got '%s'", v)
	}
	ps = testSupply("Battery", map[string]string{})
	if v := ps.Temperature(); v != "" {
		t.Errorf("expected Temperature = '' for missing key, got '%s'", v)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds int64
		want    string
	}{
		{0, "0s"},
		{30, "30s"},
		{90, "1m 30s"},
		{3600, "1h 0m 0s"},
		{3661, "1h 1m 1s"},
		{86400, "24h 0m 0s"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.seconds)
		if got != tt.want {
			t.Errorf("formatDuration(%d) = '%s', want '%s'", tt.seconds, got, tt.want)
		}
	}
}

func TestTimeToEmpty(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_TIME_TO_EMPTY_NOW": "3600"})
	if v := ps.TimeToEmpty(); v != "1h 0m 0s" {
		t.Errorf("expected TimeToEmpty = '1h 0m 0s', got '%s'", v)
	}
	ps = testSupply("Battery", map[string]string{})
	if v := ps.TimeToEmpty(); v != "" {
		t.Errorf("expected TimeToEmpty = '' for missing key, got '%s'", v)
	}
	ps = testSupply("Battery", map[string]string{"POWER_SUPPLY_TIME_TO_EMPTY_NOW": "0"})
	if v := ps.TimeToEmpty(); v != "" {
		t.Errorf("expected TimeToEmpty = '' for zero value, got '%s'", v)
	}
}

func TestTimeToFull(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_TIME_TO_FULL_NOW": "1800"})
	if v := ps.TimeToFull(); v != "30m 0s" {
		t.Errorf("expected TimeToFull = '30m 0s', got '%s'", v)
	}
}

func TestManufacturer(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_MANUFACTURER": "LGC"})
	if v := ps.Manufacturer(); v != "LGC" {
		t.Errorf("expected Manufacturer = 'LGC', got '%s'", v)
	}
}

func TestModelName(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_MODEL_NAME": "AC14B8K"})
	if v := ps.ModelName(); v != "AC14B8K" {
		t.Errorf("expected ModelName = 'AC14B8K', got '%s'", v)
	}
}

func TestSerialNumber(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_SERIAL_NUMBER": "12345"})
	if v := ps.SerialNumber(); v != "12345" {
		t.Errorf("expected SerialNumber = '12345', got '%s'", v)
	}
}

func TestTechnology(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_TECHNOLOGY": "Li-ion"})
	if v := ps.Technology(); v != "Li-ion" {
		t.Errorf("expected Technology = 'Li-ion', got '%s'", v)
	}
}

func TestCapacityLevel(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_CAPACITY_LEVEL": "Normal"})
	if v := ps.CapacityLevel(); v != "Normal" {
		t.Errorf("expected CapacityLevel = 'Normal', got '%s'", v)
	}
}

func TestStatus(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_STATUS": "Charging"})
	if v := ps.Status(); v != "Charging" {
		t.Errorf("expected Status = 'Charging', got '%s'", v)
	}
}

func TestChargeNowAh(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_CHARGE_NOW": "3000000"})
	if v := ps.ChargeNowAh(); v != 3 {
		t.Errorf("expected ChargeNowAh = 3, got %f", v)
	}
}

func TestChargeFullAh(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_CHARGE_FULL": "4000000"})
	if v := ps.ChargeFullAh(); v != 4 {
		t.Errorf("expected ChargeFullAh = 4, got %f", v)
	}
}

func TestCurrentNowA(t *testing.T) {
	ps := testSupply("Battery", map[string]string{"POWER_SUPPLY_CURRENT_NOW": "2000000"})
	if v := ps.CurrentNowA(); v != 2 {
		t.Errorf("expected CurrentNowA = 2, got %f", v)
	}
}

func TestSortedProps(t *testing.T) {
	props := map[string]string{
		"POWER_SUPPLY_CAPACITY": "75",
		"POWER_SUPPLY_STATUS":   "Charging",
		"POWER_SUPPLY_ONLINE":   "1",
	}
	keys := SortedProps(props)
	expected := []string{"POWER_SUPPLY_CAPACITY", "POWER_SUPPLY_ONLINE", "POWER_SUPPLY_STATUS"}
	if len(keys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(keys))
	}
	for i := range keys {
		if keys[i] != expected[i] {
			t.Errorf("keys[%d] = '%s', want '%s'", i, keys[i], expected[i])
		}
	}
}

func TestSortedPropsEmpty(t *testing.T) {
	keys := SortedProps(map[string]string{})
	if len(keys) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(keys))
	}
}

func TestParseFloatSafe(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"75", 75},
		{"75.5", 75.5},
		{"0", 0},
		{"-10", -10},
		{"invalid", 0},
		{"", 0},
		{"12.345", 12.345},
	}
	for _, tt := range tests {
		got := parseFloatSafe(tt.input)
		if got != tt.want {
			t.Errorf("parseFloatSafe('%s') = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestCapacityColor(t *testing.T) {
	tests := []struct {
		pct  float64
		want string
	}{
		{5, "#f7768e"},
		{10, "#ff9e64"},
		{15, "#ff9e64"},
		{20, "#e0af68"},
		{30, "#e0af68"},
		{40, "#9ece6a"},
		{100, "#9ece6a"},
	}
	for _, tt := range tests {
		got := string(capacityColor(tt.pct))
		if got != tt.want {
			t.Errorf("capacityColor(%f) = '%s', want '%s'", tt.pct, got, tt.want)
		}
	}
}

func TestStatusColor(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"Charging", "#bb9af7"},
		{"Discharging", "#e0af68"},
		{"Full", "#9ece6a"},
		{"Not charging", "#f7768e"},
		{"Unknown", "#565f89"},
		{"", "#565f89"},
	}
	for _, tt := range tests {
		got := string(statusColor(tt.status))
		if got != tt.want {
			t.Errorf("statusColor('%s') = '%s', want '%s'", tt.status, got, tt.want)
		}
	}
}

func TestFmtW(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, "0.00 W"},
		{0.5, "500.0 mW"},
		{1, "1.00 W"},
		{1500, "1.50 kW"},
		{1000000, "1.00 MW"},
	}
	for _, tt := range tests {
		got := fmtW(tt.v)
		if got != tt.want {
			t.Errorf("fmtW(%f) = '%s', want '%s'", tt.v, got, tt.want)
		}
	}
}

func TestFmtWh(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, "0.0 Wh"},
		{50, "50.0 Wh"},
		{1500, "1.50 kWh"},
		{2000000, "2.00 MWh"},
	}
	for _, tt := range tests {
		got := fmtWh(tt.v)
		if got != tt.want {
			t.Errorf("fmtWh(%f) = '%s', want '%s'", tt.v, got, tt.want)
		}
	}
}

func TestFmtV(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, ""},
		{11.5, "11.500 V"},
		{0.5, "0.500 V"},
	}
	for _, tt := range tests {
		got := fmtV(tt.v)
		if got != tt.want {
			t.Errorf("fmtV(%f) = '%s', want '%s'", tt.v, got, tt.want)
		}
	}
}

func TestFmtA(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, ""},
		{2, "2.00 A"},
		{0.5, "500.0 mA"},
		{0.0005, "0.001 A"},
	}
	for _, tt := range tests {
		got := fmtA(tt.v)
		if got != tt.want {
			t.Errorf("fmtA(%f) = '%s', want '%s'", tt.v, got, tt.want)
		}
	}
}

func TestFmtAh(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, ""},
		{3, "3.00 Ah"},
		{0.5, "500 mAh"},
		{0.0005, "0.00 Ah"},
	}
	for _, tt := range tests {
		got := fmtAh(tt.v)
		if got != tt.want {
			t.Errorf("fmtAh(%f) = '%s', want '%s'", tt.v, got, tt.want)
		}
	}
}

func TestFmtPct(t *testing.T) {
	if s := fmtPct(75.5); s != "75.5%" {
		t.Errorf("fmtPct(75.5) = '%s', want '75.5%%'", s)
	}
	if s := fmtPct(0); s != "0.0%" {
		t.Errorf("fmtPct(0) = '%s', want '0.0%%'", s)
	}
}

func TestMax(t *testing.T) {
	if max(3, 5) != 5 {
		t.Error("max(3, 5) should be 5")
	}
	if max(5, 3) != 5 {
		t.Error("max(5, 3) should be 5")
	}
	if max(-1, 0) != 0 {
		t.Error("max(-1, 0) should be 0")
	}
}

func TestMin(t *testing.T) {
	if min(3, 5) != 3 {
		t.Error("min(3, 5) should be 3")
	}
	if min(5, 3) != 3 {
		t.Error("min(5, 3) should be 3")
	}
}

func TestClamp(t *testing.T) {
	if clamp(5, 0, 10) != 5 {
		t.Error("clamp(5, 0, 10) should be 5")
	}
	if clamp(-5, 0, 10) != 0 {
		t.Error("clamp(-5, 0, 10) should be 0")
	}
	if clamp(15, 0, 10) != 10 {
		t.Error("clamp(15, 0, 10) should be 10")
	}
}

func TestBoolStr(t *testing.T) {
	if boolStr(true) == boolStr(false) {
		t.Error("boolStr(true) and boolStr(false) should differ")
	}
}

func TestTruncate(t *testing.T) {
	if s := truncate("hello", 10); s != "hello" {
		t.Errorf("truncate('hello', 10) = '%s', want 'hello'", s)
	}
	if s := truncate("hello world", 5); s != "hell..." {
		t.Errorf("truncate('hello world', 5) = '%s', want 'hell...'", s)
	}
}
