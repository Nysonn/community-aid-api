package cloudinary

import (
	"log"

	"community-aid-api/internal/config"

	"github.com/cloudinary/cloudinary-go/v2"
)

func InitCloudinary(cfg *config.Config) *cloudinary.Cloudinary {
	cld, err := cloudinary.NewFromParams(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
	)
	if err != nil {
		log.Fatalf("failed to initialise Cloudinary client: %v", err)
	}

	log.Println("Cloudinary client initialised")
	return cld
}
