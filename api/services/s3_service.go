// api/services/s3_service.go

// TODO. 미완임

package services

import (
	"context"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	app_config "github.com/seojoonrp/bapddang-server/config"
)

type S3Service interface {
	UploadFile(fileHeader *multipart.FileHeader, fileName string) (string, error)
}

type s3Service struct {
	s3Client   *s3.Client
	bucketName string
}

func NewS3Service() (S3Service, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(app_config.AppConfig.AWSRegion))
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg)

	return &s3Service{
		s3Client:   s3Client,
		bucketName: app_config.AppConfig.AWSS3BucketName,
	}, nil
}

func (s *s3Service) UploadFile(fileHeader *multipart.FileHeader, fileName string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = s.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &s.bucketName,
		Key:    &fileName,
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	fileURL := "https://" + s.bucketName + ".s3." + app_config.AppConfig.AWSRegion + ".amazonaws.com/" + fileName
	return fileURL, nil
}
