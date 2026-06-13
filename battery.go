package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type PowerSupply struct {
	Name       string
	Type       string
	Properties map[string]string
	UpdatedAt  time.Time
}

func (ps PowerSupply) Prop(key string) string {
	return ps.Properties[key]
}

func (ps PowerSupply) IntProp(key string) (int64, bool) {
	v, ok := ps.Properties[key]
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 64)
	return n, err == nil
}

func (ps PowerSupply) FloatProp(key string) (float64, bool) {
	v, ok := ps.Properties[key]
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseFloat(v, 64)
	return n, err == nil
}

func (ps PowerSupply) IsBattery() bool {
	return ps.Type == "Battery"
}

func (ps PowerSupply) IsOnline() bool {
	v, ok := ps.Properties["POWER_SUPPLY_ONLINE"]
	if !ok {
		return false
	}
	return v == "1"
}

func (ps PowerSupply) HasCapacity() bool {
	_, ok := ps.Properties["POWER_SUPPLY_CAPACITY"]
	return ok
}

func (ps PowerSupply) Capacity() float64 {
	if !ps.IsBattery() {
		return 0
	}
	v, ok := ps.FloatProp("POWER_SUPPLY_CAPACITY")
	if !ok {
		return 0
	}
	return v
}

func (ps PowerSupply) CapacityLevel() string {
	return ps.Prop("POWER_SUPPLY_CAPACITY_LEVEL")
}

func (ps PowerSupply) Status() string {
	return ps.Prop("POWER_SUPPLY_STATUS")
}

func (ps PowerSupply) IsCharging() bool {
	return ps.Status() == "Charging"
}

func (ps PowerSupply) IsDischarging() bool {
	return ps.Status() == "Discharging"
}

func (ps PowerSupply) IsFull() bool {
	return ps.Status() == "Full"
}

func (ps PowerSupply) EnergyNowWh() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_ENERGY_NOW")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) EnergyFullWh() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_ENERGY_FULL")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) EnergyFullDesignWh() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_ENERGY_FULL_DESIGN")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) PowerNowW() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_POWER_NOW")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) VoltageNowV() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_VOLTAGE_NOW")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) VoltageMinDesignV() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_VOLTAGE_MIN_DESIGN")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) VoltageMaxDesignV() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_VOLTAGE_MAX_DESIGN")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) CurrentNowA() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_CURRENT_NOW")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) ChargeNowAh() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_CHARGE_NOW")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) ChargeFullAh() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_CHARGE_FULL")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) ChargeFullDesignAh() float64 {
	v, ok := ps.FloatProp("POWER_SUPPLY_CHARGE_FULL_DESIGN")
	if !ok {
		return 0
	}
	return v / 1_000_000
}

func (ps PowerSupply) CycleCount() int64 {
	v, ok := ps.IntProp("POWER_SUPPLY_CYCLE_COUNT")
	if !ok {
		return 0
	}
	return v
}

func (ps PowerSupply) Temperature() string {
	raw, ok := ps.Properties["POWER_SUPPLY_TEMP"]
	if !ok {
		return ""
	}
	temp, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return raw
	}
	return fmt.Sprintf("%.1f°C", float64(temp)/10)
}

func (ps PowerSupply) Technology() string {
	return ps.Prop("POWER_SUPPLY_TECHNOLOGY")
}

func (ps PowerSupply) Manufacturer() string {
	return ps.Prop("POWER_SUPPLY_MANUFACTURER")
}

func (ps PowerSupply) ModelName() string {
	return ps.Prop("POWER_SUPPLY_MODEL_NAME")
}

func (ps PowerSupply) SerialNumber() string {
	return ps.Prop("POWER_SUPPLY_SERIAL_NUMBER")
}

func (ps PowerSupply) WearLevel() float64 {
	full := ps.EnergyFullWh()
	design := ps.EnergyFullDesignWh()
	if full <= 0 || design <= 0 {
		return 0
	}
	return math.Round((1-full/design)*100*10) / 10
}

func (ps PowerSupply) TimeToEmpty() string {
	v, ok := ps.IntProp("POWER_SUPPLY_TIME_TO_EMPTY_NOW")
	if !ok || v <= 0 {
		return ""
	}
	return formatDuration(v)
}

func (ps PowerSupply) TimeToFull() string {
	v, ok := ps.IntProp("POWER_SUPPLY_TIME_TO_FULL_NOW")
	if !ok || v <= 0 {
		return ""
	}
	return formatDuration(v)
}

func formatDuration(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func ScanPowerSupplies() []PowerSupply {
	dir := "/sys/class/power_supply"
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var supplies []PowerSupply
	for _, ent := range ents {
		name := ent.Name()
		ueventPath := filepath.Join(dir, name, "uevent")
		props := readUevent(ueventPath)

		// Also read the type file for non-uevent systems
		typePath := filepath.Join(dir, name, "type")
		typ := props["POWER_SUPPLY_TYPE"]
		if typ == "" {
			typ = readFirstLine(typePath)
		}

		supplies = append(supplies, PowerSupply{
			Name:       name,
			Type:       typ,
			Properties: props,
			UpdatedAt:  time.Now(),
		})
	}
	return supplies
}

func readUevent(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	props := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			props[parts[0]] = parts[1]
		}
	}
	return props
}

func readFirstLine(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func SortedProps(props map[string]string) []string {
	var keys []string
	for k := range props {
		keys = append(keys, k)
	}
	sortKeys(keys)
	return keys
}

func sortKeys(keys []string) {
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
}
