package client

import (
	"log"
	"sync"

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

var (
	s3ClientInstance *minio.Client
	s3ClientOnce     sync.Once
	s3ClientErr      error
)

func GetS3Client() (*minio.Client, error) {
	s3ClientOnce.Do(func() {
		endpoint := config.GetString("s3.endpoint")
		if endpoint == "" {
			logrus.Warn("S3 endpoint is not set, using default MinIO endpoint")
			endpoint = "10.10.10.38:9000"
		}
		accessKey := config.GetString("s3.access_key")
		if accessKey == "" {
			logrus.Warn("S3 access key is not set, using default MinIO access key")
			accessKey = "minioadmin"
		}
		secretKey := config.GetString("s3.secret_key")
		if secretKey == "" {
			logrus.Warn("S3 secret key is not set, using default MinIO secret key")
			secretKey = "minioadmin"
		}

		useSSL := config.GetBool("s3.use_ssl")

		s3ClientInstance, s3ClientErr = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: useSSL,
		})

		if s3ClientErr != nil {
			log.Printf("Error creating MinIO client: %v", s3ClientErr)
		}
	})

	return s3ClientInstance, s3ClientErr
}
