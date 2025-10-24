#!/bin/bash
# Add SSE extensions to swagger.json after generation

SWAGGER_FILE="src/docs/openapi2/swagger.json"

if [ ! -f "$SWAGGER_FILE" ]; then
    echo "Error: $SWAGGER_FILE not found"
    exit 1
fi

echo "Adding SSE extensions to $SWAGGER_FILE..."

# Use jq to add x-is-streaming-api to SSE endpoints
jq '
  .paths |= with_entries(
    .value |= with_entries(
      if .value.produces? and (.value.produces | index("text/event-stream")) then
        .value."x-is-streaming-api" = true
      else
        .
      end
    )
  )
' "$SWAGGER_FILE" > "${SWAGGER_FILE}.tmp" && mv "${SWAGGER_FILE}.tmp" "$SWAGGER_FILE"

echo "✅ SSE extensions added successfully"