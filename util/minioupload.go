package util

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioCfg struct {
	AccessKey        string
	SecretKey        string
	Name             string
	Endpoint         string
	Endpoint_preview string
	BucketName       string
	UseSSL           bool
}

func UploadToMinio(cfg *MinioCfg, objectName string, content io.Reader, fileSize int64) (string, error) {
	// 创建 MinIO 客户端对象
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})

	if err != nil {
		return "", err
	}

	// 创建上下文
	ctx := context.Background()

	// 使用 PutObjectWithContext 将图片文件上传到 MinIO
	_, err = minioClient.PutObject(ctx, cfg.BucketName, objectName, content, fileSize, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}
	//log.Printf("Successfully uploaded %s to MinIO. Size: %d bytes\n", objectName, n.Size)
	return cfg.Endpoint_preview + "/" + cfg.BucketName + "/" + objectName, nil
}
