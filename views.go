package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) renderContent() string {
	if len(m.supplies) == 0 {
		return m.renderEmpty()
	}

	switch m.activeTab {
	case tabOverview:
		return m.renderOverview()
	case tabDetails:
		return m.renderDetails()
	case tabRaw:
		return m.renderRaw()
	}
	return ""
}

func (m model) renderEmpty() string {
	msg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565f89")).
		Render("No power supply devices found.\nCheck /sys/class/power_supply/")

	w := max(40, m.width-4)
	h := max(10, m.height-6)
	return lipgloss.Place(w, h,
		lipgloss.Center, lipgloss.Center, msg)
}

func (m model) renderOverview() string {
	var cards []string
	var lastWasBattery bool
	for i, ps := range m.supplies {
		if i > 0 && ps.IsBattery() != lastWasBattery {
			sep := dimStyle.Render(strings.Repeat("─", max(20, m.width-8)))
			cards = append(cards, " "+sep)
		}
		card := m.renderSupplyCard(ps, i == m.selected)
		cards = append(cards, card)
		lastWasBattery = ps.IsBattery()
	}
	return lipgloss.JoinVertical(lipgloss.Top, cards...)
}

func (m model) renderSupplyCard(ps PowerSupply, selected bool) string {
	width := max(40, m.width-6)
	if width > 120 {
		width = 120
	}

	borderColor := lipgloss.Color("#3b4261")
	if selected {
		borderColor = lipgloss.Color("#7aa2f7")
	}

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 2).
		Width(width).
		MarginBottom(1).
		MarginRight(1)

	sel := " "
	if selected {
		sel = "▸"
	}

	titleLeft := fmt.Sprintf("%s %s", sel, cardTitleStyle.Render(ps.Name))
	if ps.Type != "" {
		titleLeft += dimStyle.Render(" (" + ps.Type + ")")
	}

	var bodyLines []string

	if ps.IsBattery() {
		hasCapacity := ps.HasCapacity()
		cap := ps.Capacity()
		level := ps.CapacityLevel()
		status := ps.Status()

		statusStr := lipgloss.NewStyle().
			Foreground(statusColor(status)).
			Render(status)

		var titleRightParts []string
		titleRightParts = append(titleRightParts, statusStr)

		if hasCapacity {
			capStr := lipgloss.NewStyle().
				Foreground(capacityColor(cap)).
				Bold(true).
				Render(fmt.Sprintf("%.0f%%", cap))
			titleRightParts = append(titleRightParts, capStr)
		}

		if level != "" {
			var levelStr string
			switch strings.ToLower(level) {
			case "critical":
				levelStr = criticalStyle.Render("! " + level)
			case "low":
				levelStr = warnStyle.Render("~ " + level)
			case "normal":
				levelStr = accentStyle.Render("● " + level)
			case "full":
				levelStr = accentStyle.Render("● " + level)
			default:
				levelStr = valueStyle.Render(level)
			}
			titleRightParts = append(titleRightParts, levelStr)
		}

		titleRight := lipgloss.JoinHorizontal(lipgloss.Center,
			titleRightParts[0],
		)
		for _, p := range titleRightParts[1:] {
			titleRight = lipgloss.JoinHorizontal(lipgloss.Center,
				titleRight,
				dimStyle.Render(" | "),
				p,
			)
		}

		titleLine := lipgloss.JoinHorizontal(lipgloss.Left, titleLeft,
			lipgloss.NewStyle().Width(max(0, width-lipgloss.Width(titleLeft)-lipgloss.Width(titleRight)-6)).Render(""),
			titleRight,
		)
		bodyLines = append(bodyLines, titleLine)

		if hasCapacity {
			gaugeWidth := width - 8
			gauge := RenderBar(cap, gaugeWidth)
			bodyLines = append(bodyLines, " "+gauge)
		}

		var stats []string
		en := ps.EnergyNowWh()
		ef := ps.EnergyFullWh()
		if en > 0 && ef > 0 {
			stats = append(stats, fmt.Sprintf("Energy: %s / %s", accentStyle.Render(fmtWh(en)), dimStyle.Render(fmtWh(ef))))
		}
		pw := ps.PowerNowW()
		if pw > 0 {
			stats = append(stats, fmt.Sprintf("Power: %s", valueStyle.Render(fmtW(pw))))
		}
		wear := ps.WearLevel()
		if wear > 0 {
			var wearStr string
			if wear < 10 {
				wearStr = accentStyle.Render(fmtPct(wear))
			} else if wear < 30 {
				wearStr = warnStyle.Render(fmtPct(wear))
			} else {
				wearStr = criticalStyle.Render(fmtPct(wear))
			}
			stats = append(stats, fmt.Sprintf("Wear: %s", wearStr))
		}
		v := ps.VoltageNowV()
		if v > 0 {
			stats = append(stats, fmt.Sprintf("Voltage: %s", valueStyle.Render(fmtV(v))))
		}
		cc := ps.CycleCount()
		if cc > 0 {
			stats = append(stats, fmt.Sprintf("Cycles: %s", valueStyle.Render(fmt.Sprintf("%d", cc))))
		}

		timeToEmpty := ps.TimeToEmpty()
		timeToFull := ps.TimeToFull()
		if timeToFull != "" {
			stats = append(stats, fmt.Sprintf("→ Full: %s", chargingStyle.Render(timeToFull)))
		}
		if timeToEmpty != "" {
			stats = append(stats, fmt.Sprintf("→ Empty: %s", warnStyle.Render(timeToEmpty)))
		}

		if ps.IsCharging() && hasCapacity {
			stats = append(stats, fmt.Sprintf("Charge: %s", chargingStyle.Render("active")))
		}

		if len(stats) > 0 {
			bodyLines = append(bodyLines, "")
			for i := 0; i < len(stats); i += 2 {
				if i+1 < len(stats) {
					line := stats[i] + dimStyle.Render("  │  ") + stats[i+1]
					bodyLines = append(bodyLines, " "+line)
				} else {
					bodyLines = append(bodyLines, " "+stats[i])
				}
			}
		}

		manuf := ps.Manufacturer()
		modelName := ps.ModelName()
		if manuf != "" || modelName != "" {
			info := ""
			if manuf != "" {
				info += manuf
			}
			if modelName != "" {
				if info != "" {
					info += " "
				}
				info += modelName
			}
			bodyLines = append(bodyLines, "")
			bodyLines = append(bodyLines, " "+dimStyle.Render(info))
		}

	} else {
		titleLine := titleLeft

		var statusItems []string
		if ps.Type == "Mains" {
			online := ps.IsOnline()
			if online {
				statusItems = append(statusItems, accentStyle.Render("Online"))
			} else {
				statusItems = append(statusItems, criticalStyle.Render("Offline"))
			}
		}

		v := ps.VoltageNowV()
		if v > 0 {
			statusItems = append(statusItems, fmtV(v))
		}
		a := ps.CurrentNowA()
		if a > 0 {
			statusItems = append(statusItems, fmtA(a))
		}

		scope := ps.Prop("POWER_SUPPLY_SCOPE")
		if scope != "" {
			statusItems = append(statusItems, scope)
		}

		manuf := ps.Manufacturer()
		model := ps.ModelName()
		if manuf != "" && model != "" {
			statusItems = append(statusItems, manuf+" "+model)
		} else if manuf != "" {
			statusItems = append(statusItems, manuf)
		} else if model != "" {
			statusItems = append(statusItems, model)
		}

		right := lipgloss.JoinHorizontal(lipgloss.Center,
			dimStyle.Render(strings.Join(statusItems, " | ")),
		)

		titleLine = lipgloss.JoinHorizontal(lipgloss.Left, titleLeft,
			lipgloss.NewStyle().Width(max(0, width-lipgloss.Width(titleLeft)-lipgloss.Width(right)-6)).Render(""),
			right,
		)
		bodyLines = append(bodyLines, titleLine)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, bodyLines...)
	return card.Render(body)
}

