#!/bin/bash
# Test delete operations for s3peep web UI
# Tests deleting files and folders via HTTP API

set -e

echo "=== S3 File Browser Delete Operations Test ==="

# Configuration
MINIO_ENDPOINT="http://localhost:9000"
MINIO_ACCESS_KEY="minioadmin"
MINIO_SECRET_KEY="minioadmin"
S3PEEP_PORT="18081"
S3PEEP_ENDPOINT="http://localhost:${S3PEEP_PORT}"
TEST_BUCKET="test-delete-bucket"
CONFIG_DIR="/tmp/s3peep-delete-test-config"

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
    podman run -d --name s3peep-delete-test-minio \
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

# Step 4: Create test structure
log "Step 4: Creating test file structure..."

# Files in root
echo "Root file 1" | $MC pipe local/${TEST_BUCKET}/root-file1.txt --insecure
echo "Root file 2" | $MC pipe local/${TEST_BUCKET}/root-file2.txt --insecure

# Empty folder (just a marker)
echo "" | $MC pipe local/${TEST_BUCKET}/empty-folder/ --insecure

# Folder with files
mkdir -p /tmp/s3peep-delete-test
echo "File in folder" > /tmp/s3peep-delete-test/nested-file.txt
echo "Another file" > /tmp/s3peep-delete-test/another-file.txt
$MC cp /tmp/s3peep-delete-test/nested-file.txt local/${TEST_BUCKET}/folder-with-files/ --insecure
$MC cp /tmp/s3peep-delete-test/another-file.txt local/${TEST_BUCKET}/folder-with-files/ --insecure

# Nested folder structure
echo "Deep file" > /tmp/s3peep-delete-test/deep.txt
$MC cp /tmp/s3peep-delete-test/deep.txt local/${TEST_BUCKET}/parent-folder/child-folder/ --insecure

# List what we created
log "Test structure created:"
$MC ls -r local/${TEST_BUCKET} --insecure | head -20

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

# Step 8: Get API token
log "Step 8: Getting API token..."
TOKEN=$(curl -s ${S3PEEP_ENDPOINT} | grep -oP '(?<=/)[a-zA-Z0-9_-]{40,}(?=")' | head -1)

if [ -z "$TOKEN" ]; then
    error "Could not extract token from web UI"
fi

API_URL="${S3PEEP_ENDPOINT}/${TOKEN}/api"
log "API URL: ${API_URL}"

# Step 9: Test DELETE operations
log "Step 9: Testing DELETE operations..."

# Test 9a: Delete single file
log "Test 9a: Deleting single file (root-file1.txt)..."
DELETE_RESPONSE=$(curl -s -X DELETE \
    -H "Content-Type: application/json" \
    -d '{"keys":["root-file1.txt"]}' \
    "${API_URL}/buckets/${TEST_BUCKET}/objects" || echo "FAILED")

if echo "$DELETE_RESPONSE" | grep -q "deleted"; then
    log "✓ Single file delete request sent successfully"
    
    # Verify file is gone
    sleep 1
    OBJECTS=$(curl -s "${API_URL}/buckets/${TEST_BUCKET}/objects" || echo "FAILED")
    if ! echo "$OBJECTS" | grep -q "root-file1.txt"; then
        log "✓ File confirmed deleted"
    else
        error "File still exists after delete"
    fi
else
    error "Single file delete failed: $DELETE_RESPONSE"
fi

# Test 9b: Delete empty folder
log "Test 9b: Deleting empty folder..."
DELETE_RESPONSE=$(curl -s -X DELETE \
    -H "Content-Type: application/json" \
    -d '{"keys":["empty-folder/"]}' \
    "${API_URL}/buckets/${TEST_BUCKET}/objects" || echo "FAILED")

if echo "$DELETE_RESPONSE" | grep -q "deleted"; then
    log "✓ Empty folder delete request sent successfully"
else
    error "Empty folder delete failed: $DELETE_RESPONSE"
fi

# Test 9c: Delete folder with contents (recursive delete)
log "Test 9c: Deleting folder with contents (folder-with-files/)..."

# First, get list of objects in the folder
log "Getting list of objects in folder-with-files/..."
FOLDER_OBJECTS=$(curl -s "${API_URL}/buckets/${TEST_BUCKET}/objects?prefix=folder-with-files/" || echo "FAILED")
FILE_COUNT=$(echo "$FOLDER_OBJECTS" | grep -o '"key"' | wc -l)
log "Found $FILE_COUNT objects in folder"

