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

*(Note: `main.go` and `go.mod` will be created as part of the initial implementation steps.)*

### Running Locally

To run the agent locally against your development backend:

1. Ensure the `lyncis-backend` is reachable and running.
2. Execute the compiled agent:

```bash
./lyncis-agent
```

The agent will attempt to register itself with the backend, save its local token (development mode uses the current directory instead of `/etc/lyncis/config.json`), and start polling for new tasks.

## Related Repositories

- [lyncis-backend](https://github.com/marcelhaerle/lyncis-backend) — Go API backend
- [lyncis-ui](https://github.com/marcelhaerle/lyncis-ui) — React dashboard
