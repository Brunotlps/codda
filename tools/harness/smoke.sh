#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BASE_URL="${BASE_URL:-http://localhost:8080}"
DATABASE_URL="${DATABASE_URL:-postgres://codda:codda@localhost:5433/codda?sslmode=disable}"
HTTP_PORT="${HTTP_PORT:-8080}"
LOG_FILE="${LOG_FILE:-/tmp/codda-smoke.log}"

SERVICE_PID=""

cleanup() {
  if [[ -n "${SERVICE_PID}" ]] && kill -0 "${SERVICE_PID}" 2>/dev/null; then
    kill "${SERVICE_PID}" 2>/dev/null || true
    wait "${SERVICE_PID}" 2>/dev/null || true
  fi
}
trap cleanup EXIT

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

wait_for_ready() {
  for _ in $(seq 1 40); do
    if curl -fsS "${BASE_URL}/ready" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.25
  done

  echo "service did not become ready at ${BASE_URL}/ready" >&2
  echo "service log: ${LOG_FILE}" >&2
  exit 1
}

json_field() {
  python3 -c 'import json,sys; print(json.load(sys.stdin)[sys.argv[1]])' "$1"
}

require_command curl
require_command docker
require_command go
require_command python3

cd "${ROOT_DIR}"

echo "starting postgres with docker compose"
docker compose up -d postgres

echo "starting orderservice on :${HTTP_PORT}"
DATABASE_URL="${DATABASE_URL}" HTTP_PORT="${HTTP_PORT}" go run ./cmd/orderservice/ >"${LOG_FILE}" 2>&1 &
SERVICE_PID="$!"

wait_for_ready

echo "checking liveness"
curl -fsS "${BASE_URL}/health" >/dev/null

echo "creating order"
CREATE_RESPONSE="$(
  curl -fsS -X POST "${BASE_URL}/orders" \
    -H "Content-Type: application/json" \
    -d '{
      "items": [
        {"product_id": "p1", "product_name": "Widget", "price_cents": 1999, "quantity": 2}
      ]
    }'
)"
ORDER_ID="$(printf '%s' "${CREATE_RESPONSE}" | json_field id)"

if [[ -z "${ORDER_ID}" ]]; then
  echo "create order did not return an id" >&2
  exit 1
fi

echo "created order ${ORDER_ID}"

echo "reading order"
curl -fsS "${BASE_URL}/orders/${ORDER_ID}" >/dev/null

echo "marking order as paid"
curl -fsS -X POST "${BASE_URL}/orders/${ORDER_ID}/pay" >/dev/null

echo "marking order as shipped"
curl -fsS -X POST "${BASE_URL}/orders/${ORDER_ID}/ship" >/dev/null

echo "listing orders"
curl -fsS "${BASE_URL}/orders?limit=10" >/dev/null

echo "smoke harness passed"
