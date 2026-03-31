# Observability Project

A scaffold and learning reference for the **PLTG observability stack** — Prometheus, Loki, Tempo, and Grafana — with **OpenTelemetry** for instrumentation.

This repo is aimed at SRE teams and Platform Engineers who want a working starting point, and at learners who want to understand the three pillars of observability — **metrics, logs, and traces** — end to end.

The project is split into three areas:

| Directory | Purpose | Status |
|---|---|---|
| `local-project/` | Run the full stack locally with Docker Compose — no Kubernetes required | Active |
| `k8s-project/` | Deploy the stack to Kubernetes using Helm | Work in progress |
| `example-apps/` | Instrumented example applications to learn OTel with | Active |

---

## How the Stack Fits Together

```
  Instrumented App (example-apps/)
         │
         │  OTLP (gRPC :4317 or HTTP :4318)
         ▼
    OTel Collector
         │
         ├── Logs  ──────► Loki
         ├── Traces ─────► Tempo
         └── Metrics ────► Prometheus
                                │
                           Grafana
                   (single pane for all three pillars)
```

Apps are instrumented with the OpenTelemetry SDK and export telemetry to the OTel Collector over OTLP. The collector fans telemetry out to the appropriate backend. Grafana sits in front of all three backends and is where you query, visualise, and alert.

---

## local-project/ — Docker Compose Stack

Everything needed to run the observability stack locally using Docker Compose. No Kubernetes or Helm required — just Docker.

**What's running:**

| Service | URL |
|---|---|
| Grafana | http://localhost:3000 |
| Prometheus | http://localhost:9090 |
| Loki | http://localhost:3100 |
| OTel Collector (gRPC) | localhost:4317 |
| OTel Collector (HTTP) | localhost:4318 |

**Quick start:**

```bash
cd local-project
docker compose up
```

See [local-project/README.md](local-project/README.md) for the full setup guide, config details, and example queries.

> **Tempo (traces) is not yet configured** in the local stack. The compose setup currently covers metrics and logs.

---

## k8s-project/ — Kubernetes (Work in Progress)

Helm values and Kubernetes manifests for deploying the full PLTG stack to a Kubernetes cluster (Minikube for local development).

**What will be included:**

- Prometheus, Loki, Tempo, Grafana — deployed via Helm
- OTel Collector — deployed via Helm
- Traefik — ingress controller with HTTP→HTTPS redirect and Prometheus metrics
- All services routed through Traefik ingress

**Stack:**
- Local cluster: [Minikube](https://minikube.sigs.k8s.io/docs/start/)
- Ingress: [Traefik](https://traefik.io/)
- Storage: embedded MinIO for Loki and Tempo (local dev convenience)

This directory is a work in progress. Refer back here as configs are added.

---

## example-apps/ — Instrumented Example Apps

Example applications instrumented with OpenTelemetry. Use these to learn how to add metrics, logs, and traces to a real service and see the telemetry flow through the stack.

### go-gin-api

A Go HTTP API built with [Gin](https://github.com/gin-gonic/gin), instrumented with the OTel Go SDK.

**What it demonstrates:**
- Setting up the OTel SDK (metric provider, logger provider) with a resource
- Exporting metrics and logs to the OTel Collector over OTLP gRPC
- Defining custom metrics (`Int64Counter`) and recording them per request
- Structured logging with trace correlation via `otelslog`

**Endpoints:**

| Endpoint | Behaviour |
|---|---|
| `GET /healthz` | Fast health check, returns `{"status":"ok"}` |
| `GET /slow` | Sleeps 3 seconds then returns 200 — useful for latency in traces |
| `GET /errorgen` | Always returns 500 — useful for error rate dashboards and alerts |

**Run it:**

```bash
cd example-apps/go-gin-api
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 go run main.go
```

Or use the helper script:

```bash
./example-apps/scripts/start-go-gin-api.sh
```

Then generate traffic against all endpoints:

```bash
./local-project/scripts/generate-traffic.sh
```

### traefik-whoami

Plain Kubernetes manifests for the [Traefik whoami](https://github.com/traefik/whoami) app. Useful for verifying that Traefik ingress routing and TLS termination are working before deploying a real service.

```bash
kubectl apply -f example-apps/traefik-whoami/
```

---

## Scripts

Scripts live alongside the area they relate to:

| Script | Purpose |
|---|---|
| `local-project/scripts/generate-traffic.sh` | Sends randomised traffic to all go-gin-api endpoints |
| `example-apps/scripts/start-go-gin-api.sh` | Starts the go-gin-api locally with sensible defaults |
| `k8s-project/scripts/setup-helm.sh` | Adds all required Helm repos for the Kubernetes stack |

---

## Prerequisites

Depending on which path you take:

**Docker Compose (local-project):**
- [Docker](https://docs.docker.com/get-docker/) with Compose v2
- [Go](https://go.dev/dl/) 1.21+ (to run example apps)

**Kubernetes (k8s-project):**
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/) v3+
- [Go](https://go.dev/dl/) 1.21+ (to run example apps)


## Status

| Component | Status |
|---|---|
| Traefik ingress | Ready |
| Prometheus | Helm values in progress |
| Loki (distributed) | Helm values ready |
| Tempo | Helm values ready |
| Grafana | Helm values ready |
| OTel Collector | Logs pipeline ready; traces/metrics TBD |
| go-gin-api example | OTel logging ready; tracing TBD |
| traefik-whoami example | Manifests ready |
| Docker Compose | In progress |
