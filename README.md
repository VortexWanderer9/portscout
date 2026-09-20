# portscout

A fast, dependency-free TCP port scanner written in Go. It uses a worker pool for concurrent connects, can grab service banners, and outputs either a clean table or JSON for scripting.

[![CI](https://github.com/VortexWanderer9/portscout/actions/workflows/ci.yml/badge.svg)](https://github.com/VortexWanderer9/portscout/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> **Only scan hosts you own or have explicit permission to test.**

## Features

- Concurrent scanning with a configurable worker pool
- Flexible port specs: `22`, `22,80,443`, `1-1024`, `all`, `web`, or a mix
- Optional banner grabbing (SSH, SMTP, FTP, etc.)
- Well-known service name detection, including LDAP, Docker, Kubernetes, and Memcached
- Table or JSON output
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
./bin/portscout -version
```

## Usage

```
portscout [flags] <host>

  -p string     ports to scan (default "1-1024")
  -t duration   timeout per connection (default 800ms)
  -w int        number of concurrent workers (default 200)
  -b            grab service banners from open ports
  -json         output results as JSON
  -quiet        output only open port numbers
  -version      print version and exit
```

The timeout passed to `-t` must be greater than zero. Use Go duration syntax,
such as `500ms` or `2s`.

### Examples

```sh
# Scan the default port range on localhost
portscout 127.0.0.1

# Scan specific ports with banner grabbing
portscout -b -p 22,80,443,8080 scanme.nmap.org

# Scan common web-service ports (80, 443, 8080, and 8443)
portscout -p web example.com

# Print open port numbers only, one per line
portscout -quiet -p 1-1024 127.0.0.1

# Scan every port, faster, JSON output piped to jq
portscout -p all -w 500 -json 192.168.1.10 | jq '.open[].port'
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
make fmt     # gofmt
make build   # build to ./bin/portscout
```

## License

[MIT](LICENSE)
