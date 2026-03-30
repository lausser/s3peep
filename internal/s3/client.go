package s3

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/lausser/s3peep/internal/config"
)

// S3API is the interface for S3 operations used by Client.
// This enables mocking in tests.
type S3API interface {
	HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

type Client struct {
	client   S3API
	bucket   string
	endpoint string
}

type Bucket struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creation_date"`
	Region       string    `json:"region"`
}

type FileObject struct {
	Key          string    `json:"key"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	IsFolder     bool      `json:"is_folder"`
	FileType     string    `json:"file_type"`
}

// ListObjectsResult represents the result of a paginated list operation
type ListObjectsResult struct {
	Objects               []FileObject `json:"objects"`
	IsTruncated           bool         `json:"is_truncated"`
	NextContinuationToken string       `json:"next_continuation_token"`
	Prefix                string       `json:"prefix"`
	CommonPrefixes        []string     `json:"common_prefixes"`
}

func NewClient(profile *config.Profile, bucket string) (*Client, error) {
	awsCreds := credentials.NewStaticCredentialsProvider(
		profile.AccessKeyID,
		profile.SecretAccessKey,
		"",
	)

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(awsCreds),
		awsconfig.WithRegion(profile.Region),
	)
	if err != nil {
		return nil, err
	}

	var opts []func(*s3.Options)
	if profile.EndpointURL != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: profile.EndpointURL}, nil
			})
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = &profile.EndpointURL
			o.UsePathStyle = true
		})
		_ = customResolver
	}

	client := s3.NewFromConfig(cfg, opts...)

	return &Client{
		client:   client,
		bucket:   bucket,
		endpoint: profile.EndpointURL,
	}, nil
}

func (c *Client) TestConnection(ctx context.Context) error {
	if c.bucket != "" {
		_, err := c.client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(c.bucket),
		})
		return err
	}
	_, err := c.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	return err
}

func (c *Client) ListBuckets(ctx context.Context) ([]Bucket, error) {
	result, err := c.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	var buckets []Bucket
	for _, b := range result.Buckets {
		if b.Name != nil {
			bucket := Bucket{Name: *b.Name}
			if b.CreationDate != nil {
				bucket.CreationDate = *b.CreationDate
			}
			buckets = append(buckets, bucket)
		}
	}
	return buckets, nil
}

func (c *Client) SetBucket(bucket string) {
	c.bucket = bucket
}

// GetFileType determines the file type from extension
func GetFileType(filename string, isFolder bool) string {
	if isFolder {
		return "folder"
	}
	
	lower := strings.ToLower(filename)
	
	if strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") || strings.HasSuffix(lower, ".png") ||
		strings.HasSuffix(lower, ".gif") || strings.HasSuffix(lower, ".webp") || strings.HasSuffix(lower, ".svg") ||
		strings.HasSuffix(lower, ".bmp") || strings.HasSuffix(lower, ".ico") {
		return "image"
	}
	if strings.HasSuffix(lower, ".pdf") || strings.HasSuffix(lower, ".doc") || strings.HasSuffix(lower, ".docx") ||
		strings.HasSuffix(lower, ".txt") || strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".csv") ||
		strings.HasSuffix(lower, ".xls") || strings.HasSuffix(lower, ".xlsx") {
		return "document"
	}
	if strings.HasSuffix(lower, ".zip") || strings.HasSuffix(lower, ".tar") || strings.HasSuffix(lower, ".gz") ||
		strings.HasSuffix(lower, ".bz2") || strings.HasSuffix(lower, ".7z") || strings.HasSuffix(lower, ".rar") {
		return "archive"
	}
	if strings.HasSuffix(lower, ".mp4") || strings.HasSuffix(lower, ".avi") || strings.HasSuffix(lower, ".mov") ||
		strings.HasSuffix(lower, ".mkv") {
		return "video"
	}
	if strings.HasSuffix(lower, ".mp3") || strings.HasSuffix(lower, ".wav") || strings.HasSuffix(lower, ".flac") ||
		strings.HasSuffix(lower, ".aac") {
		return "audio"
	}
	if strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".py") || strings.HasSuffix(lower, ".go") ||
		strings.HasSuffix(lower, ".java") || strings.HasSuffix(lower, ".html") || strings.HasSuffix(lower, ".css") ||
		strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".xml") {
		return "code"
	}
	
	return "other"
}

func (c *Client) ListObjects(ctx context.Context, prefix string) ([]FileObject, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(c.bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	}

	result, err := c.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, err
	}

	var objects []FileObject

	for _, commonPrefix := range result.CommonPrefixes {
		if commonPrefix.Prefix != nil {
			name := *commonPrefix.Prefix
			if prefix != "" && strings.HasPrefix(name, prefix) {
				name = name[len(prefix):]
			}
			// Remove trailing slash for display
			displayName := strings.TrimSuffix(name, "/")
			objects = append(objects, FileObject{
				Key:      *commonPrefix.Prefix,
				Name:     displayName,
				IsFolder: true,
				FileType: "folder",
			})
		}
	}

	for _, obj := range result.Contents {
		if obj.Key != nil && obj.Size != nil && obj.LastModified != nil {
			name := *obj.Key
			if prefix != "" && strings.HasPrefix(name, prefix) {
				name = name[len(prefix):]
			}
			// Skip empty objects that represent folders
			if *obj.Size == 0 && strings.HasSuffix(*obj.Key, "/") {
				continue
			}
			objects = append(objects, FileObject{
				Key:          *obj.Key,
				Name:         name,
				Size:         *obj.Size,
				LastModified: *obj.LastModified,
				IsFolder:     false,
				FileType:     GetFileType(name, false),
			})
		}
	}

	return objects, nil
}

// ListObjectsPaginated lists objects with pagination support using continuation tokens
func (c *Client) ListObjectsPaginated(ctx context.Context, bucket, prefix, continuationToken string, maxKeys int) (*ListObjectsResult, error) {
	if bucket == "" {
		bucket = c.bucket
	}
	
	input := &s3.ListObjectsV2Input{
		Bucket:            aws.String(bucket),
		Prefix:            aws.String(prefix),
		Delimiter:         aws.String("/"),
		MaxKeys:           aws.Int32(int32(maxKeys)),
		ContinuationToken: nil,
	}
	
	if continuationToken != "" {
		input.ContinuationToken = aws.String(continuationToken)
	}

	result, err := c.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, err
	}

	var objects []FileObject
	var commonPrefixes []string

	for _, commonPrefix := range result.CommonPrefixes {
		if commonPrefix.Prefix != nil {
			name := *commonPrefix.Prefix
			if prefix != "" && strings.HasPrefix(name, prefix) {
				name = name[len(prefix):]
			}
			displayName := strings.TrimSuffix(name, "/")
			objects = append(objects, FileObject{
				Key:      *commonPrefix.Prefix,
				Name:     displayName,
				IsFolder: true,
				FileType: "folder",
			})
			commonPrefixes = append(commonPrefixes, *commonPrefix.Prefix)
		}
	}

	for _, obj := range result.Contents {
		if obj.Key != nil && obj.Size != nil && obj.LastModified != nil {
			name := *obj.Key
			if prefix != "" && strings.HasPrefix(name, prefix) {
				name = name[len(prefix):]
			}
			// Skip empty objects that represent folders
			if *obj.Size == 0 && strings.HasSuffix(*obj.Key, "/") {
				continue
			}
			objects = append(objects, FileObject{
				Key:          *obj.Key,
				Name:         name,
				Size:         *obj.Size,
				LastModified: *obj.LastModified,
				IsFolder:     false,
				FileType:     GetFileType(name, false),
			})
		}
	}

	nextToken := ""
	if result.NextContinuationToken != nil {
		nextToken = *result.NextContinuationToken
	}

	isTruncated := false
	if result.IsTruncated != nil {
		isTruncated = *result.IsTruncated
	}

	return &ListObjectsResult{
		Objects:               objects,
		IsTruncated:           isTruncated,
		NextContinuationToken: nextToken,
		Prefix:                prefix,
		CommonPrefixes:        commonPrefixes,
	}, nil
}

func (c *Client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	if bucket == "" {
		bucket = c.bucket
	}
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (c *Client) PutObject(ctx context.Context, key string, body io.Reader, size int64) error {
	// For simplicity, using PutObject. For large files, we'd use multipart upload
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
	})
	return err
}

// HeadObject gets object metadata without downloading
func (c *Client) HeadObject(ctx context.Context, bucket, key string) (*s3.HeadObjectOutput, error) {
	if bucket == "" {
		bucket = c.bucket
	}
	
	result, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UploadObject uploads a file to S3
func (c *Client) UploadObject(ctx context.Context, bucket, key string, body io.Reader, size int64) error {
	if bucket == "" {
		bucket = c.bucket
	}
	
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
	})
	return err
}

// DeleteObject deletes a single object
func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	if bucket == "" {
		bucket = c.bucket
	}
	
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// DeleteObjects deletes multiple objects
func (c *Client) DeleteObjects(ctx context.Context, bucket string, keys []string) error {
	if bucket == "" {
		bucket = c.bucket
	}
	
	var objects []types.ObjectIdentifier
	for _, key := range keys {
		objects = append(objects, types.ObjectIdentifier{Key: aws.String(key)})
	}
	
	_, err := c.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: objects,
		},
	})
	return err
}

// ObjectExists checks if an object exists
func (c *Client) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	if bucket == "" {
		bucket = c.bucket
	}
	
	_, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "NoSuchKey") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CreateFolder creates an empty folder by creating an empty object with trailing slash
func (c *Client) CreateFolder(ctx context.Context, bucket, folderPath string) error {
	if bucket == "" {
		bucket = c.bucket
	}
	
	// Ensure path ends with /
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}
	
	// Create empty object
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(folderPath),
		Body:   strings.NewReader(""),
	})
	return err
}
