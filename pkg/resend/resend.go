package resend

import (
	"log"

	"community-aid-api/internal/config"

	"github.com/resend/resend-go/v2"
)

func InitResend(cfg *config.Config) *resend.Client {
	client := resend.NewClient(cfg.ResendAPIKey)
	if client == nil {
		log.Fatal("failed to initialise Resend client")
	}

	log.Println("Resend client initialised")
	return client
}
