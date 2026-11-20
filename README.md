# Nova TUI

Nova is a terminal dashboard for monitoring your machine. It shows CPU and
memory usage, top processes, temperatures, disk space, and network traffic.

## Run locally

Run Nova from the project directory in an interactive terminal. It requires
Go 1.26.8 or newer.

```sh
go run .
```

Use ↑ and ↓ to scroll when the dashboard overflows. Press `q` or Ctrl+C to quit.

## How it works

At startup, Nova reads host and CPU details. Background workers use gopsutil to
collect metrics and send updates to the Bubble Tea dashboard. Workers wait one
second between CPU and memory polls, two seconds for network traffic, three
for processes, five for temperature, and ten for disk.

Process CPU usage and network speeds need two samples to calculate a rate.
Disk usage covers the root filesystem, and network totals exclude loopback
interfaces. When a reading fails or gets too old, the dashboard marks the last
successful value as stale.

## License

Nova TUI is licensed under the [MIT License](LICENSE).

## TODO

- [ ] Record and save monitoring sessions to a local file.
- [ ] Replay saved sessions with pause, seek, and playback speed controls.
- [ ] Show historical charts for CPU, memory, disk, and network activity.
- [ ] Let users configure sampling intervals and alert thresholds.
- [ ] Export recorded metrics as CSV for analysis.
