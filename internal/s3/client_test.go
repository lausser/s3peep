package s3

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/lausser/s3peep/internal/config"
	"github.com/stretchr/testify/mock"
)

// Mock S3 client for testing
type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) HeadBucket(ctx context.Context, input *s3.HeadBucketInput, opts ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).(*s3.HeadBucketOutput), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockS3Client) ListBuckets(ctx context.Context, input *s3.ListBucketsInput, opts ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).(*s3.ListBucketsOutput), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, input *s3.ListObjectsV2Input, opts ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).(*s3.ListObjectsV2Output), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockS3Client) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
	}
	return nil, args.Error(1)
}

// Test NewClient
func TestNewClient(t *testing.T) {
	profile := &config.Profile{
		Name:            "test",
		Region:          "us-east-1",
		AccessKeyID:     "testkey",
		SecretAccessKey: "testsecret",
		EndpointURL:     "http://localhost:9000",
		Bucket:          "test-bucket",
	}

	client, err := NewClient(profile, "test-bucket")
	if err != nil {
		t.Fatalf("NewClient: unexpected error: %v", err)
	}
	if client == nil {
		t.Fatalf("NewClient: expected non-nil client, got nil")
	}
	if client.bucket != "test-bucket" {
		t.Fatalf("NewClient: expected bucket 'test-bucket', got '%s'", client.bucket)
	}
	if client.endpoint != "http://localhost:9000" {
		t.Fatalf("NewClient: expected endpoint 'http://localhost:9000', got '%s'", client.endpoint)
	}
}

// Test NewClient with empty endpoint
func TestNewClient_EmptyEndpoint(t *testing.T) {
	profile := &config.Profile{
		Name:            "test",
		Region:          "us-east-1",
		AccessKeyID:     "testkey",
		SecretAccessKey: "testsecret",
		EndpointURL:     "",
		Bucket:          "test-bucket",
	}

	client, err := NewClient(profile, "test-bucket")
	if err != nil {
		t.Fatalf("NewClient empty endpoint: unexpected error: %v", err)
	}
	if client == nil {
		t.Fatalf("NewClient empty endpoint: expected non-nil client, got nil")
	}
	if client.endpoint != "" {
		t.Fatalf("NewClient empty endpoint: expected empty endpoint, got '%s'", client.endpoint)
	}
}

// Test TestConnection with bucket specified
func TestTestConnection_WithBucket(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("HeadBucket", mock.Anything, mock.MatchedBy(func(input *s3.HeadBucketInput) bool {
		return input.Bucket != nil && *input.Bucket == "test-bucket"
	})).Return(&s3.HeadBucketOutput{}, nil)

	client := &Client{
		client: mockS3,
		bucket: "test-bucket",
	}

	err := client.TestConnection(context.Background())
	if err != nil {
		t.Errorf("TestConnection with bucket: unexpected error: %v", err)
	}
	mockS3.AssertExpectations(t)
}

// Test TestConnection without bucket specified
func TestTestConnection_WithoutBucket(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("ListBuckets", mock.Anything, mock.AnythingOfType("*s3.ListBucketsInput")).Return(&s3.ListBucketsOutput{}, nil)

	client := &Client{
		client: mockS3,
		bucket: "",
	}

	err := client.TestConnection(context.Background())
	if err != nil {
		t.Errorf("TestConnection without bucket: unexpected error: %v", err)
	}
	mockS3.AssertExpectations(t)
}

// Test TestConnection with error
func TestTestConnection_Error(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("HeadBucket", mock.Anything, mock.AnythingOfType("*s3.HeadBucketInput")).Return(nil, fmt.Errorf("bucket not found"))

	client := &Client{
		client: mockS3,
		bucket: "test-bucket",
	}

	err := client.TestConnection(context.Background())
	if err == nil {
		t.Errorf("TestConnection error: expected error, got nil")
	}
	mockS3.AssertExpectations(t)
}

// Test ListBuckets
func TestListBuckets(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("ListBuckets", mock.Anything, mock.AnythingOfType("*s3.ListBucketsInput")).Return(&s3.ListBucketsOutput{
		Buckets: []types.Bucket{
			{
				Name: aws.String("bucket1"),
			},
			{
				Name: aws.String("bucket2"),
			},
		},
	}, nil)

	client := &Client{
		client: mockS3,
	}

	buckets, err := client.ListBuckets(context.Background())
	if err != nil {
		t.Errorf("ListBuckets: unexpected error: %v", err)
	}
	if len(buckets) != 2 {
		t.Errorf("ListBuckets: expected 2 buckets, got %d", len(buckets))
	}
	if buckets[0].Name != "bucket1" {
		t.Errorf("ListBuckets: expected first bucket 'bucket1', got '%s'", buckets[0].Name)
	}
	if buckets[1].Name != "bucket2" {
		t.Errorf("ListBuckets: expected second bucket 'bucket2', got '%s'", buckets[1].Name)
	}
	mockS3.AssertExpectations(t)
}

