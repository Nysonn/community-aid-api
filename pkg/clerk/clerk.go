package clerk

import (
	"log"

	"community-aid-api/internal/config"

	clerkSDK "github.com/clerk/clerk-sdk-go/v2"
)

func InitClerk(cfg *config.Config) {
	clerkSDK.SetKey(cfg.ClerkSecretKey)
	log.Println("Clerk SDK initialised")
}