func (m model) renderDetails() string {
	if len(m.supplies) == 0 {
		return m.renderEmpty()
	}
	ps := m.supplies[m.selected]

	var sections []string

	title := fmt.Sprintf("▸ %s", cardTitleStyle.Render(ps.Name))
	if ps.Type != "" {
		title += dimStyle.Render(" (" + ps.Type + ")")
	}
	num := fmt.Sprintf(" %d/%d", m.selected+1, len(m.supplies))
	title += lipgloss.NewStyle().Width(max(0, m.width-lipgloss.Width(title)-lipgloss.Width(dimStyle.Render(num))-10)).Render("")
	title += dimStyle.Render(num)
	sections = append(sections, title)
	sections = append(sections, "")

	if ps.IsBattery() {
		statusEntries := []string{
			m.kvColor("Status", ps.Status(), statusColor(ps.Status())),
		}
		if ps.HasCapacity() {
			statusEntries = append(statusEntries,
				m.kvColor("Capacity", fmt.Sprintf("%.0f%%", ps.Capacity()), capacityColor(ps.Capacity())),
			)
		}
		statusEntries = append(statusEntries,
			m.kv("Level", ps.CapacityLevel()),
			m.kv("Present", boolStr(ps.Prop("POWER_SUPPLY_PRESENT") == "1")),
			m.kv("Technology", ps.Technology()),
		)
		sections = append(sections, m.section("Status", statusEntries...)...)
		sections = append(sections, "")

		en := ps.EnergyNowWh()
		ef := ps.EnergyFullWh()
		ed := ps.EnergyFullDesignWh()
		pw := ps.PowerNowW()
		if en > 0 || pw > 0 {
			var entries []string
			if en > 0 {
				entries = append(entries, m.kv("Energy Now", fmtWh(en)))
			}
			if ef > 0 {
				entries = append(entries, m.kv("Energy Full", fmtWh(ef)))
			}
			if ed > 0 {
				entries = append(entries, m.kv("Energy Design", fmtWh(ed)))
			}
			if pw > 0 {
				entries = append(entries, m.kv("Power Now", fmtW(pw)))
			}
			wear := ps.WearLevel()
			if wear > 0 {
				entries = append(entries, m.kv("Wear Level", fmtPct(wear)))
			}
			sections = append(sections, m.section("Energy & Power", entries...)...)
			sections = append(sections, "")
		}

		cn := ps.ChargeNowAh()
		cf := ps.ChargeFullAh()
		cd := ps.ChargeFullDesignAh()
		if cn > 0 || cf > 0 {
			var entries []string
			if cn > 0 {
				entries = append(entries, m.kv("Charge Now", fmtAh(cn)))
			}
			if cf > 0 {
				entries = append(entries, m.kv("Charge Full", fmtAh(cf)))
			}
			if cd > 0 {
				entries = append(entries, m.kv("Charge Design", fmtAh(cd)))
			}
			sections = append(sections, m.section("Charge", entries...)...)
			sections = append(sections, "")
		}

		vn := ps.VoltageNowV()
		vm := ps.VoltageMinDesignV()
		vx := ps.VoltageMaxDesignV()
		if vn > 0 || vm > 0 {
			var entries []string
			if vn > 0 {
				entries = append(entries, m.kv("Voltage Now", fmtV(vn)))
			}
			if vm > 0 {
				entries = append(entries, m.kv("Voltage Min Design", fmtV(vm)))
			}
			if vx > 0 {
				entries = append(entries, m.kv("Voltage Max Design", fmtV(vx)))
			}
			sections = append(sections, m.section("Voltage", entries...)...)
			sections = append(sections, "")
		}

		cc := ps.CycleCount()
		temp := ps.Temperature()
		if cc > 0 || temp != "" {
			var entries []string
			if cc > 0 {
				entries = append(entries, m.kv("Cycle Count", fmt.Sprintf("%d", cc)))
			}
			if temp != "" {
				entries = append(entries, m.kv("Temperature", temp))
			}
			alarmStr := ps.Prop("POWER_SUPPLY_ALARM")
			alarmV, _ := ps.IntProp("POWER_SUPPLY_ALARM")
			if alarmV > 0 {
				entries = append(entries, m.kv("Alarm Threshold", fmt.Sprintf("%d µWh", alarmV)))
			} else if alarmStr != "" {
				entries = append(entries, m.kv("Alarm", alarmStr))
			}
			ct := ps.Prop("POWER_SUPPLY_CHARGE_CONTROL_END_THRESHOLD")
			if ct != "" {
				entries = append(entries, m.kv("Charge Threshold", ct+"%"))
			}
			sections = append(sections, m.section("Health & Cycles", entries...)...)
			sections = append(sections, "")
		}

		timeToEmpty := ps.TimeToEmpty()
		timeToFull := ps.TimeToFull()
		if timeToEmpty != "" || timeToFull != "" {
			var entries []string
			if timeToFull != "" {
				entries = append(entries, m.kv("Time to Full", timeToFull))
			}
			if timeToEmpty != "" {
				entries = append(entries, m.kv("Time to Empty", timeToEmpty))
			}
			sections = append(sections, m.section("Time Remaining", entries...)...)
			sections = append(sections, "")
		}
	}

	var deviceInfo []string
	manuf := ps.Manufacturer()
	if manuf != "" {
		deviceInfo = append(deviceInfo, m.kv("Manufacturer", manuf))
	}
	model := ps.ModelName()
	if model != "" {
		deviceInfo = append(deviceInfo, m.kv("Model", model))
	}
	sn := ps.SerialNumber()
	if sn != "" {
		deviceInfo = append(deviceInfo, m.kv("Serial", sn))
	}
	tech := ps.Technology()
	if tech != "" {
		deviceInfo = append(deviceInfo, m.kv("Technology", tech))
	}
	scope := ps.Prop("POWER_SUPPLY_SCOPE")
	if scope != "" {
		deviceInfo = append(deviceInfo, m.kv("Scope", scope))
	}
	usbType := ps.Prop("POWER_SUPPLY_USB_TYPE")
	if usbType != "" {
		deviceInfo = append(deviceInfo, m.kv("USB Type", usbType))
	}
	if len(deviceInfo) > 0 {
		sections = append(sections, m.section("Device Info", deviceInfo...)...)
	}

	body := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Wrap in a card
	width := max(40, m.width-6)
	if width > 100 {
		width = 100
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(0, 2).
		Width(width).
		Render(body)
}

func (m model) section(title string, entries ...string) []string {
	var lines []string
	lines = append(lines, "  "+sectionTitleStyle.Render(title))
	for _, e := range entries {
		lines = append(lines, "    "+e)
	}
	return lines
}

func (m model) kv(key, value string) string {
	return m.kvColor(key, value, lipgloss.Color(""))
}

func (m model) kvColor(key, value string, color lipgloss.Color) string {
	label := labelStyle.Render(key + ":")
	pad := strings.Repeat(" ", max(0, 22-len(key)-1))

	var valStr string
	if color != "" && color != lipgloss.Color("") {
		valStr = lipgloss.NewStyle().Foreground(color).Render(value)
	} else {
		valStr = valueStyle.Render(value)
	}

	return label + pad + valStr
}

func (m model) renderRaw() string {
	if len(m.supplies) == 0 {
		return m.renderEmpty()
	}
	ps := m.supplies[m.selected]

	title := fmt.Sprintf("▸ %s", cardTitleStyle.Render(ps.Name))
	if ps.Type != "" {
		title += dimStyle.Render(" (" + ps.Type + ")")
	}
	num := fmt.Sprintf(" %d/%d", m.selected+1, len(m.supplies))
	title += lipgloss.NewStyle().Width(max(0, m.width-lipgloss.Width(title)-lipgloss.Width(dimStyle.Render(num))-10)).Render("")
	title += dimStyle.Render(num)

	var lines []string
	lines = append(lines, title)
	lines = append(lines, "")

	keys := SortedProps(ps.Properties)
	maxKeyLen := 0
	for _, k := range keys {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}
	if maxKeyLen > 40 {
		maxKeyLen = 40
	}

	for _, k := range keys {
		v := ps.Properties[k]
		kStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#7dcfff")).Render(k)
		pad := strings.Repeat(" ", max(0, maxKeyLen-len(k)+2))

		var vStr string
		switch {
		case strings.HasSuffix(k, "STATUS"):
			vStr = lipgloss.NewStyle().Foreground(statusColor(v)).Render(v)
		case strings.HasSuffix(k, "CAPACITY"):
			vStr = lipgloss.NewStyle().Foreground(capacityColor(parseFloatSafe(v))).Render(v)
		case strings.HasSuffix(k, "ONLINE"):
			if v == "1" {
				vStr = accentStyle.Render("true")
			} else {
				vStr = criticalStyle.Render("false")
			}
		case strings.HasSuffix(k, "PRESENT"):
			if v == "1" {
				vStr = accentStyle.Render("true")
			} else {
				vStr = criticalStyle.Render("false")
			}
		default:
			vStr = valueStyle.Render(v)
		}

		lines = append(lines, "  "+kStr+pad+vStr)
	}

	width := max(40, m.width-6)
	if width > 120 {
		width = 120
	}

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3b4261")).
		Padding(0, 2).
		Width(width).
		Render(body)
}

func boolStr(v bool) string {
	if v {
		return accentStyle.Render("Yes")
	}
	return criticalStyle.Render("No")
}

func parseFloatSafe(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
