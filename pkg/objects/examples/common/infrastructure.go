//go:build examples

package common

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	defaultMinIOWaitTimeout = 30 * time.Second
	defaultRetryDelay       = 1 * time.Second
)

// MinIOUser represents a MinIO user configuration
type MinIOUser struct {
	Username string
	Password string
	Bucket   string
}

// RunCommand executes a shell command with the given arguments
func RunCommand(ctx context.Context, out io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = getWriter(out)
	cmd.Stderr = getWriter(out)
	return cmd.Run()
}

// WaitForMinIO waits for MinIO to become available
func WaitForMinIO(ctx context.Context, container, accessKey, secretKey string) error {
	timeout := defaultMinIOWaitTimeout
	retryDelay := defaultRetryDelay

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.CommandContext(ctx, "docker", "exec", container,
			"mc", "alias", "set", "local", "http://localhost:9000", accessKey, secretKey)
		if err := cmd.Run(); err == nil {
			return nil
		}
		time.Sleep(retryDelay)
	}

	return ErrTimeoutWaitingForMinIO
}

// CreateMinIOUser creates a MinIO user with the specified credentials
func CreateMinIOUser(ctx context.Context, containerName, username, password string) error {
	cmd := exec.CommandContext(ctx, "docker", "exec", containerName,
		"mc", "admin", "user", "add", "local", username, password)
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "already exists") {
			return fmt.Errorf("create user %s: %w - %s", username, err, output)
		}
	}

	cmd = exec.CommandContext(ctx, "docker", "exec", containerName,
		"mc", "admin", "policy", "attach", "local", "readwrite", "--user", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "already mapped") {
			return fmt.Errorf("attach policy %s: %w - %s", username, err, output)
		}
	}

	return nil
}

// CreateMinIOBucket creates a MinIO bucket
func CreateMinIOBucket(ctx context.Context, containerName, bucket string) error {
	if bucket == "" {
		return ErrEmptyBucketName
	}

	cmd := exec.CommandContext(ctx, "docker", "exec", containerName,
		"mc", "mb", fmt.Sprintf("local/%s", bucket), "--ignore-existing")
	return cmd.Run()
}

// SetupMinIOUsers creates multiple MinIO users and their buckets
func SetupMinIOUsers(ctx context.Context, containerName string, users []MinIOUser) error {
	for _, user := range users {
		if err := CreateMinIOUser(ctx, containerName, user.Username, user.Password); err != nil {
			return err
		}

		if user.Bucket != "" {
			if err := CreateMinIOBucket(ctx, containerName, user.Bucket); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", user.Bucket, err)
			}
		}
	}

	return nil
}

// NewS3Client creates a new AWS S3 client
func NewS3Client(ctx context.Context, endpoint, accessKey, secretKey, region string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse endpoint: %w", err)
	}

	options := []func(*s3.Options){
		func(o *s3.Options) {
			o.UsePathStyle = true
			o.BaseEndpoint = aws.String(endpoint)
		},
	}

	if parsed.Scheme == "http" {
		options = append(options, func(o *s3.Options) {
			o.EndpointOptions.DisableHTTPS = true
		})
	}

	return s3.NewFromConfig(cfg, options...), nil
}

// EnsureBucket ensures an S3 bucket exists, creating it if necessary
func EnsureBucket(ctx context.Context, client *s3.Client, bucket string) error {
	if bucket == "" {
		return ErrEmptyBucketName
	}

	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		return nil
	}

	if !isNotFoundError(err) {
		return fmt.Errorf("head bucket: %w", err)
	}

	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		if isBucketExistsError(err) {
			return nil
		}

		return fmt.Errorf("create bucket: %w", err)
	}

	return nil
}

func getWriter(out io.Writer) io.Writer {
	if out == nil {
		return os.Stdout
	}
	return out
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "NotFound") || strings.Contains(msg, "404")
}

func isBucketExistsError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "BucketAlreadyExists") ||
		strings.Contains(msg, "BucketAlreadyOwnedByYou")
}
