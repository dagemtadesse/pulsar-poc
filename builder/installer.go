package builder

import (
	"archive/zip"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Installer struct {
	SrcDir string
}

func hasHiddenFolder(path string) bool {
	hiddenFolderRegex := regexp.MustCompile(`/\..*`)
	return hiddenFolderRegex.MatchString(path)
}

func (ins *Installer) copyProject(sourceFile *multipart.FileHeader) (string, error) {
	src, err := sourceFile.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	destination, err := os.Create(filepath.Join(ins.SrcDir, sourceFile.Filename))
	if err != nil {
		return "", err
	}

	defer destination.Close()

	if io.Copy(destination, src); err != nil {
		return "", err
	}

	return destination.Name(), nil
}

func (ins *Installer) extractZipFile(sourceFile string) error {
	zipReader, err := zip.OpenReader(sourceFile)
	if err != nil {
		return err
	}

	for _, file := range zipReader.Reader.File {

		zippedFile, err := file.Open()
		if err != nil {
			return err
		}
		defer zippedFile.Close()

		path, filePath := filepath.Split(file.Name)
		if hasHiddenFolder(path) || strings.HasPrefix(filePath, ".") {
			continue
		}

		extractedFilePath := filepath.Join(
			ins.SrcDir,
			file.Name,
		)

		if file.FileInfo().IsDir() {
			log.Println("Directory Created:", extractedFilePath)
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {
			log.Println("File extracted:", file.Name)

			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				log.Fatal(err)
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return nil
}

func (ins *Installer) createDockerConfig() error {
	configSrc, err := os.Open("./config/dockerfile")
	if err != nil {
		return err
	}

	dockerFile := path.Join(ins.SrcDir, "dockerfile")
	var file *os.File

	if _, err := os.Stat(dockerFile); os.IsNotExist(err) {
		if file, err = os.Create(dockerFile); err != nil {
			return err
		}
	} else {
		if file, err = os.OpenFile(dockerFile, os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
			return err
		}
	}

	defer file.Close()

	_, err = io.Copy(file, configSrc)
	return err
}

func Setup(sourceFile *multipart.FileHeader) (*Installer, error) {
	tmpDir, err := os.MkdirTemp("", "project")
	if err != nil {
		return nil, err
	}

	ins := &Installer{SrcDir: tmpDir}

	zipFile, err := ins.copyProject(sourceFile)
	if err != nil {
		return nil, err
	}

	err = ins.extractZipFile(zipFile)
	if err != nil {
		return nil, err
	}

	err = ins.createDockerConfig()
	if err != nil {
		return nil, err
	}

	return &Installer{SrcDir: tmpDir}, nil
}
