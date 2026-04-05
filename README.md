[![Certified Shovelware](https://justin.searls.co/img/shovelware.svg)](https://justin.searls.co/shovelware/)

# ports

A Go CLI for seeing what is listening on your machine.

It is very much inspired by [port-whisperer](https://github.com/LarsenCundric/port-whisperer), but follows the structure of your other Go CLIs: small `cmd/` files, a scanner package for the shell-driven plumbing, and subprocess-friendly tests.

## Features

- dev-focused port list by default
- `--all` mode for everything listening
- inspect a single port with project path, git branch, memory, uptime, and process tree
- `ps` view for developer processes
- `kill` by port or PID, with `--port` / `--pid` when a number is ambiguous
- `clean` orphaned/zombie dev processes
- `watch` for port changes
- framework detection for common JS, Python, Go, and Docker workloads

## Installation

### From source

```bash
make install
```

### Download a release binary

Pre-built macOS binaries for Apple Silicon and Intel are available on the [Releases](https://github.com/LarsEckart/ports/releases) page.

## Usage

```bash
ports
ports --all
ports <port>
ports ps
ports ps --all
ports kill <port-or-pid>
ports kill --pid <pid>
ports kill --port <port>
ports clean --yes
ports watch
```

## How it works

Like the original idea, it leans on the native tooling already on the machine:

- `lsof -nP -iTCP -sTCP:LISTEN`
- batched `ps` calls for PID metadata
- `lsof -a -d cwd` for working directories
- `docker ps` for host-port to container mapping

Right now this implementation is macOS-first, because the `ps` and `lsof` output parsing matches Darwin.
