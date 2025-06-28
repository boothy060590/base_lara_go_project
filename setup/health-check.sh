#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "[6/6] Waiting for Sentry to be ready..."
echo "   This may take 30-60 seconds on first startup. Please be patient..."
echo "   (You may see 502 errors in your browser during this time - this is normal)"

# Wait for Sentry to be ready
MAX_ATTEMPTS=60
ATTEMPT=1

while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
  if curl -k -s -f -o /dev/null --connect-timeout 5 --max-time 10 https://sentry.baselaragoproject.test/ >/dev/null 2>&1; then
    echo "   ✓ Sentry is ready!"
    break
  fi
  
  if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
    echo "   ⚠️  Sentry health check timed out after $MAX_ATTEMPTS attempts."
    echo "   The stack is running, but Sentry may still be starting up."
    echo "   You can check the status with: docker logs sentry"
    break
  fi
  
  echo "   Waiting... (attempt $ATTEMPT/$MAX_ATTEMPTS)"
  sleep 2
  ATTEMPT=$((ATTEMPT + 1))
done

echo -e "\n🎉 Setup complete! The stack is now running."
echo -e "\n📋 Available services:"
echo "   • Frontend: https://baselaragoproject.test"
echo "   • API: https://api.baselaragoproject.test"
echo "   • Sentry: https://sentry.baselaragoproject.test"
echo "   • MailHog: https://mail.baselaragoproject.test"
echo "   • SQS UI: https://sqs.baselaragoproject.test"
echo -e "\n🔧 Useful commands:"
echo "   • View logs: docker-compose logs -f [service_name]"
echo "   • Stop stack: docker-compose down"
echo "   • Clean slate: make clean" 