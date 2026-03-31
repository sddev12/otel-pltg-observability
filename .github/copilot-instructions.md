# Copilot Instructions for This Repository

## Project Goal
This repository is a scaffold and learning reference for SRE teams and Platform Engineers setting up a full observability stack on Kubernetes. It covers:

- The **PLTG observability stack**: Prometheus, Loki, Tempo, and Grafana
- **OpenTelemetry** instrumentation and collection
- **Traefik** as ingress controller
- **Kubernetes on Minikube** for local development and learning
- **Docker Compose** for quick local spin-up without Kubernetes

The repo serves two audiences:
1. **Teams who want a scaffold** — copy the Helm values, manifests, and config as a starting point for their own cluster
2. **Learners** — walk through the example apps and stack config to understand observability end-to-end across all three pillars (metrics, logs, traces)

Prioritize clarity and learning value over production complexity.

## Current Repository Structure
```
.github/
  copilot-instructions.md

example-apps/
  go-gin-api/           # Go + Gin API service with OTel instrumentation
    main.go
    go.mod
    observability/otel.go
    utils/utils.go
  traefik-whoami/       # Traefik whoami sample manifests + local TLS certs
    whoami.yaml
    whoami-service.yaml
    whoami-ingress.yaml
    whoami-cert-secret.yaml
    certs/

observability-stack/
  grafana/helm-values.yaml
  loki/helm-values.yaml       # Distributed mode, MinIO storage
  tempo/helm-values.yaml      # S3/MinIO storage, OTLP ingest
  prometheus/                 # Placeholder — values TBD

otel-collector/
  helm-values.yaml
  otel-collector-config.yaml  # Currently logs pipeline only; traces/metrics TBD
  run-collector.sh

traefik/
  helm-values.yaml            # HTTP→HTTPS redirect, dashboard, Prometheus metrics

scripts/
  setup-helm.sh               # Adds all required Helm repos
  start-app.sh                # Runs go-gin-api locally
  generate-traffic.sh         # Sends test traffic to the API

docker-compose.yaml           # Local Docker-based spin-up (in progress)
README.md                     # (in progress)
```

When making changes, keep each area focused and avoid mixing unrelated concerns in one commit.

## Environment Assumptions
- Local Kubernetes cluster is **Minikube**
- **Traefik** is used as ingress controller, deployed via Helm
- **OTel Collector** is deployed via Helm as the central telemetry pipeline
- Observability backends deployed via Helm: Prometheus, Loki (distributed), Tempo, Grafana
- Storage for Loki and Tempo uses embedded **MinIO** (local dev convenience)
- **Docker Compose** is an alternative local path that avoids Kubernetes entirely
- Changes should be runnable locally with minimal setup friction

## Coding and Design Priorities
1. Keep examples small and readable.
2. Favor explicit, easy-to-follow code over abstractions.
3. As this is a repo for learning, don't be shy with comments, add them where it helps the user to learn.
4. Include quick run/test instructions in README updates when behavior changes.
5. Keep dependencies minimal unless a dependency clearly improves learning value.

## Kubernetes and Manifests Guidance
- Prefer plain Kubernetes YAML for learning examples unless Helm is explicitly requested.
- Use clear naming:
  - app labels like `app.kubernetes.io/name`
  - resource names that match service purpose
- Keep manifests split by concern when practical:
  - deployment
  - service
  - ingress
  - config/secret
- Default to ClusterIP services behind Traefik ingress.
- Keep probes and resource requests/limits simple but present in app deployments.

## OpenTelemetry Guidance
When adding or modifying app services:
- Instrument HTTP handlers with OpenTelemetry spans.
- Propagate trace context across service-to-service HTTP calls.
- Add basic business spans inside handlers so traces are easy to inspect.
- Record key attributes (route, status code, downstream target).
- Avoid high-cardinality labels/attributes.

For metrics:
- Expose service metrics compatible with Prometheus scraping.
- Include at least request count, request latency, and error count.

For logs:
- Use structured logs where possible.
- Include trace/span correlation fields when available.

## Multi-Service Learning Patterns
For additional example apps, prefer scenarios that demonstrate trace propagation:
- `frontend` calls `api`
- `api` calls `worker` or `dependency`
- one intentional slow path to make latency visible in traces
- one intentional error path to demonstrate failed spans and alerts

Keep these examples deterministic and easy to trigger with curl.

## Traefik-Specific Guidance
- Route services through Traefik ingress resources.
- Use explicit host/path matching for clarity.
- If TLS examples are included, keep cert handling local-dev friendly.

## Suggested Defaults for New Go Services
- Use `gin` only if it improves readability for the learning objective.
- Keep startup code simple and explicit.
- Expose health endpoint(s): `/healthz` and optionally `/readyz`.
- Provide a minimal Dockerfile and Kubernetes manifests alongside the service.

## Verification Expectations
When implementing features, validate with:
1. Build success for changed service(s)
2. Kubernetes resources apply cleanly
3. Service reachable through Traefik ingress
4. Traces visible in Tempo/Grafana
5. Metrics visible in Prometheus/Grafana

If full validation cannot be run, state what was not validated and why.

## Copilot Response Style in This Repository
When proposing changes:
- Explain what files were changed and why.
- Provide copy-pasteable commands for Minikube/Kubernetes checks.
- Prefer incremental steps over large one-shot rewrites.
- Call out tradeoffs and learning notes briefly.

## Helm Chart Reference URLs
When helping with Helm configuration, consult the upstream default values for the full list of available options:

- **Prometheus**: https://raw.githubusercontent.com/prometheus-community/helm-charts/refs/heads/main/charts/prometheus/values.yaml

## Out of Scope by Default
- Production-grade hardening
- Multi-cluster setup
- Complex CI/CD pipelines
- Overly abstract framework code

Unless requested, keep examples local-first and learning-oriented.
