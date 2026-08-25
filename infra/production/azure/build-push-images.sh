#!/usr/bin/env bash
set -euo pipefail

ACR_SERVER="${ACR_SERVER:-goridecr.azurecr.io}"
ACR_NAME="${ACR_NAME:-${ACR_SERVER%%.*}}"
PLATFORM="linux/amd64"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCKER_DIR="$SCRIPT_DIR/docker"
SERVICES=(api-gateway trip-service driver-service payment-service web)

if [ "$#" -gt 0 ]; then
  SERVICES=("$@")
fi

echo "==> Logging into ACR: $ACR_NAME"
az acr login --name "$ACR_NAME"

FAILED=()
for svc in "${SERVICES[@]}"; do
  IMAGE="$ACR_SERVER/goride/$svc:latest"
  DOCKERFILE="$DOCKER_DIR/$svc.Dockerfile"

  if [ ! -f "$DOCKERFILE" ]; then
    echo "!! Dockerfile not found: $DOCKERFILE"
    FAILED+=("$svc")
    continue
  fi

  echo ""
  echo "==> Building $IMAGE"
  BUILD_ARGS=()
  if [ "$svc" = "web" ]; then
    ENV_FILE="$DOCKER_DIR/web.env.build"
    if [ -f "$ENV_FILE" ]; then
      while IFS='=' read -r key value; do
        [ -z "$key" ] && continue
        case "$key" in \#*) continue ;; esac
        BUILD_ARGS+=(--build-arg "$key=$value")
      done < <(grep -v '^[[:space:]]*$' "$ENV_FILE")
    else
      echo "!! Missing $ENV_FILE — create it with NEXT_PUBLIC_* values before building web"
      FAILED+=("$svc")
      continue
    fi
  fi

  if ! docker build --platform "$PLATFORM" "${BUILD_ARGS[@]}" -t "$IMAGE" -f "$DOCKERFILE" "$SCRIPT_DIR/../../.."; then
    echo "!! Build failed: $svc"
    FAILED+=("$svc")
    continue
  fi

  echo "==> Pushing $IMAGE"
  if ! docker push "$IMAGE"; then
    echo "!! Push failed: $svc"
    FAILED+=("$svc")
    continue
  fi

  echo "==> Done: $svc"
done

echo ""
if [ "${#FAILED[@]}" -gt 0 ]; then
  echo "Finished WITH FAILURES: ${FAILED[*]}"
  exit 1
fi

echo "All images pushed successfully:"
az acr repository list --name "$ACR_NAME" -o table
