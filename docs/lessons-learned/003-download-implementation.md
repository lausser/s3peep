# Lessons Learned: File Download Issues (003-download-implementation)

## Overview
Implementing file downloads in the Web UI turned out to be surprisingly complex, involving AWS SDK bugs, browser behavior quirks, and multiple failed approaches.

## The Problem

When clicking a file in the Web UI, we needed to:
1. Initiate a download from S3
2. Stream the file to the browser
3. Trigger the browser's download dialog
4. Save the file to the user's Downloads folder

## Failed Approaches

### Attempt 1: window.open()
```javascript
function downloadFile(key) {
    const downloadUrl = baseUrl + '/buckets/' + bucket + '/download?key=' + key;
    window.open(downloadUrl, '_blank');
}
```

**Result**: Failed with real AWS S3
- Worked with MinIO but not with real AWS S3
- Browser blocked popups or opened blank tabs
- Files didn't download consistently

### Attempt 2: Fetch API with Blob
```javascript
async function downloadFile(key) {
    const response = await fetch(downloadUrl);
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    // ... create anchor and click
}
```

**Result**: Failed with error
- "Failed to fetch" error appeared
- CORS issues with some S3 configurations
- Added unnecessary complexity

### Attempt 3: HeadObject + Content-Length Header
```go
headOutput, err := h.s3Client.HeadObject(ctx, bucket, key)
w.Header().Set("Content-Length", fmt.Sprintf("%d", headOutput.ContentLength))
```

**Result**: CRITICAL BUG - Wrong file size
- Content-Length header showed 824 GB instead of actual 26 bytes
- AWS SDK v2 issue #2524: https://github.com/aws/aws-sdk-go-v2/issues/2524
- Root cause: Module version incompatibility in aws-sdk-go-v2

**Impact**: Downloads appeared to start but never completed because browser expected 824 GB of data

## The Solution

### Backend: Remove Content-Length
Don't set Content-Length header at all:
```go
func (h *APIHandler) handleDownload(w http.ResponseWriter, r *http.Request, endpoint string) {
    // Set download headers (NO Content-Length)
    filename := path.Base(key)
    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
    
    // Stream the file
    body, err := h.s3Client.GetObject(ctx, bucket, key)
    if err != nil {
        h.writeError(w, http.StatusInternalServerError, "DOWNLOAD_ERROR", ...)
        return
    }
    defer body.Close()
    
    io.Copy(w, body)
}
```

### Frontend: Hidden Anchor Tag
```javascript
function downloadFile(key) {
    const downloadUrl = baseUrl + '/buckets/' + encodeURIComponent(state.currentBucket) + 
                        '/download?key=' + encodeURIComponent(key);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = key.split('/').pop();
    link.style.display = 'none';
    document.body.appendChild(link);
    link.click();
    setTimeout(() => {
        document.body.removeChild(link);
    }, 100);
}
```

**Why this works**:
1. Anchor tag with `download` attribute triggers native browser download
2. No popup blocking issues
3. No new tab needed
4. Works with both MinIO and real AWS S3
5. Browsers handle streaming without Content-Length header

## Key Lessons

### 1. Don't Trust SDK Metadata
AWS SDK v2's `HeadObject().ContentLength` can return garbage values due to SDK bugs. Always verify with actual testing against real services.

### 2. HTTP Headers Can Break Downloads
Content-Length header is optional. Browsers download files fine without it, streaming the response body as it arrives.

### 3. Browser Download Behavior
- `window.open()` is unreliable and blocked by popup blockers
- Fetch API adds unnecessary complexity for simple downloads
- Anchor tag with `download` attribute is the most reliable method

### 4. Test Against Real Services
MinIO and AWS S3 behave differently:
- MinIO: More lenient, worked with broken Content-Length
- AWS S3: Stricter, downloads failed with wrong Content-Length

### 5. AWS SDK Version Compatibility
aws-sdk-go-v2 has known issues with ContentLength in certain version combinations (issue #2524). The official fix (updating all modules) didn't work in our case.

## Verification Strategy

When testing file downloads, verify:
1. ✅ File content is correct (checksum comparison)
2. ✅ File size matches expected (not 824 GB!)
3. ✅ Works with MinIO
4. ✅ Works with real AWS S3
5. ✅ No browser console errors
6. ✅ Download completes without hanging

## Trade-offs

**Current Implementation**:
- ✅ Reliable downloads
- ✅ Works with all S3 providers
- ✅ No browser popup/dialog for save location (downloads directly to ~/Downloads)
- ⚠️ No download progress indicator (no Content-Length)
- ⚠️ No ability to choose save location

**Acceptable for now**: The trade-offs are acceptable for a simple file browser tool.

## References

- AWS SDK Issue #2524: https://github.com/aws/aws-sdk-go-v2/issues/2524
- AWS SDK Issue #2370: https://github.com/aws/aws-sdk-go-v2/issues/2370
- MinIO ContentLength Issue: https://github.com/minio/minio/issues/4838
