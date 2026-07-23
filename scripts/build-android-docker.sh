#!/bin/bash
# Build Android APK using Docker-based Flutter/Android SDK builder
# Usage: bash build-android-docker.sh [path-to-project]
#
# The project should be the isle/ directory containing isle_app/ and shared/
# If no path given, uses the USB mount point.

set -e

IMAGE="isle-android-builder"
PROJECT="${1:-/run/media/tomas/SIMPLEX-USB/Projects/isle}"

if [ ! -d "$PROJECT/isle_app" ]; then
    echo "❌ Project not found at $PROJECT"
    echo "   Usage: $0 /path/to/isle"
    exit 1
fi

echo "=== Building Android APK for The Isle ==="
echo "Project: $PROJECT"

# Build Docker image if not exists
if ! docker image inspect $IMAGE &>/dev/null; then
    echo "=== Building Docker image (first run, may take 15-20 min) ==="
    cd "$(dirname "$0")/../docker/android-builder"
    docker build -t $IMAGE .
else
    echo "✅ Docker image exists"
fi

echo "=== Running Docker build ==="
mkdir -p "$PROJECT/build-output"
docker run --rm -v "$PROJECT:/app" $IMAGE

echo ""
echo "=== Done! ==="
echo "APK: $PROJECT/build-output/isle-app-release.apk"
ls -lh "$PROJECT/build-output/isle-app-release.apk" 2>/dev/null || echo "   (build failed)"
