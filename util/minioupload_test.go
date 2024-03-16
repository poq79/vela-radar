package util

import (
	"os"
	"testing"
)

func TestUploadToMinio(t *testing.T) {
	filePath := "xxx"
	cfg := MinioCfg{}
	file, err := os.Open(filePath)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()
	stat, _ := file.Stat()
	if err != nil {
		t.Error(err)
	}
	fileSize := stat.Size()
	UploadToMinio(&cfg, "123.jpg", file, fileSize)
}
