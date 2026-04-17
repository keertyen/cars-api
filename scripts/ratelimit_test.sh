#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# Rate-limit stress test
# 4 dummy IPs each fire 25 requests in parallel (should hit 429 after burst)
# 1 genuine IP fires 1 request (should always get through)
# ---------------------------------------------------------------------------

BASE="http://localhost:8080"
GENUINE_IP="192.168.1.100"
DUMMY_IPS=("10.0.0.1" "10.0.0.2" "10.0.0.3" "10.0.0.4")
REQUESTS_PER_IP=25  # burst is 20, so last 5 from each dummy IP should 429

echo "======================================================"
echo " Rate Limit Test"
echo " Dummy IPs  : ${DUMMY_IPS[*]}"
echo " Requests   : ${REQUESTS_PER_IP} per dummy IP"
echo " Genuine IP : ${GENUINE_IP}"
echo "======================================================"
echo ""

# ── 4 dummy IPs — parallel flood ──────────────────────────────────────────
echo "[ Flooding from 4 dummy IPs... ]"

for ip in "${DUMMY_IPS[@]}"; do
  for i in $(seq 1 $REQUESTS_PER_IP); do
    curl -s -o /dev/null \
         -w "IP: ${ip} | req #${i} | status: %{http_code}\n" \
         -H "X-Forwarded-For: ${ip}" \
         "${BASE}/v1/cars" &
  done
done

# ── 1 genuine request — runs in parallel with the flood ───────────────────
echo ""
echo "[ Genuine request firing simultaneously... ]"
GENUINE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -H "X-Forwarded-For: ${GENUINE_IP}" \
  "${BASE}/v1/cars")

# wait for all background curl jobs
wait

# ── Print genuine result ───────────────────────────────────────────────────
echo ""
echo "======================================================"
echo " GENUINE REQUEST RESULT (${GENUINE_IP})"
echo "======================================================"
HTTP_STATUS=$(echo "$GENUINE_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
BODY=$(echo "$GENUINE_RESPONSE" | grep -v "HTTP_STATUS")

echo "Status : ${HTTP_STATUS}"
echo "Body   : ${BODY}" | python3 -m json.tool 2>/dev/null || echo "Body: ${BODY}"

echo ""
echo "======================================================"
if [ "$HTTP_STATUS" = "200" ]; then
  echo " ✓ Genuine request PASSED through rate limiter"
else
  echo " ✗ Genuine request was blocked (status: ${HTTP_STATUS})"
fi
echo "======================================================"
