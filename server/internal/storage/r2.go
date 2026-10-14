package storage

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2 struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

func NewR2(accountID, accessKeyID, secretAccessKey, bucket string) (*R2, error) {
	client := s3.New(s3.Options{
		Region:                     "auto",
		BaseEndpoint:               aws.String("https://" + accountID + ".r2.cloudflarestorage.com"),
		Credentials:                credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})

	return &R2{
		client:  client,
		presign: s3.NewPresignClient(client),
		bucket:  bucket,
	}, nil
}

const (
	putTTL = 10 * time.Minute
	getTTL = 24 * time.Hour
)

func (r *R2) PresignPut(ctx context.Context, key, contentType string, size int64) (string, error) {
	req, err := r.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}, s3.WithPresignExpires(putTTL))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (r *R2) PresignGet(ctx context.Context, key string) (string, error) {
	req, err := r.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(getTTL))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}
