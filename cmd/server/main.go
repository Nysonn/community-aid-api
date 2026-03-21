package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"community-aid-api/internal/config"
	"community-aid-api/internal/db"
	"community-aid-api/internal/middleware"
	"community-aid-api/internal/routes"
	"community-aid-api/internal/services"
	pkgClerk "community-aid-api/pkg/clerk"
	pkgCloudinary "community-aid-api/pkg/cloudinary"
	pkgResend "community-aid-api/pkg/resend"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	pkgClerk.InitClerk(cfg)

	database := db.InitDB(cfg)
	defer database.Close()

	cld := pkgCloudinary.InitCloudinary(cfg)
	mailer := pkgResend.InitResend(cfg)

	userSvc := services.NewUserService(database)
	requestSvc := services.NewRequestService(database)
	emailSvc := services.NewEmailService(mailer)
	offerSvc := services.NewOfferService(database)
	donationSvc := services.NewDonationService(database)

	router := gin.Default()
	router.HandleMethodNotAllowed = false
	router.MaxMultipartMemory = 10 << 20 // 10 MB
	router.Use(middleware.RateLimit())
	router.Use(middleware.CORS(cfg.AllowedOrigins))

	routes.SetupRoutes(router, database, cld, mailer, userSvc, requestSvc, emailSvc, offerSvc, donationSvc, cfg.AppEnv)

	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("CommunityAid API starting on port %s (env: %s)", cfg.Port, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutdown signal received, draining connections...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server shut down cleanly")
}
