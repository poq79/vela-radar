package util

import (
	"context"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioCfg struct {
	AccessKey string
	SecretKey string
	Name      string
	Endpoint  string
	UseSSL    bool
}

func UploadToMinio(cfg *MinioCfg, bucketName string, objectName string, content io.Reader, fileSize int64) (string, error) {
	// 创建 MinIO 客户端对象
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// 创建上下文
	ctx := context.Background()

	// 使用 PutObjectWithContext 将图片文件上传到 MinIO
	n, err := minioClient.PutObject(ctx, bucketName, objectName, content, fileSize, minio.PutObjectOptions{})
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	log.Printf("Successfully uploaded %s to MinIO. Size: %d bytes\n", objectName, n.Size)
	return "https://" + cfg.Endpoint + "/" + bucketName + "/" + objectName, nil
}
