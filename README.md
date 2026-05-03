<p align="center">
  <img src="utils/docklog.png" alt="Docklog Mascot" width="400">
</p>


**Docklog** is a optimized, multiplexed, real-time log aggregator and tracking CLI for Docker. 

Designed for power-users, SREs, and backend developers, Docklog solves the problem of "log blindness" by allowing you to attach to multiple containers simultaneously, filter noise at the source, deduplicate spam, and format outputs for both human readability and machine ingestion.

---

## Table of Contents

- [Why Docklog?](#why-docklog)
- [Architecture](#architecture)
- [Installation](#installation)
- [Getting Started](#getting-started)
- [Advanced Usage](#advanced-usage)
  - [Container Targeting](#container-targeting)
  - [Regex & Exclusion Filtering](#regex--exclusion-filtering)
  - [JSON & File Output](#json--file-output)
  - [Spam Deduplication (Smell Error)](#spam-deduplication-smell-error)
- [Configuration](#configuration)
- [Contributing](#contributing)

---

## Why Docklog?

Standard Docker commands (`docker logs -f`) are limited when dealing with complex, multi-container environments. Third-party UI tools can be slow or bloated. Docklog provides a surgical CLI approach:

1. **Multiplexed Streaming:** Watch dozens of containers in a single terminal window, color-coded automatically.
2. **Zero-Overhead Filtering:** Drop noisy logs (e.g., health checks) before they ever reach your terminal.
3. **Smart Deduplication:** Prevent aggressive error loops from burying the root cause with consecutive message suppression.
4. **Structured Output:** Stream directly to JSON files for immediate ingestion by Splunk, ELK, or Datadog agents.

---

## Architecture

Docklog interfaces directly with the local Docker Engine API via the Unix socket (`/var/run/docker.sock` or Windows Pipe). 

It dynamically discovers running containers and listens for Docker Engine `start` and `die` events to attach or detach streams automatically—meaning you can start Docklog, and it will automatically pick up containers that are started later.

The internal **Aggregator** utilizes concurrent `io.Pipe` streams, processing Stdout and Stderr asynchronously through a centralized formatting and filtering pipeline, ensuring zero race conditions and thread-safe outputs.

---

## Installation

Ensure you have [Go 1.21+](https://go.dev/) installed.

```bash
# Clone the repository
git clone https://github.com/yourusername/docklog.git
cd docklog

# Install the binary
go install
```

Ensure your `$GOPATH/bin` is in your system's `$PATH`.

---

## Getting Started

To start streaming logs from all currently running containers:

```bash
docklog start
```

Press `Ctrl+C` to gracefully terminate the streams and exit.

---

## Advanced Usage

Docklog provides POSIX-compliant flags for granular control over your log streams.

### Container Targeting

Listen only to containers whose names match a specific regular expression. This prevents Docklog from attaching to unrelated containers, saving CPU and memory.

```bash
# Listen only to containers starting with "api-" or "db-"
docklog start --container "^(api|db)-.*"
```

### Regex & Exclusion Filtering

Find exactly what you are looking for, and ignore the noise.

```bash
# Only show logs containing the word "timeout"
docklog start --filter "timeout"

# Show all logs, EXCEPT those containing "healthcheck"
docklog start --exclude "healthcheck"

# Advanced Regex: Find panic traces or specific ID formats
docklog start --regex "panic: runtime error|req_id=[A-Z0-9]+"
```

### JSON & File Output

Export your aggregated logs to a file, either in raw text or structured JSON.

```bash
# Output colored logs to terminal, and raw logs to logs.txt
docklog start --output logs.txt

# Output structured JSON (disables terminal colors)
docklog start --json --output logs.json
```

### Masking / Censoring (Redact)

Automatically mask sensitive data (Emails, IPv4, Bearer Tokens, API Keys) in your logs. When enabled, matching patterns are replaced with `***` before being printed or written to a file.

```bash
docklog start --redact
```

> **Eğer ki AI ile çalışıyorsanız ona hatanızı yapıştırmadan önce şifrelerinizi korumak için önerilir.**


### Spam Deduplication (Smell Error)

Docklog comes with a specialized command specifically for debugging critical failures without being overwhelmed by recursive error spam.

```bash
docklog start smell-error
```

This command enforces an `"error"` filter and enables the `--dedupe` flag. If a container spams the exact same error 10,000 times a second, Docklog will only print it **once**, until a different log is emitted.

You can also target specific broken containers:
```bash
docklog start smell-error --container "broken-postgres"
```

---

## Configuration

Docklog is built with Viper, meaning it supports configuration files to save your favorite parameters.

Create a `.docklog.yaml` file in your home directory (`~/.docklog.yaml`) or in your current working directory:

```yaml
# ~/.docklog.yaml
container: "^(core|auth)-.*"
exclude: "DEBUG"
tail: "50"
json: false
```

Flags passed via the CLI will automatically override the values specified in the configuration file.

---

## Contributing

We welcome contributions from the community! 

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Ensure tests pass (`go test -v ./...`)
4. Commit your changes (`git commit -m 'feat: add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
