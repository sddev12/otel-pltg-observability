# Copilot Instructions for This Repository

## Project Goal
This repository is a scaffold and learning reference for SRE teams and Platform Engineers setting up a full observability stack. It covers:

- The **PLTG observability stack**: Prometheus, Loki, Tempo, and Grafana
- **OpenTelemetry** instrumentation and collection
- **Traefik** as ingress controller
- **Kubernetes on Minikube** for a realistic cluster-based setup
- **Docker Compose** for a quick local spin-up without Kubernetes

The repo serves two audiences:
1. **Teams who want a scaffold** — copy the Helm values, manifests, and config as a starting point for their own cluster
2. **Learners** — walk through the example apps and stack config to understand observability end-to-end across all three pillars (metrics, logs, traces)

Prioritize clarity and learning value over production complexity.

---

## Repository Structure

The repo is split into three top-level areas:

```
.github/
  copilot-instructions.md

example-apps/                     # Instrumented example services
  go-gin-api/                     # Go + Gin API — the primary learning example
    main.go                       # Wires OTel, metrics, and routes
    go.mod                        # Module: github.com/sddev12/go-gin-api
    handlers/
      healthz.go                  # GET /healthz — fast 200 response
      slow.go                     # GET /slow — sleeps 3s, demonstrates latency
      error_gen.go                # GET /errorgen — always returns 500
      init_metrics.go             # Registers OTel metric instruments (Int64Counters)
    observability/
      otel.go                     # SetupOpenTelemetry: logging + metrics providers
    utils/
      utils.go                    # Loads .env file and validates required env vars
  scripts/
    start-go-gin-api.sh           # cd ../go-gin-api && go run main.go
  traefik-whoami/                 # Traefik whoami sample manifests + local TLS certs
    whoami.yaml
    whoami-service.yaml
    whoami-ingress.yaml
    whoami-cert-secret.yaml
    certs/

local-project/                    # Docker Compose path — no Kubernetes required
  docker-compose.yaml             # Brings up Prometheus, Loki, OTel Collector, Grafana
  README.md                       # How to start and use the local stack
  prometheus/
    prometheus.yml                # Scrape config; remote-write receiver enabled
  loki/
    loki.yaml                     # Single-node, filesystem-backed Loki
  otel-collector/
    otel-collector-config.yaml    # OTLP receiver → Loki (logs) + Prometheus (metrics)
  grafana/
    grafana.ini                   # Server config, auth, anonymous access, feature toggles
    provisioning/
      datasources/
        datasources.yaml          # Auto-provisions Prometheus + Loki datasources
  tempo/                          # TODO — empty; traces not yet configured
  scripts/
    generate-traffic.sh           # Sends randomised traffic to all go-gin-api endpoints

k8s-project/                      # Kubernetes / Minikube path
  observability-stack/
    grafana/helm-values.yaml
    loki/helm-values.yaml         # Distributed mode, MinIO storage
    tempo/helm-values.yaml        # MinIO storage, OTLP gRPC + HTTP ingest
    prometheus/helm-values.yaml   # OTLP write receiver, resource attribute promotion
    otel-collector/helm-values.yaml
  traefik/
    helm-values.yaml              # HTTP→HTTPS redirect, dashboard with BasicAuth
  scripts/
    setup-helm.sh                 # Adds prometheus-community, grafana, open-telemetry repos
```

---

## Two Project Paths

### local-project (Docker Compose)
Run the full stack locally without a cluster. Start from the `local-project/` directory:

```bash
cd local-project
docker compose up
```

| Service | URL | Notes |
|---|---|---|
| Grafana | http://localhost:3000 | `admin` / `admin`; anonymous viewer access enabled |
| Prometheus | http://localhost:9090 | Remote-write receiver enabled |
| Loki | http://localhost:3100 | OTLP HTTP ingest at `/otlp` |
| OTel Collector gRPC | `localhost:4317` | App sends telemetry here |
| OTel Collector HTTP | `localhost:4318` | |

Run the example app alongside the stack:

```bash
cd example-apps/go-gin-api
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 LOG_LEVEL=INFO go run main.go
```

Generate traffic:

```bash
./local-project/scripts/generate-traffic.sh
```

**What is working in local-project:**
- Logs: app → OTel Collector (OTLP gRPC) → Loki (OTLP HTTP) → Grafana Explore
- Metrics: app → OTel Collector (OTLP gRPC) → Prometheus (remote-write) → Grafana Explore
- Grafana datasources are provisioned automatically on startup

**What is TODO in local-project:**
- `local-project/tempo/` is empty — Tempo container and config not yet added
- Once Tempo is added, the Grafana datasource UID `tempo` referenced in `datasources.yaml` will resolve

