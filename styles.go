package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	headerBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1b26"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a9b1d6")).
			Background(lipgloss.Color("#1a1b26")).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(lipgloss.Color("#7aa2f7")).
			Padding(0, 2).
			MarginRight(1)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a9b1d6")).
				Background(lipgloss.Color("#24283b")).
				Padding(0, 2).
				MarginRight(1)

	cardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7aa2f7"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565f89"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#c0caf5"))

	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ece6a"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e0af68"))

	criticalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f7768e"))

	chargingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#bb9af7"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b4261"))

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7dcfff")).
				Underline(true).
				MarginBottom(1)
)

func capacityColor(pct float64) lipgloss.Color {
	switch {
	case pct < 10:
		return lipgloss.Color("#f7768e")
	case pct < 20:
		return lipgloss.Color("#ff9e64")
	case pct < 40:
		return lipgloss.Color("#e0af68")
	default:
		return lipgloss.Color("#9ece6a")
	}
}

func statusColor(status string) lipgloss.Color {
	switch strings.ToLower(status) {
	case "charging":
		return lipgloss.Color("#bb9af7")
	case "discharging":
		return lipgloss.Color("#e0af68")
	case "full":
		return lipgloss.Color("#9ece6a")
	case "not charging":
		return lipgloss.Color("#f7768e")
	default:
		return lipgloss.Color("#565f89")
	}
}

func RenderBar(pct float64, width int) string {
	if width < 2 {
		return ""
	}
	pct = math.Max(0, math.Min(100, pct))
	filled := int(math.Round(float64(width) * pct / 100))
	if filled > width {
		filled = width
	}
	empty := width - filled

	color := capacityColor(pct)

	fillChar := "━"
	emptyChar := "─"

	filledStr := lipgloss.NewStyle().
		Foreground(color).
		Background(lipgloss.Color("#1a1b26")).
		Render(strings.Repeat(fillChar, filled))

	emptyStr := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3b4261")).
		Render(strings.Repeat(emptyChar, empty))

	return filledStr + emptyStr
}

func fmtW(v float64) string {
	switch {
	case v >= 1_000_000:
		return fmt.Sprintf("%.2f MW", v/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.2f kW", v/1_000)
	case v >= 1:
		return fmt.Sprintf("%.2f W", v)
	case v >= 0.001:
		return fmt.Sprintf("%.1f mW", v*1000)
	default:
		return fmt.Sprintf("%.2f W", v)
	}
}

func fmtWh(v float64) string {
	switch {
	case v >= 1_000_000:
		return fmt.Sprintf("%.2f MWh", v/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.2f kWh", v/1_000)
	default:
		return fmt.Sprintf("%.1f Wh", v)
	}
}

func fmtV(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%.3f V", v)
}

func fmtA(v float64) string {
	if v == 0 {
		return ""
	}
	switch {
	case v >= 1:
		return fmt.Sprintf("%.2f A", v)
	case v >= 0.001:
		return fmt.Sprintf("%.1f mA", v*1000)
	default:
		return fmt.Sprintf("%.3f A", v)
	}
}

func fmtAh(v float64) string {
	if v == 0 {
		return ""
	}
	switch {
	case v >= 1:
		return fmt.Sprintf("%.2f Ah", v)
	case v >= 0.001:
		return fmt.Sprintf("%.0f mAh", v*1000)
	default:
		return fmt.Sprintf("%.2f Ah", v)
	}
}

func fmtPct(v float64) string {
	return fmt.Sprintf("%.1f%%", v)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "..."
}
