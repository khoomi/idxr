package util

import (
	"context"
	"khoomi-api-io/api/pkg/models"
	"log"
	"time"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/go-playground/validator/v10"
)

type MediaUpload interface {
	FileUpload(file models.File) (string, error)
	RemoteUpload(url models.Url) (string, error)
}

var validate = validator.New()

func init_cloudinary() (*cloudinary.Cloudinary, error) {
	cloudName := LoadEnvFor("CLOUDINARY_CLOUDNAME")
	apiKey := LoadEnvFor("CLOUDINARY_API_KEY")
	apiSecret := LoadEnvFor("CLOUDINARY_API_SECRET")
	//create cloudinary instance
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return &cloudinary.Cloudinary{}, err
	}

	return cld, nil
}

func ImageUploadHelper(input interface{}) (uploader.UploadResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	//create cloudinary instance
	cld, err := init_cloudinary()
	if err != nil {
		return uploader.UploadResult{}, err
	}

	//upload file
	uploadFolder := LoadEnvFor("CLOUDINARY_UPLOAD_FOLDER")
	uploadRes, err := cld.Upload.Upload(ctx, input, uploader.UploadParams{Folder: uploadFolder})
	if err != nil {
		return uploader.UploadResult{}, err
	}

	return *uploadRes, nil
}

func ImageDeletionHelper(params uploader.DestroyParams) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	//create cloudinary instance
	cld, err := init_cloudinary()
	if err != nil {
		return "", err
	}


	deleteResult, err := cld.Upload.Destroy(ctx, params)
	if err != nil {
		return "", err
	}
	return deleteResult.Result, nil
}

func FileUpload(file models.File) (uploader.UploadResult, error) {
	err := validate.Struct(file)
	if err != nil {
		return uploader.UploadResult{}, err
	}

	uploadRes, err := ImageUploadHelper(file.File)
	if err != nil {
		return uploader.UploadResult{}, err
	}

	return uploadRes, nil
}

func RemoteUpload(url models.Url) (uploader.UploadResult, error) {
	err := validate.Struct(url)
	if err != nil {
		return uploader.UploadResult{}, err
	}

	uploadRes, errUrl := ImageUploadHelper(url.Url)
	if errUrl != nil {
		return uploader.UploadResult{}, err
	}

	return uploadRes, nil
}

func DestroyMedia(id string) (string, error) {
	err := validate.Struct(id)
	if err != nil {
		log.Println(err)
		return "", err
	}

	res, errUrl := ImageDeletionHelper(uploader.DestroyParams{PublicID: id})
	if errUrl != nil {
		return "", err
	}
	return res, nil
}