// Test ListBuckets with error
func TestListBuckets_Error(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("ListBuckets", mock.Anything, mock.AnythingOfType("*s3.ListBucketsInput")).Return(nil, fmt.Errorf("internal error"))

	client := &Client{
		client: mockS3,
	}

	_, err := client.ListBuckets(context.Background())
	if err == nil {
		t.Errorf("ListBuckets error: expected error, got nil")
	}
	mockS3.AssertExpectations(t)
}

// Test SetBucket
func TestSetBucket(t *testing.T) {
	client := &Client{
		bucket: "old-bucket",
	}
	client.SetBucket("new-bucket")
	if client.bucket != "new-bucket" {
		t.Errorf("SetBucket: expected bucket 'new-bucket', got '%s'", client.bucket)
	}
}

// Test ListObjects with folders and files
func TestListObjects(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("ListObjectsV2", mock.Anything, mock.MatchedBy(func(input *s3.ListObjectsV2Input) bool {
		return input.Bucket != nil && *input.Bucket == "test-bucket" &&
			input.Prefix != nil && *input.Prefix == "folder/" &&
			input.Delimiter != nil && *input.Delimiter == "/"
	})).Return(&s3.ListObjectsV2Output{
		CommonPrefixes: []types.CommonPrefix{
			{
				Prefix: aws.String("folder/subfolder/"),
			},
		},
		Contents: []types.Object{
			{
				Key:          aws.String("folder/file1.txt"),
				Size:         aws.Int64(100),
				LastModified: aws.Time(time.Now()),
			},
			{
				Key:          aws.String("folder/file2.txt"),
				Size:         aws.Int64(200),
				LastModified: aws.Time(time.Now()),
			},
		},
	}, nil)

	client := &Client{
		client: mockS3,
		bucket: "test-bucket",
	}

	objects, err := client.ListObjects(context.Background(), "folder/")
	if err != nil {
		t.Errorf("ListObjects: unexpected error: %v", err)
	}
	if len(objects) != 3 {
		t.Errorf("ListObjects: expected 3 objects (1 folder + 2 files), got %d", len(objects))
	}

	// Check folder object
	folderFound := false
	fileCount := 0
	totalSize := int64(0)
	for _, obj := range objects {
		if obj.IsFolder {
			if obj.Key != "folder/subfolder/" {
				t.Errorf("ListObjects: folder key expected 'folder/subfolder/', got '%s'", obj.Key)
			}
			folderFound = true
		} else {
			fileCount++
			totalSize += obj.Size
			if obj.Key == "folder/file1.txt" {
				if obj.Name != "file1.txt" {
					t.Errorf("ListObjects: file1 name expected 'file1.txt', got '%s'", obj.Name)
				}
				if obj.Size != 100 {
					t.Errorf("ListObjects: file1 size expected 100, got %d", obj.Size)
				}
			} else if obj.Key == "folder/file2.txt" {
				if obj.Name != "file2.txt" {
					t.Errorf("ListObjects: file2 name expected 'file2.txt', got '%s'", obj.Name)
				}
				if obj.Size != 200 {
					t.Errorf("ListObjects: file2 size expected 200, got %d", obj.Size)
				}
			} else {
				t.Errorf("ListObjects: unexpected file key '%s'", obj.Key)
			}
		}
	}

	if !folderFound {
		t.Errorf("ListObjects: expected folder object not found")
	}
	if fileCount != 2 {
		t.Errorf("ListObjects: expected 2 file objects, got %d", fileCount)
	}
	if totalSize != 300 {
		t.Errorf("ListObjects: expected total size 300, got %d", totalSize)
	}
	mockS3.AssertExpectations(t)
}

