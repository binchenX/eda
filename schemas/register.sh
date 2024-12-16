#!/bin/bash

# Function to register a schema
register_schema() {
  local schema_file=$1
  local subject_name=$2
  local schema_registry_url=$3

  # Read the schema from the file
  local schema=$(cat "$schema_file")

  # Register the schema with the Schema Registry
  local response=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
  --data "{\"schema\": $(jq -Rs . <<< "$schema")}" \
  "$schema_registry_url/subjects/$subject_name/versions")

  # Check the response code
  if [ "$response" -eq 200 ]; then
    echo "Schema for subject '$subject_name' registered successfully."
  else
    echo "Failed to register schema for subject '$subject_name'. HTTP response code: $response"
  fi
}

# Function to list all subjects and their schema versions with IDs
list_subjects() {
  local schema_registry_url=$1

  # Fetch all subjects
  local subjects=$(curl -s "$schema_registry_url/subjects")

  # Print table header
  printf "%-30s %-10s %-10s\n" "Subject" "Version" "Schema ID"
  printf "%-30s %-10s %-10s\n" "-------" "-------" "--------"

  for subject in $(echo "$subjects" | jq -r '.[]'); do
    # Fetch all versions for the subject
    local versions=$(curl -s "$schema_registry_url/subjects/$subject/versions")
    for version in $(echo "$versions" | jq -r '.[]'); do
      # Fetch the schema ID for each version
      local schema_id=$(curl -s "$schema_registry_url/subjects/$subject/versions/$version" | jq -r '.id')
      # Print table row
      printf "%-30s %-10s %-10s\n" "$subject" "$version" "$schema_id"
    done
  done
}

# Schema Registry URL
SCHEMA_REGISTRY_URL="http://localhost:8081"

# Register schemas
register_schema "order.avsc" "OrderEventTopic-value" "$SCHEMA_REGISTRY_URL"
# Add more schemas here if needed
# register_schema "another_schema.avsc" "AnotherTopic-value" "$SCHEMA_REGISTRY_URL"

# List all subjects and their schema versions with IDs
list_subjects "$SCHEMA_REGISTRY_URL"
