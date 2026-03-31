#!/bin/bash

# Prometheus: https://github.com/prometheus-community/helm-charts
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

# Tempo: https://grafana.com/docs/helm-charts/tempo-distributed/next/get-started-helm-charts/
# Grafana: https://grafana.com/docs/grafana/latest/setup-grafana/installation/helm/
# Loki: https://grafana.com/docs/loki/latest/setup/install/helm/install-microservices/
helm repo add grafana https://grafana.github.io/helm-charts

# Tempo: https://grafana.com/docs/helm-charts/tempo-distributed/next/get-started-helm-charts/
# Grafana: https://grafana.com/docs/grafana/latest/setup-grafana/installation/helm/
# Loki: https://grafana.com/docs/loki/latest/setup/install/helm/install-microservices/
helm repo add grafana-community https://grafana-community.github.io/helm-charts

# Open Telemetry Collector: https://opentelemetry.io/docs/platforms/kubernetes/helm/collector/
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts

helm repo update