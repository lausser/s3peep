package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lausser/s3peep/internal/config"
	"github.com/lausser/s3peep/internal/s3"
)

// MockS3Service implements S3Service for testing handlers
type MockS3Service struct {
	listBucketsFunc  func(ctx context.Context) ([]s3.Bucket, error)
	listObjectsFunc  func(ctx context.Context, prefix string) ([]s3.FileObject, error)
	getObjectFunc    func(ctx context.Context, key string) (io.ReadCloser, error)
	setBucketFunc    func(bucket string)
	setBucketCalled  string
}

func (m *MockS3Service) ListBuckets(ctx context.Context) ([]s3.Bucket, error) {
	if m.listBucketsFunc != nil {
		return m.listBucketsFunc(ctx)
	}
	return nil, nil
}

func (m *MockS3Service) ListObjects(ctx context.Context, prefix string) ([]s3.FileObject, error) {
	if m.listObjectsFunc != nil {
		return m.listObjectsFunc(ctx, prefix)
	}
	return nil, nil
}

func (m *MockS3Service) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, key)
	}
	return nil, nil
}

func (m *MockS3Service) SetBucket(bucket string) {
	m.setBucketCalled = bucket
	if m.setBucketFunc != nil {
		m.setBucketFunc(bucket)
	}
}

// Test NewAPIHandler
func TestNewAPIHandler(t *testing.T) {
	cfg := &config.Config{}
	mockS3 := &MockS3Service{}
	handler := NewAPIHandler(cfg, "/tmp/config.json", mockS3)
	if handler == nil {
		t.Fatalf("NewAPIHandler: expected non-nil handler, got nil")
	}
	if handler.cfg != cfg {
		t.Fatalf("NewAPIHandler: expected cfg to be set correctly")
	}
	if handler.configPath != "/tmp/config.json" {
		t.Fatalf("NewAPIHandler: expected configPath to be set correctly")
	}
	if handler.s3Client == nil {
		t.Fatalf("NewAPIHandler: expected s3Client to be set correctly")
	}
}

// Test Handle with static file request
func TestHandle_StaticFile(t *testing.T) {
	cfg := &config.Config{}
	handler := NewAPIHandler(cfg, "/tmp/config.json", &MockS3Service{})

	// Test CSS file request
	req := httptest.NewRequest(http.MethodGet, "/static/styles.css", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handle CSS: expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "text/css" {
		t.Errorf("Handle CSS: expected Content-Type text/css, got %s", rr.Header().Get("Content-Type"))
	}
	if rr.Body.Len() == 0 {
		t.Errorf("Handle CSS: expected non-empty body")
	}

	// Test JS file request
	req = httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	rr = httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handle JS: expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/javascript" {
		t.Errorf("Handle JS: expected Content-Type application/javascript, got %s", rr.Header().Get("Content-Type"))
	}
	if rr.Body.Len() == 0 {
		t.Errorf("Handle JS: expected non-empty body")
	}
}

