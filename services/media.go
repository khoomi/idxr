package services

import (
	"khoomi-api-io/khoomi_api/helper"
	"khoomi-api-io/khoomi_api/models"
	"log"

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
	err := validate.Struct(file)
	if err != nil {
		return uploader.UploadResult{}, err
	}

	uploadRes, err := helper.ImageUploadHelper(file.File)
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

	uploadRes, errUrl := helper.ImageUploadHelper(url.Url)
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

	res, errUrl := helper.ImageDeletionHelper(uploader.DestroyParams{PublicID: id})
	if errUrl != nil {
		return "", err
	}
	return res, nil
}
