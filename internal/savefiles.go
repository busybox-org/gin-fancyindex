package internal

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
)

// saveFiles 保存为本地文件
func saveFiles(c *gin.Context, filePath string) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files := form.File["files"]
	if len(files) == 0 {
		return fmt.Errorf("files is null")
	}
	var wg = new(sync.WaitGroup)
	var errs error
	// create _tmpSaveDir
	if !FileOrPathExist(c.Request.URL.Path) {
		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	for _, f := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader, filePath string) {
			defer wg.Done()
			//file.Filename does not contain the directory path
			// RFC 7578, Section 4.2 requires that if a filename is provided, the
			// directory path information must not be used.
			// 🙂🙂🙂🙂🙂🙂
			v := file.Header.Get("Content-Disposition")
			_, dispositionParams, err := mime.ParseMediaType(v)
			if err != nil {
				logrus.Error(err)
				return
			}
			fileName, ok := dispositionParams["filename"]
			if !ok {
				logrus.Error("filename does not exist")
				return
			}

			file.Filename = fileName
			// Default save path
			uploadFileName := filepath.Base(file.Filename)
			uploadFPath := filepath.Dir(file.Filename)
			// Process folder upload
			if uploadFPath != "." {
				filePath = filepath.Join(filePath, uploadFPath)
				if !FileOrPathExist(filePath) {
					_ = os.MkdirAll(filePath, os.ModePerm)
				}
			}

			// save file to local in _tp
			err = c.SaveUploadedFile(file, filepath.Join(filePath, uploadFileName))
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf(file.Filename, err.Error()))
			}
		}(f, filePath)
	}
	wg.Wait()
	if errs != nil {
		return errs
	}
	return nil
}
