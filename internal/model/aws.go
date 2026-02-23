package model

type AWS struct {
	AccessKeyID     string `gorm:"size:255"`
	SecretAccessKey string `gorm:"size:255"`
	Bucket          string `gorm:"size:255"`
	Region          string `gorm:"size:255"`
}
