package helper

import (
	"context"
	"khoomi-api-io/khoomi_api/config"
	"time"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
)

func ImageUploadHelper(input interface{}) (uploader.UploadResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	cloudName := config.LoadEnvFor("CLOUDINARY_CLOUDNAME")
	apiKey := config.LoadEnvFor("CLOUDINARY_API_KEY")
	apiSecret := config.LoadEnvFor("CLOUDINARY_API_SECRET")
	//create cloudinary instance
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return uploader.UploadResult{}, err
	}

	//upload file
	uploadFolder := config.LoadEnvFor("CLOUDINARY_UPLOAD_FOLDER")
	uploadRes, err := cld.Upload.Upload(ctx, input, uploader.UploadParams{Folder: uploadFolder})
	if err != nil {
		return uploader.UploadResult{}, err
	}

	return *uploadRes, nil
}

func ImageDeletionHelper(params uploader.DestroyParams) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	cloudName := config.LoadEnvFor("CLOUDINARY_CLOUDNAME")
	apiKey := config.LoadEnvFor("CLOUDINARY_API_KEY")
	apiSecret := config.LoadEnvFor("CLOUDINARY_API_SECRET")
	//create cloudinary instance
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return "", err
	}

	//delete file
	deleteResult, err := cld.Upload.Destroy(ctx, params)
	if err != nil {
		return "", err
	}
	return deleteResult.Result, nil
}
