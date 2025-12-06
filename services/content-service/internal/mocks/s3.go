package mocks

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type S3Mock struct{}

func NewS3() *S3Mock {
	return &S3Mock{}
}

func (s *S3Mock) CreateBucket(userID string) (string, error) {
	bucketName := fmt.Sprintf("user-%s-videos", userID)
	logrus.WithFields(logrus.Fields{"user_id": userID, "bucket_name": bucketName}).Info("Mock: Creating S3 bucket")
	return bucketName, nil
}

func (s *S3Mock) DeleteBucket(bucketName string) error {
	logrus.WithField("bucket_name", bucketName).Info("Mock: Deleting S3 bucket")
	return nil
}
