#!/bin/bash

# Colors
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${GREEN}>>> Building Aeterna Core...${NC}"
go build -o bin/aeterna cmd/aeterna/main.go

echo -e "${GREEN}>>> Preparing Configuration...${NC}"
cat > aeterna_test.yaml <<EOF
version: "v1"
service:
  name: "test-agent"
  command: ["python3", "examples/agent_mock.py"]
  env:
    - "PORT=8080"
    - "PYTHONUNBUFFERED=1"

orchestration:
  strategy: "canary"
  soak_time: "5s"
  state_handoff:
    enabled: true
    socket_path: "/tmp/aeterna_test.sock"
    timeout: "3s"

observability:
  metrics_port: ":9091"
  log_level: "debug"
EOF

# Ensure clean slate
pkill -f "aeterna" || true
pkill -f "agent_mock.py" || true

echo -e "${GREEN}>>> Starting Aeterna (Cold Start)...${NC}"
./bin/aeterna start -c aeterna_test.yaml &
PID=$!
sleep 3

echo -e "${GREEN}>>> Step 1: Injecting Memory into Agent v1...${NC}"
curl -X POST -d "Hello, I am user 1" http://localhost:8080
curl -X POST -d "This is important context" http://localhost:8080

echo -e "\n${GREEN}>>> Step 2: Triggering Hot Relay (SIGHUP)...${NC}"
# Simulate updating the code (we use the same script, but PID will change)
kill -HUP $PID

echo -e "${GREEN}>>> Waiting for SRP Handover and Soaking...${NC}"
sleep 8

echo -e "${GREEN}>>> Step 3: Verifying Memory Persistence in Agent v2...${NC}"
RESPONSE=$(curl -s http://localhost:8080)
echo "Response from Agent: $RESPONSE"

if [[ "$RESPONSE" == *"This is important context"* ]]; then
    echo -e "${GREEN}SUCCESS: Context preserved across process restart!${NC}"
else
    echo -e "\033[0;31mFAILURE: Context lost!${NC}"
    exit 1
fi

echo -e "${GREEN}>>> Cleaning up...${NC}"
kill $PID
rm aeterna_test.yaml