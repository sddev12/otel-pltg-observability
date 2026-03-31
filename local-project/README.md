# Local Observability Stack — Docker Compose

Run Prometheus, Loki, the OTel Collector, and Grafana locally in Docker — no Kubernetes required. All configs live in this `local-project/` directory.

> **Tempo (traces) is not yet configured.** The stack currently covers metrics and logs. Tempo support is TODO.

---

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) with Compose v2 (`docker compose version`)

---

## Directory Structure

```
local-project/
  docker-compose.yaml           # Brings up the full stack
  prometheus/
    prometheus.yml              # Scrape config + remote-write receiver enabled
  loki/
    loki.yaml                   # Single-node, filesystem storage backend
  otel-collector/
    otel-collector-config.yaml  # OTLP receiver → Loki (logs) + Prometheus (metrics)
  grafana/
    grafana.ini                 # Server config, auth, feature toggles
    grafana-12.4.2/             # Upstream Grafana binary distribution (reference only)
                                # Docker Compose uses the grafana/grafana image, not this
    provisioning/
      datasources/
        datasources.yaml        # Auto-provisions Prometheus + Loki datasources
  tempo/                        # TODO — empty, traces not yet configured
```

---

## Starting the Stack

Run from the `local-project/` directory:

```bash
cd local-project
docker compose up
```

To run in the background:

```bash
docker compose up -d
```

---

## Service Endpoints

| Service | URL | Credentials |
|---|---|---|
| Grafana | http://localhost:3000 | `admin` / `admin` |
| Prometheus | http://localhost:9090 | — |
| Loki | http://localhost:3100 | — |
| OTel Collector (gRPC) | `localhost:4317` | — |
| OTel Collector (HTTP) | `localhost:4318` | — |

---

## Sending Telemetry from an App

Point your OTel SDK at the collector before starting your app:

```bash
# gRPC (default for most SDKs)
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# HTTP
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

For the Go Gin API example app:

```bash
cd ../example-apps/go-gin-api
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go
```

Generate traffic:

```bash
./scripts/generate-traffic.sh
```

---

## Viewing Data in Grafana

Open http://localhost:3000. Both datasources are provisioned automatically on startup — no manual setup needed.

### Metrics (Prometheus)

Go to **Explore → Prometheus**. Your app's OTel resource attributes (e.g. `service.name`) are promoted to Prometheus labels via `resource_to_telemetry_conversion` in the collector config.

```promql
# All metrics from the go-gin-api example app
{service_name="go-gin-api"}

# Request counter rate
rate(go_gin_api_healthcheck_total_requests_total[1m])
rate(go_gin_api_slow_total_requests_total[1m])
```

Use the **Metrics browser** button in Explore to browse all available metric names and filter by label.

### Logs (Loki)

Go to **Explore → Loki**.

```logql
# All logs from the go-gin-api example app
{service_name="go-gin-api"}

# Filter to errors only
{service_name="go-gin-api"} |= "ERROR"
```

---

## How the Pipeline Works

```
  Instrumented App
       │
       │  OTLP (gRPC :4317 or HTTP :4318)
       ▼
  OTel Collector
       │
       ├── Logs  ──► Loki  (:3100)
       │             (OTLP HTTP)
       │
       └── Metrics ──► Prometheus  (:9090)
                        (remote-write)
                             │
                        Grafana (:3000)
                    queries Prometheus + Loki
```

**OTel Collector** (`otel-collector/otel-collector-config.yaml`):
- Receives all telemetry over OTLP
- Fans logs out to Loki via OTLP HTTP
- Fans metrics out to Prometheus via remote-write
- `resource_to_telemetry_conversion: enabled: true` ensures resource attributes like `service_name` appear as Prometheus labels
- The `debug` exporter is also enabled in both pipelines — it prints detailed telemetry to the collector's **container stdout**, which is very useful when learning or troubleshooting (`docker compose logs otel-collector`)

**Prometheus** (`prometheus/prometheus.yml`):
- Scrapes its own `/metrics` endpoint
- Started with `--web.enable-remote-write-receiver` so the OTel Collector can push metrics to it
- Data is stored inside the container and is **not persisted** across `docker compose down` restarts

**Loki** (`loki/loki.yaml`):
- Single-node mode with filesystem storage under `/tmp/loki` inside the container
- Data is **not persisted** across `docker compose down` restarts
- `allow_structured_metadata: true` enables OTel log body fields to be stored as metadata

**Grafana** (`grafana/grafana.ini` + `grafana/provisioning/`):
- Datasources for Prometheus and Loki are provisioned from `provisioning/datasources/datasources.yaml` — no manual UI config needed
- Anonymous viewer access is enabled for easy local browsing
- Any dashboards created manually in the UI are **not persisted** — they will be lost on restart. Use provisioning files to keep dashboards across restarts (TODO)
- Loki derived fields are configured to link `traceID` log fields to Tempo (ready for when Tempo is added)

---

## Stopping the Stack

```bash
docker compose down
```

---

## TODO

- [ ] Add Tempo for distributed tracing
- [ ] Add the go-gin-api as a service in docker-compose so the full stack runs in one command
- [ ] Persist Loki data across restarts with a named Docker volume
- [ ] Persist Prometheus data across restarts with a named Docker volume
- [ ] Provision Grafana dashboards from files so they survive container restarts
