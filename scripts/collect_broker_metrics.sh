#!/usr/bin/env bash
set -euo pipefail

# Lightweight broker-side metrics collection using tools already installed in the benchmark AMI.
# Usage:
#   EMQX:      ./collect_broker_metrics.sh emqx 90 1 ~/metrics/emqx-run-01
#   Mosquitto: ./collect_broker_metrics.sh mosquitto 90 1 ~/metrics/mosquitto-run-01

BROKER="${1:-}"
DURATION="${2:-90}"
INTERVAL="${3:-1}"
OUT="${4:-$HOME/tcc-metrics/$(date -u +%Y%m%dT%H%M%SZ)}"

if [[ "$BROKER" != "emqx" && "$BROKER" != "mosquitto" ]]; then
  echo "Usage: $0 <emqx|mosquitto> [duration_seconds] [interval_seconds] [output_dir]" >&2
  exit 2
fi

mkdir -p "$OUT"
COUNT=$(( DURATION / INTERVAL ))
if (( COUNT < 1 )); then COUNT=1; fi

if [[ "$BROKER" == "emqx" ]]; then
  PROCESS_PATTERN="beam.smp"
  SERVICE="emqx"
else
  PROCESS_PATTERN="mosquitto"
  SERVICE="mosquitto"
fi

{
  echo "timestamp_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "broker=$BROKER"
  echo "hostname=$(hostname)"
  echo "duration_seconds=$DURATION"
  echo "interval_seconds=$INTERVAL"
  echo "kernel=$(uname -r)"
  echo "service=$SERVICE"
  systemctl is-active "$SERVICE" || true
  systemctl status "$SERVICE" --no-pager || true
} > "$OUT/metadata.txt"

# System CPU, memory and network statistics.
sar -u "$INTERVAL" "$COUNT" > "$OUT/sar_cpu.txt" & P1=$!
sar -r "$INTERVAL" "$COUNT" > "$OUT/sar_memory.txt" & P2=$!
sar -n DEV "$INTERVAL" "$COUNT" > "$OUT/sar_network.txt" & P3=$!

# Broker process CPU/RSS/faults. EMQX runs mainly inside the Erlang beam.smp process.
pidstat -u -r -h -C "$PROCESS_PATTERN" "$INTERVAL" "$COUNT" > "$OUT/pidstat_broker.txt" & P4=$!

# Established TCP connections whose local port is 1883.
(
  echo "timestamp_utc,connections_1883"
  for ((i=0; i<COUNT; i++)); do
    TS=$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)
    N=$(ss -Htn state established | awk '$4 ~ /:1883$/ {n++} END {print n+0}')
    echo "$TS,$N"
    sleep "$INTERVAL"
  done
) > "$OUT/connections.csv" & P5=$!

wait "$P1" "$P2" "$P3" "$P4" "$P5"

echo "Metrics saved to: $OUT"
