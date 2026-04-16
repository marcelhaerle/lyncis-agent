# Lyncis Agent

The Lyncis Agent is a lightweight, statically compiled Go binary designed for automated security auditing using [Lynis](https://cisofy.com/lynis/). It acts as the "sensor" for the Lyncis security ecosystem, reporting system health and security findings to a centralized backend.

## Architecture & Workflow

1. **Registration:** Upon first run, the agent registers itself with the `lyncis-backend` using its hostname and OS info.
2. **Persistence:** It securely stores its unique `AgentID` and `Token` in a configuration file (default: `/etc/lyncis/config.json`).
3. **Heartbeat & Task Polling:** The agent maintains a persistent polling loop, checking the backend for tasks (like `run_lynis`).
4. **Execution:** When commanded, it triggers a local Lynis audit, captures the output, and sends the structured results back to the backend.

## Features

- **Automated Audit:** Triggers full-system security scans on demand.
- **Stateless/Lightweight:** Zero runtime dependencies beyond the Lynis binary.
- **Secure Communication:** Token-based authentication per agent.
- **Resilient:** Built-in backoff for connection failures to ensure logging stays clean.

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/doc/install) (1.20+)
- [Lynis](https://cisofy.com/download/lynis/) must be pre-installed on all target nodes.

### Build from Source

```bash
git clone https://github.com/marcelhaerle/lyncis-agent.git
cd lyncis-agent
go mod download
go build -o lyncis-agent main.go
```

---

## Deployment Guide

### 1. Installation
We recommend distributing the binary via your preferred configuration management tool (e.g., **Ansible**).

```bash
# Example deployment via shell/Ansible
wget https://github.com/marcelhaerle/lyncis-agent/releases/latest/download/lyncis-agent-linux-amd64
chmod +x lyncis-agent-linux-amd64
sudo mv lyncis-agent-linux-amd64 /usr/local/bin/lyncis-agent
```

### 2. Configuration
The agent needs the environment variables defined to communicate with your backend.

| Variable | Description | Default |
| :--- | :--- | :--- |
| `LYNCIS_BACKEND_URL` | Full URL of the backend API | Required |
| `LYNCIS_CONFIG_PATH` | Path for agent ID/token storage | `/etc/lyncis/config.json` |

### 3. Service Management (systemd)
Create `/etc/systemd/system/lyncis-agent.service`:

```ini
[Unit]
Description=Lyncis Security Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/lyncis-agent
Restart=always
RestartSec=10
# Inject environment variables
Environment=LYNCIS_BACKEND_URL=https://api.lyncis.yourdomain.com
Environment=LYNCIS_CONFIG_PATH=/etc/lyncis/config.json

[Install]
WantedBy=multi-user.target
```

Apply the changes:
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now lyncis-agent
```

---

## Troubleshooting

- **Check logs:** `journalctl -u lyncis-agent -f`
- **Verify Registration:** Check `/etc/lyncis/config.json` existence.
- **Lynis Dependency:** Ensure the `lynis` command is executable by the user running the agent (ideally `root` for full scan coverage).

## Related Projects

- [lyncis-backend](https://github.com/marcelhaerle/lyncis-backend)
- [lyncis-ui](https://github.com/marcelhaerle/lyncis-ui)
