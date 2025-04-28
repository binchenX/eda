#!/bin/bash

set -e  # Exit on error

echo "Starting services..."
docker-compose up -d

echo "Waiting for services and registering schema..."
sleep 20
./scripts/register-schema.sh

# Verify schema registration
for i in {1..5}; do
    if curl -s http://localhost:8081/subjects/OrderEventTopic-value/versions/latest | jq . > /dev/null 2>&1; then
        echo "Setup complete ✓"
        exit 0
    fi
    sleep 5
done

echo "Failed to verify schema registration ✗"
exit 1 