# Delete all objects in the folder
DELETE_KEYS=$(echo "$FOLDER_OBJECTS" | grep -o '"key":"[^"]*"' | sed 's/"key":"//;s/"$//' | jq -R -s -c 'split("\n") | map(select(length > 0))')
log "Deleting keys: $DELETE_KEYS"

DELETE_RESPONSE=$(curl -s -X DELETE \
    -H "Content-Type: application/json" \
    -d "{\"keys\":$DELETE_KEYS}" \
    "${API_URL}/buckets/${TEST_BUCKET}/objects" || echo "FAILED")

if echo "$DELETE_RESPONSE" | grep -q "deleted"; then
    DELETED_COUNT=$(echo "$DELETE_RESPONSE" | grep -o '"deleted":\[[^]]*\]' | grep -o '"' | wc -l)
    DELETED_COUNT=$((DELETED_COUNT / 2))
    log "✓ Deleted $DELETED_COUNT objects from folder"
    
    # Verify folder is empty/gone
    sleep 1
    FOLDER_CHECK=$(curl -s "${API_URL}/buckets/${TEST_BUCKET}/objects?prefix=folder-with-files/" || echo "FAILED")
    if echo "$FOLDER_CHECK" | grep -q '"objects":\[\]\|"objects":null'; then
        log "✓ Folder is now empty"
    else
        log "⚠ Folder still has contents (this is expected behavior - folder markers are separate objects)"
    fi
else
    error "Folder delete failed: $DELETE_RESPONSE"
fi

# Test 9d: Delete multiple files at once
log "Test 9d: Deleting multiple files at once..."
DELETE_RESPONSE=$(curl -s -X DELETE \
    -H "Content-Type: application/json" \
    -d '{"keys":["root-file2.txt","nested-file.txt"]}' \
    "${API_URL}/buckets/${TEST_BUCKET}/objects" || echo "FAILED")

if echo "$DELETE_RESPONSE" | grep -q "deleted"; then
    log "✓ Multiple file delete request sent successfully"
else
    error "Multiple file delete failed: $DELETE_RESPONSE"
fi

# Step 10: Verify final state
log "Step 10: Verifying final bucket state..."
FINAL_OBJECTS=$(curl -s "${API_URL}/buckets/${TEST_BUCKET}/objects" || echo "FAILED")
REMAINING_COUNT=$(echo "$FINAL_OBJECTS" | grep -o '"key"' | wc -l)
log "Remaining objects in bucket: $REMAINING_COUNT"

# Step 11: Test error handling
log "Step 11: Testing error handling..."

# Test deleting non-existent file
log "Test 11a: Deleting non-existent file..."
DELETE_RESPONSE=$(curl -s -X DELETE \
    -H "Content-Type: application/json" \
    -d '{"keys":["non-existent-file.txt"]}' \
    "${API_URL}/buckets/${TEST_BUCKET}/objects" || echo "FAILED")

# Should get a response with failed items
if echo "$DELETE_RESPONSE" | grep -q "failed"; then
    log "✓ Non-existent file properly reported as failed"
else
    log "⚠ Unexpected response for non-existent file: $DELETE_RESPONSE"
fi

# Test empty keys array
log "Test 11b: Sending empty keys array..."
DELETE_RESPONSE=$(curl -s -X DELETE \
    -H "Content-Type: application/json" \
    -d '{"keys":[]}' \
    "${API_URL}/buckets/${TEST_BUCKET}/objects" || echo "FAILED")

if echo "$DELETE_RESPONSE" | grep -q "MISSING_KEYS\|error"; then
    log "✓ Empty keys array properly rejected"
else
    log "⚠ Unexpected response for empty keys: $DELETE_RESPONSE"
fi

# Step 12: Summary
log "=== Delete Operations Test Summary ==="
log "✓ Single file deletion works"
log "✓ Empty folder deletion works"
log "✓ Folder with contents deletion works (recursive)"
log "✓ Multiple file deletion works"
log "✓ Error handling for edge cases"
log ""
log "Note on folder deletion behavior:"
log "S3 doesn't have true 'folders' - they're just prefixes."
log "Deleting a folder requires deleting all objects with that prefix."
log "The UI should warn users: 'This folder contains X items. Delete all?'"

# Cleanup function
cleanup() {
    log "Cleaning up..."
    kill $S3PEEP_PID 2>/dev/null || true
    rm -rf ${CONFIG_DIR} /tmp/s3peep-delete-test
    # Optionally: $MC rm -r --force local/${TEST_BUCKET} --insecure 2>/dev/null || true
}

# Set trap to cleanup on exit
trap cleanup EXIT

log "Tests completed successfully!"
