package services

import (
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/go-playground/validator/v10"
)

var (
	validate = validator.New()
)

type MediaUpload interface {
	FileUpload(file models.File) (string, error)
	RemoteUpload(url models.Url) (string, error)
}

func FileUpload(file models.File) (uploader.UploadResult, error) {
	//validate
	err := validate.Struct(file)
	if err != nil {
		return uploader.UploadResult{}, err
	}

	//upload
	uploadRes, err := helper.ImageUploadHelper(file.File)
	if err != nil {
		return uploader.UploadResult{}, err
	}
	return uploadRes, nil
}

func RemoteUpload(url models.Url) (uploader.UploadResult, error) {
	//validate
	err := validate.Struct(url)
	if err != nil {
		return uploader.UploadResult{}, err
	}

	//upload
	uploadRes, errUrl := helper.ImageUploadHelper(url.Url)
	if errUrl != nil {
		return uploader.UploadResult{}, err
	}
	return uploadRes, nil
}

func DestroyMedia(id string) (string, error) {
	//validate
	err := validate.Struct(id)
	if err != nil {
		return "", err
	}

	//upload
	uploadUrl, errUrl := helper.ImageDeletionHelper(uploader.DestroyParams{PublicID: id})
	if errUrl != nil {
		return "", err
	}
	return uploadUrl, nil
}
