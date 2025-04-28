#!/bin/bash

set -e  # Exit on error

# Check Schema Registry
for i in {1..10}; do
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/subjects | grep -q "200"; then
        break
    fi
    sleep 5
done

# Register schema
schema='{
    "schema": "{\"type\":\"record\",\"name\":\"Order\",\"namespace\":\"com.example\",\"fields\":[{\"name\":\"orderId\",\"type\":\"string\"},{\"name\":\"userId\",\"type\":\"string\"},{\"name\":\"items\",\"type\":{\"type\":\"array\",\"items\":\"string\"}}]}"
}'

response=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST \
    -H "Content-Type: application/vnd.schemaregistry.v1+json" \
    --data "$schema" \
    http://localhost:8081/subjects/OrderEventTopic-value/versions)

if [ "$response" = "200" ] || [ "$response" = "409" ]; then
    exit 0
else
    echo "Schema registration failed with status $response"
    exit 1
fi 
