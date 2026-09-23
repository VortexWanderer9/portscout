# portscout

A fast, dependency-free TCP port scanner written in Go. It uses a worker pool for concurrent connects, can grab service banners, and exports results for terminals, scripts, and spreadsheets.

[![CI](https://github.com/VortexWanderer9/portscout/actions/workflows/ci.yml/badge.svg)](https://github.com/VortexWanderer9/portscout/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> **Only scan hosts you own or have explicit permission to test.**

## Features

- Concurrent scanning with a configurable worker pool
- Flexible port specs: `22`, `22,80,443`, `1-1024`, `all`, `web`, `database`, `mail`, `remote`, or a mix; presets can be combined with explicit ports
- Optional banner grabbing (SSH, SMTP, FTP, etc.)
- Well-known service name detection, including LDAP, Docker, Kubernetes, and Memcached
- Table, JSON, CSV, or port-only output
- Graceful Ctrl+C handling with partial results
- Standard library only, no external dependencies

## Install

```sh
go install github.com/VortexWanderer9/portscout/cmd/portscout@latest
```

Or build from source:

```sh
git clone https://github.com/VortexWanderer9/portscout.git
cd portscout
make build
./bin/portscout version
```

## Usage

```
portscout scan [flags] <host>
portscout version

  --ports string       ports to scan; -p also works (default "1-1024")
  --exclude string     ports or ranges to skip
  --timeout duration   timeout per connection; -t also works (default 800ms)
  --workers int        concurrent workers; -w also works (default 200)
  --banners            grab service banners; -b also works
  --format string      table, json, csv, or ports (default "table")
  --network string     tcp, tcp4, or tcp6 (default "tcp")
  --help               print help; -h also works
```

The timeout passed to `--timeout` must be greater than zero. Use Go duration syntax,
such as `500ms` or `2s`.

The legacy form, `portscout [flags] <host>`, is still supported. The old
`-json`, `-csv`, and `-quiet` flags remain available as compatibility aliases.

### Examples

```sh
# Scan the default port range on localhost
portscout scan 127.0.0.1

# Scan specific ports with banner grabbing
portscout scan --banners --ports 22,80,443,8080 scanme.nmap.org

# Scan common web-service ports (80, 443, 8080, and 8443)
portscout scan --ports web example.com

# Scan a range while skipping sensitive ports
portscout scan --ports 1-1024 --exclude 22,3389 example.com

# Print open port numbers only, one per line
portscout scan --format ports --ports 1-1024 127.0.0.1

# Send a CSV report to a file
portscout scan --format csv --ports web example.com > report.csv

# Scan every port, faster, JSON output piped to jq
portscout scan --format json --ports all --workers 500 192.168.1.10 | jq '.open[].port'
```

Example output:

```
Scan report for 127.0.0.1
PORT      STATE  SERVICE     LATENCY  BANNER
22/tcp    open   ssh         112µs    SSH-2.0-OpenSSH_9.6
5432/tcp  open   postgresql  98µs

2 open, 1024 scanned in 143ms
```

## Project layout

```
cmd/portscout/        CLI entry point
internal/ports/       port spec parsing
internal/scanner/     concurrent scanning engine
internal/report/      table and JSON output
```

## Development

```sh
make test    # run tests with the race detector
make vet     # go vet
make check   # run tests and static checks
make coverage # print test coverage by function
make fmt     # gofmt
make build   # build to ./bin/portscout
```

## License

[MIT](LICENSE)