// Test ListObjects with empty prefix
func TestListObjects_EmptyPrefix(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("ListObjectsV2", mock.Anything, mock.MatchedBy(func(input *s3.ListObjectsV2Input) bool {
		return input.Bucket != nil && *input.Bucket == "test-bucket" &&
			input.Prefix != nil && *input.Prefix == "" &&
			input.Delimiter != nil && *input.Delimiter == "/"
	})).Return(&s3.ListObjectsV2Output{
		CommonPrefixes: []types.CommonPrefix{
			{
				Prefix: aws.String("folder1/"),
			},
			{
				Prefix: aws.String("folder2/"),
			},
		},
		Contents: []types.Object{
			{
				Key:          aws.String("file1.txt"),
				Size:         aws.Int64(150),
				LastModified: aws.Time(time.Now()),
			},
		},
	}, nil)

	client := &Client{
		client: mockS3,
		bucket: "test-bucket",
	}

	objects, err := client.ListObjects(context.Background(), "")
	if err != nil {
		t.Errorf("ListObjects empty prefix: unexpected error: %v", err)
	}
	if len(objects) != 3 {
		t.Errorf("ListObjects empty prefix: expected 3 objects (2 folders + 1 file), got %d", len(objects))
	}

	folderCount := 0
	fileCount := 0
	for _, obj := range objects {
		if obj.IsFolder {
			folderCount++
		} else {
			fileCount++
		}
	}
	if folderCount != 2 {
		t.Errorf("ListObjects empty prefix: expected 2 folders, got %d", folderCount)
	}
	if fileCount != 1 {
		t.Errorf("ListObjects empty prefix: expected 1 file, got %d", fileCount)
	}
	mockS3.AssertExpectations(t)
}

// Test ListObjects with error
func TestListObjects_Error(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("ListObjectsV2", mock.Anything, mock.AnythingOfType("*s3.ListObjectsV2Input")).Return(nil, fmt.Errorf("access denied"))

	client := &Client{
		client: mockS3,
		bucket: "test-bucket",
	}

	_, err := client.ListObjects(context.Background(), "prefix/")
	if err == nil {
		t.Errorf("ListObjects error: expected error, got nil")
	}
	mockS3.AssertExpectations(t)
}

// Test GetObject
func TestGetObject(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("GetObject", mock.Anything, mock.MatchedBy(func(input *s3.GetObjectInput) bool {
		return input.Bucket != nil && *input.Bucket == "test-bucket" &&
			input.Key != nil && *input.Key == "test-key"
	})).Return(&s3.GetObjectOutput{
		Body: io.NopCloser(strings.NewReader("test content")),
	}, nil)

	client := &Client{
		client: mockS3,
		bucket: "test-bucket",
	}

	body, err := client.GetObject(context.Background(), "test-key")
	if err != nil {
		t.Errorf("GetObject: unexpected error: %v", err)
	}
	if body == nil {
		t.Errorf("GetObject: expected non-nil body, got nil")
	}

	content, err := io.ReadAll(body)
	if err != nil {
		t.Errorf("GetObject: failed to read body: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("GetObject: expected content 'test content', got '%s'", string(content))
	}
	body.Close()
	mockS3.AssertExpectations(t)
}

// Test GetObject with error
func TestGetObject_Error(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("GetObject", mock.Anything, mock.AnythingOfType("*s3.GetObjectInput")).Return(nil, fmt.Errorf("no such key"))

	client := &Client{
		client: mockS3,
		bucket: "test-bucket",
	}

	_, err := client.GetObject(context.Background(), "non-existent-key")
	if err == nil {
		t.Errorf("GetObject error: expected error, got nil")
	}
	mockS3.AssertExpectations(t)
}

// Test PutObject
func TestPutObject(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("PutObject", mock.Anything, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
		return input.Bucket != nil && *input.Bucket == "test-bucket" &&
			input.Key != nil && *input.Key == "test-key" &&
			input.Body != nil
	})).Return(&s3.PutObjectOutput{}, nil)

	client := &Client{
		client: mockS3,
		bucket: "test-bucket",
	}

	body := strings.NewReader("test content")
	err := client.PutObject(context.Background(), "test-key", body)
	if err != nil {
		t.Errorf("PutObject: unexpected error: %v", err)
	}
	mockS3.AssertExpectations(t)
}

// Test PutObject with error
func TestPutObject_Error(t *testing.T) {
	mockS3 := &MockS3Client{}
	mockS3.On("PutObject", mock.Anything, mock.AnythingOfType("*s3.PutObjectInput")).Return(nil, fmt.Errorf("access denied"))

	client := &Client{
		client: mockS3,
		bucket: "test-bucket",
	}

	body := strings.NewReader("test content")
	err := client.PutObject(context.Background(), "test-key", body)
	if err == nil {
		t.Errorf("PutObject error: expected error, got nil")
	}
	mockS3.AssertExpectations(t)
}
