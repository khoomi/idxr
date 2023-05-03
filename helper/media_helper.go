package helper

import (
	"context"
	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"khoomi-api-io/khoomi_api/configs"
	"time"
)

func ImageUploadHelper(input interface{}) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cloudName := configs.LoadEnvFor("CLOUDINARY_CLOUDNAME")
	apiKey := configs.LoadEnvFor("CLOUDINARY_API_KEY")
	apiSecret := configs.LoadEnvFor("CLOUDINARY_API_SECRET")
	//create cloudinary instance
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return "", err
	}

	//upload file
	uploadFolder := configs.LoadEnvFor("CLOUDINARY_UPLOAD_FOLDER")
	uploadParam, err := cld.Upload.Upload(ctx, input, uploader.UploadParams{Folder: uploadFolder})
	if err != nil {
		return "", err
	}
	return uploadParam.SecureURL, nil
}
