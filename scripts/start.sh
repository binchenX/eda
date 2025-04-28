#!/bin/bash

set -e  # Exit on error

# Get host IP address (for macOS)
HOST_IP=$(ifconfig | grep "inet " | grep -v 127.0.0.1 | awk '{print $2}' | head -n1)
if [ -z "$HOST_IP" ]; then
    echo "Failed to get host IP ✗"
    exit 1
fi

# Update docker-compose.yml with host IP
echo "Configuring Kafka advertised listeners with host IP: $HOST_IP"
sed -i.bak "s/KAFKA_ADVERTISED_LISTENERS: PLAINTEXT:\/\/[0-9.]*:9092/KAFKA_ADVERTISED_LISTENERS: PLAINTEXT:\/\/$HOST_IP:9092/" docker-compose.yml
sed -i.bak "s/CONNECT_BOOTSTRAP_SERVERS: \"[0-9.]*:9092\"/CONNECT_BOOTSTRAP_SERVERS: \"$HOST_IP:9092\"/" docker-compose.yml
rm docker-compose.yml.bak

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
