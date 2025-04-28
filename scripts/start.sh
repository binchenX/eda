#!/bin/bash

set -e  # Exit on error

echo "Starting services..."
docker-compose up -d

echo "Waiting for services to be ready..."
sleep 20  # Increased wait time to ensure services are fully up

echo "Registering schema..."
./scripts/register-schema.sh

echo "Verifying schema registration..."
# Try multiple times to verify the schema
for i in {1..5}; do
    echo "Verification attempt $i..."
    if curl -s http://localhost:8081/subjects/OrderEventTopic-value/versions/latest | jq . > /dev/null 2>&1; then
        echo "Schema verification successful!"
        curl -s http://localhost:8081/subjects/OrderEventTopic-value/versions/latest | jq .
        exit 0
    fi
    sleep 5
done

echo "Failed to verify schema registration after multiple attempts"
exit 1

echo "Setup complete!" 
