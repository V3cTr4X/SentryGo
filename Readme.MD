# SentryGo – Simple yet powerful infrastructure monitoring

SentryGo is a lightweight, cross-platform monitoring solution designed for small to medium-sized environments. It follows a classic client-server architecture: a central server collects metrics from agents installed on monitored hosts, displays them on a clean dashboard, and sends alerts when thresholds are exceeded.

Built with **Go**, SentryGo compiles into a single binary with no external dependencies, making deployment trivial on **Linux**, **Windows**, and **macOS**.

## Key Features

- **Agent-based monitoring** – Deploy a tiny agent on any host.
- **Cross-platform** – Agents and server run on Linux, Windows, and macOS.
- **Minimal resource usage** – Written in Go, compiled to native binaries.
- **Built-in dashboard** – Web UI with real-time graphs and host overview.
- **Alerting** – Email or webhook notifications when metrics exceed thresholds.
- **No external database** – Uses SQLite embedded; zero configuration.
- **Easy management** – Single binary for server, single binary for agent.

## Architecture

- **Server** – REST API + dashboard. Stores metrics in SQLite, serves web UI.
- **Agent** – Collects CPU, memory, disk, and network latency, sends them to the server via HTTP.
- **Authentication** – Each agent uses a unique token for secure communication.

## Use Cases

- Monitor on-premise servers, VMs, or cloud instances.
- Keep an eye on Raspberry Pi clusters or home lab.
- Lightweight alternative to Nagios or Zabbix for smaller setups.

## Getting Started

Check the [Installation Guide](docs/install.md) to set up your server and deploy agents in minutes.

## License

SentryGo is open-source under the MIT license. Commercial support and Pro features (unlimited hosts, advanced alerts) are available upon request.