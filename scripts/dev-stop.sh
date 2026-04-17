#!/bin/bash
# ==========================================
# SocietyKro — Stop All Backend Services
# ==========================================

cd "$(dirname "$0")/.."
ROOT=$(pwd)

echo "Stopping all services..."
for pidfile in "$ROOT"/logs/*.pid; do
  if [ -f "$pidfile" ]; then
    pid=$(cat "$pidfile")
    svc=$(basename "$pidfile" .pid)
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid"
      echo "  Stopped $svc (PID $pid)"
    fi
    rm -f "$pidfile"
  fi
done

# Force kill by port as fallback
for port in 8081 8082 8083 8084 8085 8086 8089 8090; do
  fuser -k $port/tcp 2>/dev/null || true
done

echo "All services stopped."
echo ""
echo "To also stop Docker containers: docker compose down"
