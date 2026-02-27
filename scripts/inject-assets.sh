#!/usr/bin/env bash
set -euo pipefail

# Inject dealer-specific branding into Flutter project
# Usage: ./inject-assets.sh <dealer-slug> <dealer-name> <logo-url>

DEALER_SLUG="${1:?Usage: $0 <dealer-slug> <dealer-name> <logo-url>}"
DEALER_NAME="${2:?Dealer name required}"
LOGO_URL="${3:-}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
MOBILE_DIR="$PROJECT_DIR/mobile"
ASSETS_DIR="$MOBILE_DIR/assets/images"

echo "==> Injecting assets for: $DEALER_NAME ($DEALER_SLUG)"

# Download logo if URL provided
if [ -n "$LOGO_URL" ] && [ "$LOGO_URL" != "null" ] && [ "$LOGO_URL" != "" ]; then
  echo "    Downloading logo..."
  curl -sf -o "$ASSETS_DIR/dealer_logo.png" "$LOGO_URL" || echo "    Warning: Failed to download logo"
fi

# Update Android app name in AndroidManifest.xml
ANDROID_MANIFEST="$MOBILE_DIR/android/app/src/main/AndroidManifest.xml"
if [ -f "$ANDROID_MANIFEST" ]; then
  echo "    Updating Android app name..."
  sed -i "s/android:label=\"[^\"]*\"/android:label=\"$DEALER_NAME\"/" "$ANDROID_MANIFEST"
fi

# Update iOS app name in Info.plist
IOS_PLIST="$MOBILE_DIR/ios/Runner/Info.plist"
if [ -f "$IOS_PLIST" ]; then
  echo "    Updating iOS app name..."
  # Replace CFBundleDisplayName value
  sed -i "/<key>CFBundleDisplayName<\/key>/{n;s/<string>[^<]*<\/string>/<string>$DEALER_NAME<\/string>/;}" "$IOS_PLIST"
  sed -i "/<key>CFBundleName<\/key>/{n;s/<string>[^<]*<\/string>/<string>$DEALER_NAME<\/string>/;}" "$IOS_PLIST"
fi

# Update Android package name for unique installs (optional)
PACKAGE_SLUG=$(echo "$DEALER_SLUG" | tr '-' '_' | tr '[:upper:]' '[:lower:]')
echo "    Package suffix: $PACKAGE_SLUG"

echo "==> Asset injection complete"
