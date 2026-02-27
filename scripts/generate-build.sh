#!/usr/bin/env bash
set -euo pipefail

# White-label Flutter build orchestrator
# Usage: ./generate-build.sh <dealer-slug> [android|ios|both]

DEALER_SLUG="${1:?Usage: $0 <dealer-slug> [android|ios|both]}"
PLATFORM="${2:-both}"
API_BASE_URL="${API_BASE_URL:-https://api.lumbernow.com/v1}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
MOBILE_DIR="$PROJECT_DIR/mobile"

echo "==> Fetching dealer config for: $DEALER_SLUG"

CONFIG=$(curl -sf "$API_BASE_URL/tenant/config?slug=$DEALER_SLUG")
if [ -z "$CONFIG" ]; then
  echo "ERROR: Failed to fetch config for dealer: $DEALER_SLUG"
  exit 1
fi

DEALER_NAME=$(echo "$CONFIG" | jq -r '.name')
DEALER_ID=$(echo "$CONFIG" | jq -r '.dealer_id')
PRIMARY_COLOR=$(echo "$CONFIG" | jq -r '.primary_color' | sed 's/#//')
SECONDARY_COLOR=$(echo "$CONFIG" | jq -r '.secondary_color' | sed 's/#//')
LOGO_URL=$(echo "$CONFIG" | jq -r '.logo_url')

echo "==> Building for: $DEALER_NAME ($DEALER_ID)"
echo "    Primary: #$PRIMARY_COLOR | Secondary: #$SECONDARY_COLOR"

# Inject assets (logos, app name, etc.)
"$SCRIPT_DIR/inject-assets.sh" "$DEALER_SLUG" "$DEALER_NAME" "$LOGO_URL"

DART_DEFINES=(
  "--dart-define=API_BASE_URL=$API_BASE_URL"
  "--dart-define=TENANT_SLUG=$DEALER_SLUG"
  "--dart-define=PRIMARY_COLOR=$PRIMARY_COLOR"
  "--dart-define=SECONDARY_COLOR=$SECONDARY_COLOR"
  "--dart-define=APP_NAME=$DEALER_NAME"
)

cd "$MOBILE_DIR"

if [ "$PLATFORM" = "android" ] || [ "$PLATFORM" = "both" ]; then
  echo "==> Building Android APK..."
  flutter build apk --release "${DART_DEFINES[@]}"
  echo "==> Android APK: $MOBILE_DIR/build/app/outputs/flutter-apk/app-release.apk"
fi

if [ "$PLATFORM" = "ios" ] || [ "$PLATFORM" = "both" ]; then
  echo "==> Building iOS..."
  flutter build ios --release --no-codesign "${DART_DEFINES[@]}"
  echo "==> iOS build complete"
fi

echo "==> Build complete for $DEALER_NAME"