### k8s-project (Kubernetes / Minikube)
Helm-based deployment to a local Minikube cluster. All component Helm values live under `k8s-project/observability-stack/` and `k8s-project/traefik/`.

Add Helm repos first:

```bash
./k8s-project/scripts/setup-helm.sh
```

Key configuration decisions in the Kubernetes path:
- **Prometheus**: OTLP write receiver enabled; promotes OTel resource attributes (`service.name`, `k8s.pod.name`, etc.) directly to Prometheus labels
- **Loki**: Distributed mode with 3 ingester replicas; MinIO for object storage
- **Tempo**: Standalone with embedded MinIO; accepts OTLP gRPC and HTTP
- **Traefik**: HTTP→HTTPS redirect on ports 80/443; dashboard at `dashboard.minikube.local` behind BasicAuth; add `$(minikube ip) <hostname>` entries to `/etc/hosts` for each ingress host
- **OTel Collector**: Deployed via the `open-telemetry/opentelemetry-collector` Helm chart

---

## Example App: go-gin-api

### Endpoints
| Route | Handler | Behaviour |
|---|---|---|
| `GET /healthz` | `handlers.HandleHealthz` | Returns 200; increments `go_gin_api.healthcheck.total_requests` |
| `GET /slow` | `handlers.HandleSlow` | Sleeps 3 s then returns 200; increments `go_gin_api.slow.total_requests` |
| `GET /errorgen` | `handlers.HandleErrorGen` | Always returns 500; increments `go_gin_api.healthcheck.total_requests` (reuses same counter — worth noting as a learning point) |

### OTel Instrumentation Status
| Signal | Status | Detail |
|---|---|---|
| Logs | ✅ Working | `slog` + `otelslog` bridge; exported via OTLP gRPC to `localhost:4317` AND stdout |
| Metrics | ✅ Working | OTel `metric.Int64Counter` per endpoint; exported via OTLP gRPC to `localhost:4317` AND stdout |
| Traces | ❌ TODO | No span instrumentation yet; Tempo datasource referenced but not wired up |

### OTel Setup (`observability/otel.go`)
- `SetupOpenTelemetry` returns a single `shutdownOtel` func that cleans up all providers
- Logging: stdout exporter + OTLP gRPC exporter → both attached as `BatchProcessor`s on a single `LoggerProvider`
- Metrics: stdout exporter + OTLP gRPC metric exporter → `MeterProvider`
- The OTLP endpoint is hardcoded to `localhost:4317` — when running in Docker Compose, override with `OTEL_EXPORTER_OTLP_ENDPOINT`
- `slog.SetDefault` is called so all `slog` calls throughout the app use the OTel-backed handler

### Environment Variables
| Variable | Required | Default | Notes |
|---|---|---|---|
| `LOG_LEVEL` | Yes | — | Loaded from `.env` or shell; validated on startup |
| `GIN_MODE` | No | `debug` | Set to `release` in the start script |

### Running the App
```bash
# Via start script (sets LOG_LEVEL and GIN_MODE):
cd example-apps
./scripts/start-go-gin-api.sh

# Or directly (e.g. pointed at local Docker Compose stack):
cd example-apps/go-gin-api
LOG_LEVEL=INFO go run main.go
```

### What is Missing from go-gin-api
- No Dockerfile
- No Kubernetes manifests (Deployment, Service, Ingress)
- No trace instrumentation (the next logical step is adding `otel.Tracer` spans to handlers)

---

## Telemetry Pipeline Summary

### Docker Compose pipeline
```
go-gin-api
  └── OTLP gRPC (localhost:4317)
        └── otel-collector
              ├── logs  → Loki   (OTLP HTTP http://loki:3100/otlp)
              └── metrics → Prometheus (remote-write http://prometheus:9090/api/v1/write)
                              └── Grafana (provisioned datasource)
```

### Kubernetes pipeline (intended, Helm-based)
```
go-gin-api (pod)
  └── OTLP gRPC → otel-collector (Helm)
                    ├── logs    → Loki (distributed, MinIO)
                    ├── metrics → Prometheus (OTLP write receiver)
                    └── traces  → Tempo (MinIO)
                                    └── Grafana (all three datasources)
```

---

## Coding and Design Priorities
1. Keep examples small and readable.
2. Favor explicit, easy-to-follow code over abstractions.
3. This is a learning repo — add comments where they help explain *why*, not just *what*.
4. Include quick run/test instructions in README updates when behaviour changes.
5. Keep dependencies minimal unless they clearly improve learning value.

---

