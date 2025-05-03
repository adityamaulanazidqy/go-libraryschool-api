package helpers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SaveImage struct {
}

func SaveImages() *SaveImage {
	return &SaveImage{}
}

func (helper *SaveImage) Cover(file multipart.File, handler *multipart.FileHeader, separator string) (string, error) {
	path := "assets/images/cover/"

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		return "", err
	}
	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/png" {
		return "", fmt.Errorf("unsupported file type: only JPEG or PNG allowed")
	}
	file.Seek(0, io.SeekStart)

	if handler.Size > 2*1024*1024 {
		return "", fmt.Errorf("file too big: %d", handler.Size)
	}

	safeFilename := strings.ReplaceAll(filepath.Base(handler.Filename), " ", "_")
	filename := fmt.Sprintf("%d%s%s", time.Now().UnixNano(), separator, safeFilename)
	fullpath := filepath.Join(path, filename)

	dst, err := os.Create(fullpath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (helper *SaveImage) Profile(file multipart.File, handler *multipart.FileHeader, separator string) (string, error) {
	path := "assets/images/profile/"

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}

	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		return "", err
	}
	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/png" {
		return "", fmt.Errorf("unsupported file type: only JPEG or PNG allowed")
	}
	file.Seek(0, io.SeekStart)

	if handler.Size > 2*1024*1024 {
		return "", fmt.Errorf("file too big: %d", handler.Size)
	}

	safeFilename := filepath.Base(handler.Filename)
	filename := fmt.Sprintf("%d%s%s", time.Now().UnixNano(), separator, safeFilename)
	fullpath := filepath.Join(path, filename)

	dst, err := os.Create(fullpath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return filename, nil
}
