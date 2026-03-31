#!/bin/bash
#
# Sends traffic to all go-gin-api endpoints in a randomised order.
#
# Endpoints:
#   GET /healthz   — fast health check, returns 200
#   GET /slow      — sleeps 3s then returns 200 (good for latency in traces)
#   GET /errorgen  — always returns 500 (good for error rate dashboards)
#
# Usage:
#   ./generate-traffic.sh              # target default http://localhost:8080
#   ./generate-traffic.sh http://localhost:9090  # custom base URL

BASE_URL="${1:-http://localhost:8080}"

ENDPOINTS=(
  "/healthz"
  "/healthz"
  "/healthz"   # weighted 3x so healthy traffic dominates
  "/slow"
  "/errorgen"
)

echo "Sending traffic to ${BASE_URL} — press Ctrl+C to stop"
echo ""

while true; do
  # Pick a random endpoint from the array
  ENDPOINT="${ENDPOINTS[$RANDOM % ${#ENDPOINTS[@]}]}"
  URL="${BASE_URL}${ENDPOINT}"

  # Call the endpoint, capture HTTP status code
  HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${URL}")

  echo "$(date '+%H:%M:%S')  ${HTTP_STATUS}  ${ENDPOINT}"

  # Random sleep between 0.5s and 2.5s to vary the request rate
  sleep "$(awk 'BEGIN { srand(); printf "%.1f\n", 0.5 + rand() * 2 }')"
done
