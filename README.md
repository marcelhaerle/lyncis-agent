# Lyncis Agent

Lyncis Agent is a lightweight, statically compiled Go binary that runs on target systems to execute security audits using [Lynis](https://cisofy.com/lynis/). It communicates with the centralized `lyncis-backend` to register itself, poll for tasks, run security scans, and report findings back for visualization.

## Features

- Automatic self-registration with the backend.
- Token-based heartbeat and task polling system.
- Leverages Lynis to run local environment security audits (`lynis audit system`).
- Parses scan reports to extract warnings and suggestions.
- Reports detailed findings back to the Lyncis platform.

## Development Setup

### Prerequisites

- [Go](https://go.dev/doc/install) (1.20 or later recommended)
- [Lynis](https://cisofy.com/download/lynis/) installed on the development/testing machine.
- A running instance of the `lyncis-backend`.

### Building the Agent

Clone this repository and build the binary:

```bash
# Clone the repository
git clone https://github.com/marcelhaerle/lyncis-agent.git
cd lyncis-agent

# Download dependencies
go mod download

# Build the executable
go build -o lyncis-agent main.go
```

### Running Locally

To run the agent locally against your development backend:

1. Ensure the `lyncis-backend` is reachable and running.
2. Execute the compiled agent:

```bash
./lyncis-agent
```

The agent will attempt to register itself with the backend, save its local token (development mode uses the current directory instead of `/etc/lyncis/config.json`), and start polling for new tasks.

## Deployment Setup

### Download using GitHub Releases

The recommended way to deploy the `lyncis-agent` to target systems is by downloading the statically compiled binaries from the [GitHub Releases](https://github.com/marcelhaerle/lyncis-agent/releases) page. We provide binaries for Linux (`amd64` and `arm64`).

1: Download the latest binary for your architecture:

```bash
# Example for Linux AMD64
wget https://github.com/marcelhaerle/lyncis-agent/releases/latest/download/lyncis-agent-linux-amd64
chmod +x lyncis-agent-linux-amd64
sudo mv lyncis-agent-linux-amd64 /usr/local/bin/lyncis-agent
```

2: Ensure [Lynis](https://cisofy.com/download/lynis/) is installed on the target machine as the agent relies on it to perform security audits.

### Configuration

The agent requires configuration to know where the central backend is located. By default, it looks for its configuration file at `/etc/lyncis/config.json`.

Create the configuration directory:

```bash
sudo mkdir -p /etc/lyncis
```

### Running as a Systemd Service

To keep the agent running continuously and start it on boot, set it up as a systemd service:

1: Create a service file at `/etc/systemd/system/lyncis-agent.service`:

```ini
[Unit]
Description=Lyncis Security Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/lyncis-agent
Restart=always
RestartSec=10
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=lyncis-agent
Environment=LYNCIS_BACKEND_URL=<YOUR_BACKEND_URL>
Environment=LYNCIS_CONFIG_PATH=/etc/lyncis/config.json

[Install]
WantedBy=multi-user.target
```

2: Enable and start the service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable lyncis-agent
sudo systemctl start lyncis-agent
```

3: Check the status and logs:

```bash
sudo systemctl status lyncis-agent
sudo journalctl -u lyncis-agent -f
```

## Related Repositories

- [lyncis-backend](https://github.com/marcelhaerle/lyncis-backend) — Go API backend
- [lyncis-ui](https://github.com/marcelhaerle/lyncis-ui) — React dashboard