## Kubernetes and Manifests Guidance
- Prefer plain Kubernetes YAML for learning examples unless Helm is explicitly requested.
- Use clear naming:
  - app labels: `app.kubernetes.io/name`
  - resource names that match the service purpose
- Split manifests by concern: Deployment, Service, Ingress, ConfigMap/Secret in separate files.
- Default to ClusterIP services routed through Traefik ingress.
- Keep probes and resource requests/limits simple but present in app deployments.
- For Minikube ingress hosts, add entries to `/etc/hosts`: `$(minikube ip)  <hostname>`

---

## OpenTelemetry Guidance
When adding or modifying app services:
- Instrument HTTP handlers with OpenTelemetry spans using `otel.Tracer`.
- Propagate trace context across service-to-service HTTP calls via W3C TraceContext headers.
- Add at least one child span per handler so traces have visible structure.
- Record key span attributes: route, HTTP status code, any downstream target.
- Avoid high-cardinality attributes (e.g. raw user IDs, full URLs with query strings).

For metrics:
- Use OTel `metric` SDK instruments (Counter, Histogram, Gauge) rather than a separate Prometheus client.
- Name metrics with a service prefix, e.g. `go_gin_api.<signal>.<name>`.
- Include at least: request count, request latency (histogram), error count.

For logs:
- Use `slog` with the `otelslog` bridge so log records are exported via the OTel log pipeline.
- Include trace/span ID correlation fields when a span is active — the bridge does this automatically when the context is propagated correctly.

---

## Multi-Service Learning Patterns
For additional example apps, prefer scenarios that demonstrate trace propagation:
- `frontend` calls `api`
- `api` calls `worker` or `dependency`
- one intentional slow path (`time.Sleep`) to make latency visible in traces
- one intentional error path to demonstrate failed spans and alerts

Keep these examples deterministic and easy to trigger with curl.

---

## Traefik-Specific Guidance
- Route services through Traefik Ingress resources using `ingressClassName: traefik`.
- Use explicit host/path matching for clarity.
- Dashboard is served at `dashboard.minikube.local` (HTTPS only, BasicAuth).
- Port mapping: HTTP 80 → NodePort 30000 (redirects to HTTPS); HTTPS 443 → NodePort 30001.
- If TLS examples are included, keep cert handling local-dev friendly (self-signed is fine).

---

## Suggested Defaults for New Go Services
- Use `gin` if it improves readability; otherwise `net/http` is fine.
- Keep startup code simple and explicit (no DI frameworks).
- Expose `/healthz` and optionally `/readyz`.
- Reuse the pattern in `observability/otel.go`: call `SetupOpenTelemetry` early in `main`, defer the shutdown func.
- Provide a minimal Dockerfile and Kubernetes manifests alongside any new service.

---

## Verification Expectations
When implementing features, validate with:
1. Build success for changed service(s) (`go build ./...`)
2. Kubernetes resources apply cleanly (`kubectl apply --dry-run=client -f <file>`)
3. Service reachable through Traefik ingress
4. Traces visible in Tempo via Grafana Explore
5. Metrics visible in Prometheus via Grafana Explore
6. Logs visible in Loki via Grafana Explore

If full validation cannot be run, state what was not validated and why.

---

## Copilot Response Style in This Repository
When proposing changes:
- Explain what files were changed and why.
- Provide copy-pasteable commands for local or Minikube checks.
- Prefer incremental steps over large one-shot rewrites.
- Call out tradeoffs and learning notes briefly.
- When a feature is partially implemented (e.g. traces TODO), say so clearly.

---

## Helm Chart Reference URLs
When helping with Helm configuration, fetch the upstream default values for the complete list of available options:

- **Prometheus**: https://raw.githubusercontent.com/prometheus-community/helm-charts/refs/heads/main/charts/prometheus/values.yaml
- **Grafana**: https://raw.githubusercontent.com/grafana/helm-charts/refs/heads/main/charts/grafana/values.yaml
- **Loki (distributed)**: https://raw.githubusercontent.com/grafana/helm-charts/refs/heads/main/charts/loki-distributed/values.yaml
- **Tempo**: https://raw.githubusercontent.com/grafana/helm-charts/refs/heads/main/charts/tempo/values.yaml
- **OTel Collector**: https://raw.githubusercontent.com/open-telemetry/opentelemetry-helm-charts/refs/heads/main/charts/opentelemetry-collector/values.yaml
- **Traefik**: https://raw.githubusercontent.com/traefik/traefik-helm-chart/refs/heads/master/traefik/values.yaml

---

## Out of Scope by Default
- Production-grade hardening
- Multi-cluster setup
- Complex CI/CD pipelines
- Overly abstract framework code

Unless requested, keep examples local-first and learning-oriented.
