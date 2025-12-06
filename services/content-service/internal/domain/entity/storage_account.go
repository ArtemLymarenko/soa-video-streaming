package entity

import "time"

const (
	StorageAccountStatusActive = "ACTIVE"
	StorageAccountStatusFailed = "FAILED"
)

type StorageAccount struct {
	UserID     string
	BucketName string
	Status     string
	CreatedAt  time.Time
}
