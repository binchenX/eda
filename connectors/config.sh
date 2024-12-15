#!/bin/bash

# Function to check if Kafka Connect is running
check_connect() {
  local retries=10
  local wait=5

  for ((i=1; i<=retries; i++)); do
    echo "Checking Kafka Connect status (attempt $i/$retries)..."
    response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8083/connectors)
    if [ "$response" -eq 200 ]; then
      echo "Kafka Connect is running."
      return 0
    fi
    echo "Kafka Connect is not running. Waiting for $wait seconds..."
    sleep $wait
  done

  echo "Kafka Connect did not start within the expected time."
  return 1
}

# Function to deploy a connector
deploy_connector() {
  local connector_file=$1
  local connector_name=$(jq -r '.name' "$connector_file")

  echo "Deploying connector: $connector_name"
  curl -X POST -H "Content-Type: application/json" --data @"$connector_file" http://localhost:8083/connectors
  echo ""
}

# Ensure Kafka Connect is running before deploying connectors
if check_connect; then
  # Loop through all JSON files in the current directory and deploy them
  for connector_file in *.json; do
    deploy_connector "$connector_file"
  done
else
  echo "Failed to deploy connectors because Kafka Connect is not running."
  exit 1
fi
