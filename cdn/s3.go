package cdn

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
)

var s3Client *s3.S3
var bucket string

func CreateSession(key string, secret string, endpoint string, region string, b string) {
	bucket = b
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(true),
	}))
	s3Client = s3.New(sess)
}

func GetAllObjects(bucket string, path string) (*s3.ListObjectsOutput, error) {
	result, err := s3Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetExactObject(bucket string, key string) (*s3.GetObjectOutput, error) {
	result, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetObject(path string) (*s3.GetObjectOutput, error) {
	result, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func UploadToS3(file io.ReadSeeker, path string) error {
	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path),
		Body:   file,
	})
	return err
}
