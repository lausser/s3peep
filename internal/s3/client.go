package s3

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

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
}

type Client struct {
	client   S3API
	bucket   string
	endpoint string
}

type Bucket struct {
	Name string `json:"name"`
}

type FileObject struct {
	Key          string    `json:"key"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	IsFolder     bool      `json:"is_folder"`
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
			buckets = append(buckets, Bucket{Name: *b.Name})
		}
	}
	return buckets, nil
}

func (c *Client) SetBucket(bucket string) {
	c.bucket = bucket
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
			objects = append(objects, FileObject{
				Key:      *commonPrefix.Prefix,
				Name:     *commonPrefix.Prefix,
				IsFolder: true,
			})
		}
	}

	for _, obj := range result.Contents {
		if obj.Key != nil && obj.Size != nil && obj.LastModified != nil {
			name := *obj.Key
			if prefix != "" {
				name = name[len(prefix):]
			}
			objects = append(objects, FileObject{
				Key:          *obj.Key,
				Name:         name,
				Size:         *obj.Size,
				LastModified: *obj.LastModified,
				IsFolder:     false,
			})
		}
	}

	return objects, nil
}

func (c *Client) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (c *Client) PutObject(ctx context.Context, key string, body io.Reader) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}