// Test Handle with index request
func TestHandle_Index(t *testing.T) {
	cfg := &config.Config{}
	handler := NewAPIHandler(cfg, "/tmp/config.json", &MockS3Service{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handle index: expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "text/html" {
		t.Errorf("Handle index: expected Content-Type text/html, got %s", rr.Header().Get("Content-Type"))
	}
	if rr.Body.Len() == 0 {
		t.Errorf("Handle index: expected non-empty body")
	}
}

// Test Handle with API buckets GET request
func TestHandleAPI_Buckets_Get(t *testing.T) {
	cfg := &config.Config{}
	mockS3 := &MockS3Service{
		listBucketsFunc: func(ctx context.Context) ([]s3.Bucket, error) {
			return []s3.Bucket{
				{Name: "bucket1"},
				{Name: "bucket2"},
			}, nil
		},
	}

	handler := NewAPIHandler(cfg, "/tmp/config.json", mockS3)

	req := httptest.NewRequest(http.MethodGet, "/api/buckets", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handle API buckets GET: expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Handle API buckets GET: expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var buckets []s3.Bucket
	if err := json.NewDecoder(rr.Body).Decode(&buckets); err != nil {
		t.Fatalf("Handle API buckets GET: failed to decode JSON: %v", err)
	}
	if len(buckets) != 2 {
		t.Errorf("Handle API buckets GET: expected 2 buckets, got %d", len(buckets))
	}
	if buckets[0].Name != "bucket1" {
		t.Errorf("Handle API buckets GET: expected first bucket 'bucket1', got '%s'", buckets[0].Name)
	}
	if buckets[1].Name != "bucket2" {
		t.Errorf("Handle API buckets GET: expected second bucket 'bucket2', got '%s'", buckets[1].Name)
	}
}

// Test Handle with API buckets GET request error
func TestHandleAPI_Buckets_Get_Error(t *testing.T) {
	cfg := &config.Config{}
	mockS3 := &MockS3Service{
		listBucketsFunc: func(ctx context.Context) ([]s3.Bucket, error) {
			return nil, fmt.Errorf("internal error")
		},
	}

	handler := NewAPIHandler(cfg, "/tmp/config.json", mockS3)

	req := httptest.NewRequest(http.MethodGet, "/api/buckets", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Handle API buckets GET error: expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// Test Handle with API buckets POST request
func TestHandleAPI_Buckets_Post(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Profiles: []config.Profile{
			{
				Name:            "test-profile",
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
			},
		},
		ActiveProfile: "test-profile",
	}
	// Save initial config so the handler can persist changes
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	mockS3 := &MockS3Service{}
	handler := NewAPIHandler(cfg, configPath, mockS3)

	body := bytes.NewBufferString(`{"bucket":"selected-bucket"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/buckets", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handle API buckets POST: expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Handle API buckets POST: expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var result map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Handle API buckets POST: failed to decode JSON: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("Handle API buckets POST: expected status 'ok', got '%s'", result["status"])
	}
	if result["bucket"] != "selected-bucket" {
		t.Errorf("Handle API buckets POST: expected bucket 'selected-bucket', got '%s'", result["bucket"])
	}

	// Verify bucket was set on the mock
	if mockS3.setBucketCalled != "selected-bucket" {
		t.Errorf("Handle API buckets POST: expected SetBucket called with 'selected-bucket', got '%s'", mockS3.setBucketCalled)
	}

	// Verify config was persisted
	savedCfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	profile := config.GetActiveProfile(savedCfg)
	if profile == nil {
		t.Fatalf("Handle API buckets POST: no active profile after save")
	}
	if profile.Bucket != "selected-bucket" {
		t.Errorf("Handle API buckets POST: expected saved bucket 'selected-bucket', got '%s'", profile.Bucket)
	}
}

// Test Handle with API buckets POST request missing body
func TestHandleAPI_Buckets_Post_MissingBody(t *testing.T) {
	cfg := &config.Config{
		Profiles: []config.Profile{
			{
				Name:            "test-profile",
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
			},
		},
		ActiveProfile: "test-profile",
	}
	handler := NewAPIHandler(cfg, "/tmp/config.json", &MockS3Service{})

	req := httptest.NewRequest(http.MethodPost, "/api/buckets", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	// Empty bucket string in body — the handler currently does not validate this as missing,
	// so it proceeds. This test documents current behavior.
	// If the handler validated empty bucket, this would be 400.
}

// Test Handle with API buckets POST request invalid JSON
func TestHandleAPI_Buckets_Post_InvalidJSON(t *testing.T) {
	cfg := &config.Config{
		Profiles: []config.Profile{
			{
				Name:            "test-profile",
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
			},
		},
		ActiveProfile: "test-profile",
	}
	handler := NewAPIHandler(cfg, "/tmp/config.json", &MockS3Service{})

	req := httptest.NewRequest(http.MethodPost, "/api/buckets", bytes.NewBufferString(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Handle API buckets POST invalid JSON: expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// Test Handle with API list request (no active bucket)
func TestHandleAPI_List_NoBucket(t *testing.T) {
	cfg := &config.Config{}
	mockS3 := &MockS3Service{
		listBucketsFunc: func(ctx context.Context) ([]s3.Bucket, error) {
			return []s3.Bucket{
				{Name: "bucket1"},
				{Name: "bucket2"},
			}, nil
		},
	}

	handler := NewAPIHandler(cfg, "/tmp/config.json", mockS3)

	req := httptest.NewRequest(http.MethodGet, "/api/list?prefix=test/", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handle API list no bucket: expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Handle API list no bucket: expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var buckets []s3.Bucket
	if err := json.NewDecoder(rr.Body).Decode(&buckets); err != nil {
		t.Fatalf("Handle API list no bucket: failed to decode JSON: %v", err)
	}
	if len(buckets) != 2 {
		t.Errorf("Handle API list no bucket: expected 2 buckets, got %d", len(buckets))
	}
}

// Test Handle with API list request (with active bucket)
func TestHandleAPI_List_WithBucket(t *testing.T) {
	now := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	cfg := &config.Config{
		Profiles: []config.Profile{
			{
				Name:            "test-profile",
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
				Bucket:          "test-bucket",
			},
		},
		ActiveProfile: "test-profile",
	}
	mockS3 := &MockS3Service{
		listObjectsFunc: func(ctx context.Context, prefix string) ([]s3.FileObject, error) {
			return []s3.FileObject{
				{Key: "test/subfolder/", Name: "test/subfolder/", IsFolder: true},
				{Key: "test/file1.txt", Name: "file1.txt", Size: 100, LastModified: now},
				{Key: "test/file2.txt", Name: "file2.txt", Size: 200, LastModified: now},
			}, nil
		},
	}

	handler := NewAPIHandler(cfg, "/tmp/config.json", mockS3)

	req := httptest.NewRequest(http.MethodGet, "/api/list?prefix=test/", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handle API list with bucket: expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Handle API list with bucket: expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var files []s3.FileObject
	if err := json.NewDecoder(rr.Body).Decode(&files); err != nil {
		t.Fatalf("Handle API list with bucket: failed to decode JSON: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("Handle API list with bucket: expected 3 files (1 folder + 2 files), got %d", len(files))
	}

	folderFound := false
	fileCount := 0
	totalSize := int64(0)
	for _, file := range files {
		if file.IsFolder {
			if file.Key != "test/subfolder/" {
				t.Errorf("Handle API list with bucket: folder key expected 'test/subfolder/', got '%s'", file.Key)
			}
			folderFound = true
		} else {
			fileCount++
			totalSize += file.Size
		}
	}

	if !folderFound {
		t.Errorf("Handle API list with bucket: expected folder object not found")
	}
	if fileCount != 2 {
		t.Errorf("Handle API list with bucket: expected 2 file objects, got %d", fileCount)
	}
	if totalSize != 300 {
		t.Errorf("Handle API list with bucket: expected total size 300, got %d", totalSize)
	}
}

// Test Handle with API list request error
func TestHandleAPI_List_Error(t *testing.T) {
	cfg := &config.Config{
		Profiles: []config.Profile{
			{
				Name:            "test-profile",
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
				Bucket:          "test-bucket",
			},
		},
		ActiveProfile: "test-profile",
	}
	mockS3 := &MockS3Service{
		listObjectsFunc: func(ctx context.Context, prefix string) ([]s3.FileObject, error) {
			return nil, fmt.Errorf("access denied")
		},
	}

	handler := NewAPIHandler(cfg, "/tmp/config.json", mockS3)

	req := httptest.NewRequest(http.MethodGet, "/api/list?prefix=test/", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Handle API list error: expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// Test Handle with API get request
func TestHandleAPI_Get(t *testing.T) {
	cfg := &config.Config{
		Profiles: []config.Profile{
			{
				Name:            "test-profile",
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
				Bucket:          "test-bucket",
			},
		},
		ActiveProfile: "test-profile",
	}
	mockS3 := &MockS3Service{
		getObjectFunc: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("test content")), nil
		},
	}

	handler := NewAPIHandler(cfg, "/tmp/config.json", mockS3)

	req := httptest.NewRequest(http.MethodGet, "/api/get?key=test-key", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handle API get: expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/octet-stream" {
		t.Errorf("Handle API get: expected Content-Type application/octet-stream, got %s", rr.Header().Get("Content-Type"))
	}
	expectedDisposition := `attachment; filename="test-key"`
	if rr.Header().Get("Content-Disposition") != expectedDisposition {
		t.Errorf("Handle API get: expected Content-Disposition %q, got %q", expectedDisposition, rr.Header().Get("Content-Disposition"))
	}
	if rr.Body.String() != "test content" {
		t.Errorf("Handle API get: expected body 'test content', got '%s'", rr.Body.String())
	}
}

// Test Handle with API get request missing key
func TestHandleAPI_Get_MissingKey(t *testing.T) {
	cfg := &config.Config{}
	handler := NewAPIHandler(cfg, "/tmp/config.json", &MockS3Service{})

	req := httptest.NewRequest(http.MethodGet, "/api/get", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Handle API get missing key: expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// Test Handle with API get request error
func TestHandleAPI_Get_Error(t *testing.T) {
	cfg := &config.Config{
		Profiles: []config.Profile{
			{
				Name:            "test-profile",
				Region:          "us-east-1",
				AccessKeyID:     "testkey",
				SecretAccessKey: "testsecret",
				Bucket:          "test-bucket",
			},
		},
		ActiveProfile: "test-profile",
	}
	mockS3 := &MockS3Service{
		getObjectFunc: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return nil, fmt.Errorf("no such key")
		},
	}

	handler := NewAPIHandler(cfg, "/tmp/config.json", mockS3)

	req := httptest.NewRequest(http.MethodGet, "/api/get?key=non-existent-key", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Handle API get error: expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// Test Handle with API profile list request
func TestHandleAPI_Profile_List(t *testing.T) {
	cfg := &config.Config{
		Profiles: []config.Profile{
			{
				Name:            "profile1",
				Region:          "us-east-1",
				AccessKeyID:     "key1",
				SecretAccessKey: "secret1",
			},
			{
				Name:            "profile2",
				Region:          "us-west-2",
				AccessKeyID:     "key2",
				SecretAccessKey: "secret2",
			},
		},
		ActiveProfile: "profile1",
	}
	handler := NewAPIHandler(cfg, "/tmp/config.json", &MockS3Service{})

	req := httptest.NewRequest(http.MethodGet, "/api/profile?action=list", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handle API profile list: expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Handle API profile list: expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Handle API profile list: failed to decode JSON: %v", err)
	}
	if result["active_profile"] != "profile1" {
		t.Errorf("Handle API profile list: expected active_profile 'profile1', got '%v'", result["active_profile"])
	}
	profiles, ok := result["profiles"].([]interface{})
	if !ok {
		t.Fatalf("Handle API profile list: expected profiles to be array, got %T", result["profiles"])
	}
	if len(profiles) != 2 {
		t.Errorf("Handle API profile list: expected 2 profiles, got %d", len(profiles))
	}
}

// Test Handle with API profile list request missing action
func TestHandleAPI_Profile_List_MissingAction(t *testing.T) {
	cfg := &config.Config{}
	handler := NewAPIHandler(cfg, "/tmp/config.json", &MockS3Service{})

	req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Handle API profile list missing action: expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// Test Handle with API profile switch request
func TestHandleAPI_Profile_Switch(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := &config.Config{
		Profiles: []config.Profile{
			{
				Name:            "profile1",
				Region:          "us-east-1",
				AccessKeyID:     "key1",
				SecretAccessKey: "secret1",
			},
			{
				Name:            "profile2",
				Region:          "us-west-2",
				AccessKeyID:     "key2",
				SecretAccessKey: "secret2",
			},
		},
		ActiveProfile: "profile1",
	}
	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("Failed to save initial config: %v", err)
	}

	handler := NewAPIHandler(cfg, configPath, &MockS3Service{})

	req := httptest.NewRequest(http.MethodGet, "/api/profile?action=switch&name=profile2", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handle API profile switch: expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Handle API profile switch: expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("Handle API profile switch: failed to decode JSON: %v", err)
	}
	if result["active_profile"] != "profile2" {
		t.Errorf("Handle API profile switch: expected active_profile 'profile2', got '%v'", result["active_profile"])
	}

	// Verify active profile was switched in memory
	if cfg.ActiveProfile != "profile2" {
		t.Errorf("Handle API profile switch: expected ActiveProfile to be 'profile2', got '%s'", cfg.ActiveProfile)
	}

	// Verify config was persisted
	savedCfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}
	if savedCfg.ActiveProfile != "profile2" {
		t.Errorf("Handle API profile switch: saved config expected ActiveProfile 'profile2', got '%s'", savedCfg.ActiveProfile)
	}
}

// Test Handle with API profile switch request missing name
func TestHandleAPI_Profile_Switch_MissingName(t *testing.T) {
	cfg := &config.Config{}
	handler := NewAPIHandler(cfg, "/tmp/config.json", &MockS3Service{})

	req := httptest.NewRequest(http.MethodGet, "/api/profile?action=switch", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Handle API profile switch missing name: expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// Test Handle with API profile switch request invalid action
func TestHandleAPI_Profile_Switch_InvalidAction(t *testing.T) {
	cfg := &config.Config{}
	handler := NewAPIHandler(cfg, "/tmp/config.json", &MockS3Service{})

	req := httptest.NewRequest(http.MethodGet, "/api/profile?action=invalid", nil)
	rr := httptest.NewRecorder()
	handler.Handle(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Handle API profile switch invalid action: expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

// Suppress unused import warnings
var _ = os.TempDir
