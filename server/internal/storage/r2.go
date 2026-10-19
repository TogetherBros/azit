package storage

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var ErrNotFound = errors.New("파일이 없어요")

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

// Copy 는 버킷 안에서 파일을 복사한다 (채팅 사진 → 내 짤 폴더).
func (r *R2) Copy(ctx context.Context, srcKey, dstKey string) error {
	_, err := r.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(r.bucket),
		CopySource: aws.String(r.bucket + "/" + srcKey),
		Key:        aws.String(dstKey),
	})
	return err
}

func (r *R2) Head(ctx context.Context, key string) (int64, string, error) {
	out, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		var nf *types.NotFound
		if errors.As(err, &nf) {
			return 0, "", ErrNotFound
		}
		return 0, "", err
	}

	return aws.ToInt64(out.ContentLength), aws.ToString(out.ContentType), nil
}
