#!/bin/bash

set -e  # Exit on error

# Function to check if Schema Registry is ready
check_schema_registry() {
    local retries=10
    local wait=5

    for ((i=1; i<=retries; i++)); do
        echo "Checking Schema Registry status (attempt $i/$retries)..."
        response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/subjects)
        if [ "$response" -eq 200 ]; then
            echo "Schema Registry is running."
            return 0
        fi
        echo "Schema Registry is not ready. Waiting for $wait seconds..."
        sleep $wait
    done

    echo "Schema Registry did not start within the expected time."
    return 1
}

# Function to register schema
register_schema() {
    local schema='{
        "schema": "{\"type\":\"record\",\"name\":\"Order\",\"namespace\":\"com.example\",\"fields\":[{\"name\":\"orderId\",\"type\":\"string\"},{\"name\":\"userId\",\"type\":\"string\"},{\"name\":\"items\",\"type\":{\"type\":\"array\",\"items\":\"string\"}}]}"
    }'

    echo "Registering schema..."
    echo "Schema content: $schema"
    
    # First, try to get the current schema to see if it exists
    echo "Checking if schema already exists..."
    response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/subjects/OrderEventTopic-value/versions/latest)
    
    if [ "$response" -eq 200 ]; then
        echo "Schema already exists. Content:"
        curl -s http://localhost:8081/subjects/OrderEventTopic-value/versions/latest | jq .
        return 0
    elif [ "$response" -eq 404 ]; then
        echo "Schema does not exist. Proceeding with registration..."
    else
        echo "Unexpected response code: $response"
        return 1
    fi

    # Register the schema
    echo "Registering new schema..."
    response=$(curl -v -s -o /dev/null -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/vnd.schemaregistry.v1+json" \
        --data "$schema" \
        http://localhost:8081/subjects/OrderEventTopic-value/versions)

    echo "Registration response code: $response"
    
    if [ "$response" -eq 200 ]; then
        echo "Schema registered successfully."
        # Verify the registration
        echo "Verifying registration..."
        curl -s http://localhost:8081/subjects/OrderEventTopic-value/versions/latest | jq .
        return 0
    else
        echo "Failed to register schema. Status code: $response"
        return 1
    fi
}

# Main execution
echo "Starting schema registration process..."
if check_schema_registry; then
    register_schema
    exit $?
else
    echo "Failed to connect to Schema Registry"
    exit 1
fi 
