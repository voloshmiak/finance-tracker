#!/bin/bash

set -e

IMAGE_NAME="finance-tracker-dev"
CONTAINER_NAME="finance-tracker"

docker rm -f "$CONTAINER_NAME" 2>/dev/null || true

docker run -it \
  --name "$CONTAINER_NAME" \
  --env-file .env \
  -p 80:80 \
  -v "$(pwd)":/app \
  "$IMAGE_NAME"