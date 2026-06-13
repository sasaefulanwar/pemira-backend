package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(file interface{}) (string, error) {
	ctx := context.Background()

	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return "", fmt.Errorf("gagal inisialisasi cloudinary: %w", err)
	}

	// Proses Upload
	resp, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: "pemira_sengketa", // Folder di dashboard Cloudinary lu
	})
	if err != nil {
		return "", fmt.Errorf("gagal upload ke cloudinary: %w", err)
	}

	return resp.SecureURL, nil // URL inilah yang bakal disimpen di DB
}
