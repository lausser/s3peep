#!/bin/bash
# Integration test for s3peep web UI
# Creates bucket, uploads files, verifies via HTTP

set -e

echo "=== S3 File Browser Integration Test ==="

# Configuration
MINIO_ENDPOINT="http://localhost:9000"
MINIO_ACCESS_KEY="minioadmin"
MINIO_SECRET_KEY="minioadmin"
S3PEEP_PORT="18080"
S3PEEP_ENDPOINT="http://localhost:${S3PEEP_PORT}"
TEST_BUCKET="test-bucket"
CONFIG_DIR="/tmp/s3peep-test-config"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Helper functions
log() { echo -e "${GREEN}[TEST]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Step 1: Start MinIO if not running
log "Step 1: Checking MinIO..."
if ! curl -s ${MINIO_ENDPOINT}/minio/health/live > /dev/null 2>&1; then
    log "Starting MinIO container..."
    podman run -d --name s3peep-test-minio \
        -p 9000:9000 \
        -e MINIO_ROOT_USER=${MINIO_ACCESS_KEY} \
        -e MINIO_ROOT_PASSWORD=${MINIO_SECRET_KEY} \
        minio/minio:latest server /data --console-address ":9001" \
        || error "Failed to start MinIO"
    
    # Wait for MinIO to be ready
    log "Waiting for MinIO to be ready..."
    for i in {1..30}; do
        if curl -s ${MINIO_ENDPOINT}/minio/health/live > /dev/null 2>&1; then
            log "MinIO is ready"
            break
        fi
        sleep 1
    done
else
    log "MinIO is already running"
fi

# Step 2: Install mc (MinIO client) if needed
log "Step 2: Setting up MinIO client..."
if ! command -v mc &> /dev/null; then
    log "Downloading mc..."
    curl -s -o /tmp/mc https://dl.min.io/client/mc/release/linux-amd64/mc
    chmod +x /tmp/mc
    MC="/tmp/mc"
else
    MC="mc"
fi

# Configure mc alias
$MC alias set local ${MINIO_ENDPOINT} ${MINIO_ACCESS_KEY} ${MINIO_SECRET_KEY} --insecure 2>/dev/null || true

# Step 3: Create test bucket
log "Step 3: Creating test bucket '${TEST_BUCKET}'..."
$MC mb local/${TEST_BUCKET} --ignore-existing --insecure 2>/dev/null || log "Bucket may already exist"

# Step 4: Upload test files
log "Step 4: Uploading test files..."

# Create test files
mkdir -p /tmp/s3peep-test-files
echo "Hello, World!" > /tmp/s3peep-test-files/hello.txt
echo "This is a test document" > /tmp/s3peep-test-files/readme.md
echo "col1,col2,col3" > /tmp/s3peep-test-files/data.csv
echo "Sample config" > /tmp/s3peep-test-files/config.json

# Create a subdirectory
mkdir -p /tmp/s3peep-test-files/documents
echo "Document 1" > /tmp/s3peep-test-files/documents/doc1.txt
echo "Document 2" > /tmp/s3peep-test-files/documents/doc2.txt

# Upload files
$MC cp /tmp/s3peep-test-files/hello.txt local/${TEST_BUCKET}/ --insecure
$MC cp /tmp/s3peep-test-files/readme.md local/${TEST_BUCKET}/ --insecure
$MC cp /tmp/s3peep-test-files/data.csv local/${TEST_BUCKET}/ --insecure
$MC cp /tmp/s3peep-test-files/config.json local/${TEST_BUCKET}/ --insecure
$MC cp -r /tmp/s3peep-test-files/documents local/${TEST_BUCKET}/ --insecure

log "Files uploaded successfully"

# Step 5: Create s3peep config
log "Step 5: Creating s3peep configuration..."
mkdir -p ${CONFIG_DIR}
cat > ${CONFIG_DIR}/config.json <<EOF
{
  "active_profile": "test",
  "profiles": [
    {
      "name": "test",
      "region": "us-east-1",
      "access_key_id": "${MINIO_ACCESS_KEY}",
      "secret_access_key": "${MINIO_SECRET_KEY}",
      "endpoint_url": "${MINIO_ENDPOINT}",
      "bucket": "${TEST_BUCKET}"
    }
  ]
}
EOF

log "Config created at ${CONFIG_DIR}/config.json"

# Step 6: Build s3peep
log "Step 6: Building s3peep..."
cd /home/lausser/git/s3peep
podman run --rm -v $(pwd):/app -w /app golang:1.24-alpine go build -o s3peep ./cmd/s3peep || error "Build failed"
log "Build successful"

# Step 7: Start s3peep server
log "Step 7: Starting s3peep server on port ${S3PEEP_PORT}..."
# Kill any existing s3peep on this port
pkill -f "s3peep serve --port ${S3PEEP_PORT}" 2>/dev/null || true
sleep 1

# Start s3peep
./s3peep --config ${CONFIG_DIR}/config.json serve --port ${S3PEEP_PORT} &
S3PEEP_PID=$!

# Wait for server to start
log "Waiting for s3peep server to start..."
for i in {1..30}; do
    if curl -s ${S3PEEP_ENDPOINT} > /dev/null 2>&1; then
        log "S3peep server is ready"
        break
    fi
    sleep 1
done

# Check if server started
if ! curl -s ${S3PEEP_ENDPOINT} > /dev/null 2>&1; then
    error "S3peep server failed to start"
fi

# Step 8: Test HTTP endpoints
log "Step 8: Testing HTTP endpoints..."

# Get the token from the server output
TOKEN=$(curl -s ${S3PEEP_ENDPOINT} | grep -oP '(?<=/)[a-zA-Z0-9_-]{40,}(?=")' | head -1)

if [ -z "$TOKEN" ]; then
    # Try to get token from the page
    log "Trying to extract token from response..."
    curl -s ${S3PEEP_ENDPOINT} | head -20
fi

# Try accessing API directly
log "Testing API access..."
API_URL="${S3PEEP_ENDPOINT}/${TOKEN}/api"

# Test buckets endpoint
log "Testing /buckets endpoint..."
BUCKETS_RESPONSE=$(curl -s "${API_URL}/buckets" || echo "FAILED")

if echo "$BUCKETS_RESPONSE" | grep -q "test-bucket"; then
    log "✓ Buckets endpoint working - found test-bucket"
else
    error "Buckets endpoint not working: $BUCKETS_RESPONSE"
fi

# Test profile endpoint
log "Testing /profile endpoint..."
PROFILE_RESPONSE=$(curl -s "${API_URL}/profile" || echo "FAILED")

if echo "$PROFILE_RESPONSE" | grep -q "test"; then
    log "✓ Profile endpoint working"
else
    error "Profile endpoint not working: $PROFILE_RESPONSE"
fi

# Test objects endpoint
log "Testing /objects endpoint..."
OBJECTS_RESPONSE=$(curl -s "${API_URL}/buckets/${TEST_BUCKET}/objects" || echo "FAILED")

if echo "$OBJECTS_RESPONSE" | grep -q "hello.txt"; then
    log "✓ Objects endpoint working - found hello.txt"
else
    error "Objects endpoint not working: $OBJECTS_RESPONSE"
fi

# Step 9: Test web UI
log "Step 9: Testing web UI..."

# Test HTML page
HTML_RESPONSE=$(curl -s ${S3PEEP_ENDPOINT})

if echo "$HTML_RESPONSE" | grep -q "S3 File Browser"; then
    log "✓ Web UI HTML loads correctly"
else
    error "Web UI HTML not loading correctly"
fi

if echo "$HTML_RESPONSE" | grep -q "bucket-list"; then
    log "✓ Bucket list element found in HTML"
else
    error "Bucket list element not found"
fi

# Step 10: Summary
log "=== Test Summary ==="
log "✓ MinIO running at ${MINIO_ENDPOINT}"
log "✓ Test bucket '${TEST_BUCKET}' created with files"
log "✓ S3peep server running at ${S3PEEP_ENDPOINT}"
log "✓ API endpoints responding"
log "✓ Web UI loading"

echo ""
echo "Access the web UI at: ${S3PEEP_ENDPOINT}"
echo "API base URL: ${API_URL}"
echo ""

# Cleanup function
cleanup() {
    log "Cleaning up..."
    kill $S3PEEP_PID 2>/dev/null || true
    rm -rf ${CONFIG_DIR} /tmp/s3peep-test-files
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Keep server running for manual testing
log "Server will keep running for manual testing. Press Ctrl+C to stop."
wait $S3PEEP_PID
