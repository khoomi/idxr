package helpers

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"sync"

	"khoomi-api-io/api/pkg/models"
	"khoomi-api-io/api/pkg/util"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
)

const (
	MAX_FILE_SIZE = 10 << 20
	IMAGE_COUNT   = 5
)

// HandleSequentialImages handles image array from formdata
func HandleSequentialImages(c *gin.Context) ([]string, []uploader.UploadResult, error) {
	var (
		uploadedImagesUrl    []string
		uploadedImagesResult []uploader.UploadResult
		wg                   sync.WaitGroup
		mu                   sync.Mutex
		errs                 []error
	)

	if err := c.Request.ParseMultipartForm(MAX_FILE_SIZE); err != nil {
		return nil, nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	files := c.Request.MultipartForm.File["images"]

	if len(files) > IMAGE_COUNT {
		files = files[:IMAGE_COUNT]
	}

	for i, fileHeader := range files {
		wg.Add(1)
		go func(index int, fh *multipart.FileHeader) {
			defer wg.Done()

			file, err := fh.Open()
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("error opening file %d: %w", index, err))
				mu.Unlock()
				return
			}
			defer file.Close()

			imageUpload, err := util.FileUpload(models.File{File: file})
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to upload image %d: %w", index, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			uploadedImagesUrl = append(uploadedImagesUrl, imageUpload.SecureURL)
			uploadedImagesResult = append(uploadedImagesResult, imageUpload)
			mu.Unlock()
		}(i, fileHeader)
	}

	wg.Wait()

	if len(errs) > 0 {
		errMsg := "Failed to upload some images:"
		for _, err := range errs {
			errMsg += "\n" + err.Error()
		}
		return uploadedImagesUrl, uploadedImagesResult, errors.New(errMsg)
	}

	return uploadedImagesUrl, uploadedImagesResult, nil
}

// ExtractFilenameAndExtension extracts filename and extension from URL
func ExtractFilenameAndExtension(urlString string) (filename, extension string, err error) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse URL: %w", err)
	}

	filenameWithExtension := filepath.Base(parsedURL.Path)

	name := filenameWithExtension[:len(filenameWithExtension)-len(filepath.Ext(filenameWithExtension))]
	ext := filepath.Ext(filenameWithExtension)

	return name, ext, nil
}
