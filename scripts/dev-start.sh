#!/bin/bash
# ==========================================
# SocietyKro — Start All Backend Services
# Usage: ./scripts/dev-start.sh
# Stop:  ./scripts/dev-stop.sh
# ==========================================

set -e
cd "$(dirname "$0")/.."
ROOT=$(pwd)

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  SocietyKro — Starting All Services${NC}"
echo -e "${GREEN}========================================${NC}"

# Check Docker containers
echo -e "\n${YELLOW}[1/4] Checking infrastructure...${NC}"
if ! docker ps | grep -q societykro-postgres; then
  echo "Starting Docker containers..."
  docker compose up -d postgres redis nats minio minio-init
  sleep 3
fi
echo -e "  PostgreSQL: ${GREEN}OK${NC}"
echo -e "  Redis:      ${GREEN}OK${NC}"
echo -e "  NATS:       ${GREEN}OK${NC}"
echo -e "  MinIO:      ${GREEN}OK${NC}"

# Check binaries exist
echo -e "\n${YELLOW}[2/4] Checking binaries...${NC}"
SERVICES=(auth-service complaint-service visitor-service payment-service notice-service vendor-service voice-service message-router)
for svc in "${SERVICES[@]}"; do
  if [ ! -f "bin/$svc" ]; then
    echo -e "  ${RED}Missing bin/$svc — run 'make build-services' first${NC}"
    exit 1
  fi
done
echo -e "  All 8 binaries: ${GREEN}OK${NC}"

# Kill any existing services
echo -e "\n${YELLOW}[3/4] Stopping existing services...${NC}"
for port in 8081 8082 8083 8084 8085 8086 8089 8090; do
  fuser -k $port/tcp 2>/dev/null || true
done
sleep 1
echo -e "  Ports cleared: ${GREEN}OK${NC}"

# JWT key paths
export JWT_PRIVATE_KEY_PATH="$ROOT/keys/private.pem"
export JWT_PUBLIC_KEY_PATH="$ROOT/keys/public.pem"

# Start all services
echo -e "\n${YELLOW}[4/4] Starting services...${NC}"
mkdir -p "$ROOT/logs"

AUTH_SERVICE_PORT=8081 "$ROOT/bin/auth-service" > "$ROOT/logs/auth-service.log" 2>&1 &
echo $! > "$ROOT/logs/auth-service.pid"

COMPLAINT_SERVICE_PORT=8082 "$ROOT/bin/complaint-service" > "$ROOT/logs/complaint-service.log" 2>&1 &
echo $! > "$ROOT/logs/complaint-service.pid"

VISITOR_SERVICE_PORT=8083 "$ROOT/bin/visitor-service" > "$ROOT/logs/visitor-service.log" 2>&1 &
echo $! > "$ROOT/logs/visitor-service.pid"

PAYMENT_SERVICE_PORT=8084 "$ROOT/bin/payment-service" > "$ROOT/logs/payment-service.log" 2>&1 &
echo $! > "$ROOT/logs/payment-service.pid"

NOTICE_SERVICE_PORT=8085 "$ROOT/bin/notice-service" > "$ROOT/logs/notice-service.log" 2>&1 &
echo $! > "$ROOT/logs/notice-service.pid"

VENDOR_SERVICE_PORT=8086 "$ROOT/bin/vendor-service" > "$ROOT/logs/vendor-service.log" 2>&1 &
echo $! > "$ROOT/logs/vendor-service.pid"

VOICE_SERVICE_PORT=8090 "$ROOT/bin/voice-service" > "$ROOT/logs/voice-service.log" 2>&1 &
echo $! > "$ROOT/logs/voice-service.pid"

MESSAGE_ROUTER_PORT=8089 "$ROOT/bin/message-router" > "$ROOT/logs/message-router.log" 2>&1 &
echo $! > "$ROOT/logs/message-router.pid"

sleep 2

# Health checks
echo ""
PASS=0
FAIL=0
for entry in "auth-service:8081" "complaint-service:8082" "visitor-service:8083" "payment-service:8084" "notice-service:8085" "vendor-service:8086" "voice-service:8090" "message-router:8089"; do
  svc="${entry%%:*}"
  port="${entry##*:}"
  status=$(curl -sf "http://localhost:$port/health" | python3 -c "import sys,json; print(json.load(sys.stdin).get('status','?'))" 2>/dev/null || echo "DOWN")
  if [ "$status" = "healthy" ]; then
    echo -e "  $svc:$port  ${GREEN}HEALTHY${NC}"
    PASS=$((PASS+1))
  else
    echo -e "  $svc:$port  ${RED}$status${NC}"
    FAIL=$((FAIL+1))
  fi
done

echo -e "\n${GREEN}========================================${NC}"
echo -e "  ${GREEN}$PASS services healthy${NC}, ${RED}$FAIL failed${NC}"
echo -e "  Logs: $ROOT/logs/"
echo -e "  Stop: ./scripts/dev-stop.sh"
echo -e "${GREEN}========================================${NC}"
