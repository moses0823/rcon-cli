# rcontui

A terminal-based RCON client with an interactive TUI and optional Secure RCON authentication.

This project is a fork of [gorcon/rcon-cli](https://github.com/gorcon/rcon-cli), originally released under the MIT License.

This fork adds additional features and modifications while retaining the original project's license and attribution.

## Features

* Interactive terminal UI
* Minecraft RCON console
* Multiple server configurations
* Secure RCON authentication
* HMAC-SHA256 challenge-response authentication
* Separate authentication secret files
* Windows support

## Installation

Download `rcontui.exe` and place it in a dedicated directory:

```text
C:\rcontui\
├── rcontui.exe
├── rcon.yaml
└── secrets\
    └── 6b7t-moses.key
```

## Configuration

Create `rcon.yaml`:

```yaml
servers:
  6b7t:
    address: "abula.tw:25576"

    security:
      enabled: true
      client-id: "moses"
      secret-file: "secrets/6b7t-moses.key"
```

### Configuration Options

| Option                 | Description                        |
| ---------------------- | ---------------------------------- |
| `address`              | RCON or Secure RCON server address |
| `security.enabled`     | Enable Secure RCON                 |
| `security.client-id`   | Client identifier                  |
| `security.secret-file` | Path to the authentication secret  |

## Secure RCON

When enabled, `rcontui` connects to the Secure RCON Gateway instead of directly connecting to Minecraft's native RCON.

```text
rcontui
   │
   │ Secure authentication
   ▼
Secure RCON Gateway
   │
   ▼
Minecraft Native RCON
```

Example:

```yaml
address: "abula.tw:25576"

security:
  enabled: true
  client-id: "moses"
  secret-file: "secrets/6b7t-moses.key"
```

The Minecraft native RCON port should remain private and should not be exposed directly to the Internet.

## Secret File

The secret is stored separately:

```text
secrets\
└── 6b7t-moses.key
```

Do not commit secret files to Git.

Do not share them publicly.

## Running

Open PowerShell:

```powershell
cd C:\rcontui
.\rcontui.exe
```

Use the keyboard to navigate:

```text
↑ / ↓     Select server
Enter     Connect
Q         Quit
```

## Multiple Servers

Multiple servers can be configured:

```yaml
servers:
  6b7t:
    address: "abula.tw:25576"
    security:
      enabled: true
      client-id: "moses"
      secret-file: "secrets/6b7t-moses.key"

  local:
    address: "127.0.0.1:25575"
    security:
      enabled: false
```

Each server can independently enable or disable Secure RCON.

## Security

Secure RCON uses HMAC-SHA256 challenge-response authentication.

The authentication process uses:

* Client ID
* Server-generated nonce
* Time-based authentication
* Shared secret
* HMAC-SHA256

The shared secret is not transmitted directly during authentication.

## Building

This project requires Go.

Run:

```powershell
go test ./...
```

Build the Windows executable:

```powershell
go build -o rcontui.exe .
```

## Fork Information

This repository is based on:

**Original project:** `gorcon/rcon-cli`

**Original repository:** https://github.com/gorcon/rcon-cli

The original project is licensed under the MIT License.

This fork retains the original copyright notices and license.

## License

This project is distributed under the MIT License.

See [`LICENSE`](LICENSE) for the full license text.

The original project and its respective copyright notices remain subject to their original license terms.
