#!/bin/bash

DURATION=${1:-90}
INTERVAL=1
OUTDIR=${2:-"$HOME/tcc-metrics/run-$(date -u +%Y%m%dT%H%M%SZ)"}

mkdir -p "$OUTDIR"

echo "Coletando métricas por ${DURATION}s em:"
echo "$OUTDIR"

echo "started_at_utc=$(date -u --iso-8601=seconds)" > "$OUTDIR/metadata.txt"
echo "hostname=$(hostname)" >> "$OUTDIR/metadata.txt"
echo "duration_seconds=$DURATION" >> "$OUTDIR/metadata.txt"
echo "kernel=$(uname -r)" >> "$OUTDIR/metadata.txt"

# ---------------------------------------------------------
# Identifica broker e PID principal através do systemd
# ---------------------------------------------------------

if systemctl is-active --quiet emqx; then
    BROKER="emqx"
    BROKER_PID=$(systemctl show -p MainPID --value emqx)
elif systemctl is-active --quiet mosquitto; then
    BROKER="mosquitto"
    BROKER_PID=$(systemctl show -p MainPID --value mosquitto)
else
    echo "ERRO: nenhum broker suportado está ativo."
    exit 1
fi

echo "broker=$BROKER" >> "$OUTDIR/metadata.txt"
echo "broker_pid=$BROKER_PID" >> "$OUTDIR/metadata.txt"

echo "Broker detectado: $BROKER"
echo "PID: $BROKER_PID"

# ---------------------------------------------------------
# CPU total da EC2
# ---------------------------------------------------------

sar -u "$INTERVAL" "$DURATION" \
    > "$OUTDIR/cpu.txt" &
PID_CPU=$!

# ---------------------------------------------------------
# Memória total da EC2
# ---------------------------------------------------------

sar -r "$INTERVAL" "$DURATION" \
    > "$OUTDIR/memory.txt" &
PID_MEM=$!

# ---------------------------------------------------------
# Rede
# ---------------------------------------------------------

sar -n DEV "$INTERVAL" "$DURATION" \
    > "$OUTDIR/network.txt" &
PID_NET=$!

# ---------------------------------------------------------
# CPU e memória do processo principal do broker
# ---------------------------------------------------------

pidstat -u -r -p "$BROKER_PID" "$INTERVAL" "$DURATION" \
    > "$OUTDIR/process.txt" &
PID_PROCESS=$!

# ---------------------------------------------------------
# Número de conexões MQTT estabelecidas
# ---------------------------------------------------------

(
    echo "timestamp_utc,connections"

    for ((i=0; i<DURATION; i++)); do
        TS=$(date -u +"%Y-%m-%dT%H:%M:%S.%3NZ")

        COUNT=$(ss -tan state established \
            '( sport = :1883 )' \
            | tail -n +2 \
            | wc -l)

        echo "$TS,$COUNT"

        sleep "$INTERVAL"
    done
) > "$OUTDIR/connections.csv" &
PID_CONN=$!

wait "$PID_CPU"
wait "$PID_MEM"
wait "$PID_NET"
wait "$PID_PROCESS"
wait "$PID_CONN"

echo "finished_at_utc=$(date -u --iso-8601=seconds)" \
    >> "$OUTDIR/metadata.txt"

echo
echo "Coleta concluída."
echo "Resultados:"
echo "$OUTDIR"