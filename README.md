# battui

A terminal UI for every piece of battery info the Linux kernel can expose.

Reads all devices from `/sys/class/power_supply/` — batteries, AC adapters, USB
ports, peripheral batteries — and surfaces every property in a real-time
terminal dashboard.

## Features

- **Every property, every device** — energy, power, voltage, current, charge,
  temperature, cycle count, wear level, manufacturer, model, serial, technology,
  alarm thresholds, charge thresholds, time estimates, and every raw uevent key
- **Three views** — Overview (cards with gauges), Details (structured sections),
  All Properties (raw uevent dump)
- **Live polling** — refreshes every 2 seconds
- **Color-coded** — gauges change from red to yellow to green, status and level
  indicators use semantic colours
- **Wear tracking** — battery health degradation displayed prominently
- **Multiple device types** — internal batteries, wireless peripherals,
  AC adapters, USB-C power sources
- **Scrolling** — mouse wheel, Page Up/Down, and arrow keys
- **Zero configuration** — reads the kernel's sysfs directly

## Quick start

```bash
make build
./battui
```

### Keybindings

| Key | Action |
|-----|--------|
| `q` / `Esc` / `Ctrl+C` | Quit |
| `Tab` / `h` / `l` | Cycle tabs |
| `1` / `2` / `3` | Jump to tab |
| `UP` / `DOWN` / `j` / `k` | Select device (Overview) / scroll |
| `g` / `G` | First / last device |
| `r` | Force refresh |
| Mouse wheel / `PgUp` / `PgDn` | Scroll content |

### Tabs

1. **Overview** — one card per power supply with gauge bar, status, key metrics
2. **Details** — structured view of all properties for the selected device
3. **All Props** — raw key-value dump from uevent, coloured by type

## Install

### From source

```bash
git clone git@github.com:ReubenPercival/battui.git
cd battui
make install
```

Or without make:

```bash
go build -ldflags="-s -w" -o battui .
sudo cp battui /usr/local/bin/
```

### Make targets

| Target | Description |
|--------|-------------|
| `build` | Build the binary |
| `install` | Build and copy to `/usr/local/bin` |
| `clean` | Remove built binary |
| `run` | Build and run |
| `test` | Run tests |
| `vet` | Run `go vet` |
| `lint` | Run `staticcheck` (if installed) |

### Dependencies

- Go 1.26+
- Linux with `/sys/class/power_supply/`
- Terminal with Unicode and true colour support

## Requirements

- **Linux only** — reads kernel sysfs, does not work on macOS or BSD
- **Terminal** — needs Unicode support for gauge characters

## License

European Union Public Licence v. 1.2 — see [LICENSE](./LICENSE).
