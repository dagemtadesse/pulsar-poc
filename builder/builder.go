package builder

import (
	"fmt"
	"github.com/artdarek/go-unzip"
	"io"
	"mime/multipart"
	"os"
)

func Build(fileHeader *multipart.FileHeader) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	tempFile, err := os.Create(fileHeader.Filename)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())

	// Copy
	if _, err = io.Copy(tempFile, src); err != nil {
		return "", err
	}

	fmt.Println(tempFile.Name())

	uz := unzip.New(tempFile.Name(), "./data")

	err = uz.Extract()
	if err != nil {
		fmt.Println(err)
	}

	return "", nil
}
