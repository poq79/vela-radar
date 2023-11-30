package util

import (
	"log"
	"os"
	"testing"
)

func TestUploadToMinio(t *testing.T) {
	filePath := "xxx"
	cfg := MinioCfg{}
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	stat, _ := file.Stat()
	if err != nil {
		log.Fatalln(err)
	}
	fileSize := stat.Size()
	UploadToMinio(&cfg, "radar-screenshoot", "123.jpg", file, fileSize)
}
