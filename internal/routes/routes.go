package routes

import (
	"database/sql"
	"net/http"

	"community-aid-api/internal/handlers"
	"community-aid-api/internal/middleware"
	"community-aid-api/internal/services"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v2"
)

func SetupRoutes(
	router *gin.Engine,
	db *sql.DB,
	cld *cloudinary.Cloudinary,
	mailer *resend.Client,
	userSvc *services.UserService,
	requestSvc *services.RequestService,
	emailSvc *services.EmailService,
	offerSvc *services.OfferService,
	donationSvc *services.DonationService,
	disbursementSvc *services.DisbursementService,
	appEnv string,
) {
	authHandler := handlers.NewAuthHandler(userSvc)
	userHandler := handlers.NewUserHandler(userSvc, cld)
	requestHandler := handlers.NewRequestHandler(requestSvc, emailSvc, cld, db)
	offerHandler := handlers.NewOfferHandler(offerSvc, requestSvc, emailSvc, db)
	donationHandler := handlers.NewDonationHandler(donationSvc)
	adminHandler := handlers.NewAdminHandler(userSvc, requestSvc, offerSvc, donationSvc, disbursementSvc, emailSvc, db)

	v1 := router.Group("/api/v1")

	// Health check
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "env": appEnv})
	})

	// Auth
	v1.POST("/auth/register", authHandler.Register)

	// Public request routes
	v1.GET("/requests", requestHandler.GetAllRequests)
	v1.GET("/requests/:id", requestHandler.GetRequestByID)

	// Public offer routes
	v1.POST("/offers", offerHandler.CreateOffer)
	v1.GET("/offers/request/:request_id", offerHandler.GetOffersByRequestID)

	// Authenticated routes (ClerkAuth required)
	authRoutes := v1.Group("")
	authRoutes.Use(middleware.ClerkAuth(userSvc))
	{
		authRoutes.POST("/requests", requestHandler.CreateRequest)
		authRoutes.GET("/requests/me", requestHandler.GetMyRequests)
		authRoutes.PUT("/requests/:id", requestHandler.UpdateRequest)

		authRoutes.PUT("/offers/:id/status", offerHandler.UpdateOfferStatus)

		authRoutes.GET("/users/me", userHandler.GetMe)
		authRoutes.PUT("/users/me", userHandler.UpdateMe)
		authRoutes.POST("/users/me/avatar", userHandler.UploadAvatar)
	}

	// Admin-only routes (ClerkAuth + RequireAdmin)
	adminRoutes := v1.Group("")
	adminRoutes.Use(middleware.ClerkAuth(userSvc), middleware.RequireAdmin())
	{
		adminRoutes.DELETE("/requests/:id", requestHandler.DeleteRequest)
		adminRoutes.POST("/requests/:id/approve", requestHandler.ApproveRequest)
		adminRoutes.POST("/requests/:id/reject", requestHandler.RejectRequest)

		adminRoutes.GET("/admin/offers", offerHandler.GetAllOffersAdmin)

		adminRoutes.GET("/admin/donations", donationHandler.GetAllDonationsAdmin)
		adminRoutes.GET("/admin/donations/:request_id", donationHandler.GetDonationsByRequestID)

		adminRoutes.GET("/admin/users", userHandler.GetAllUsers)
		adminRoutes.PUT("/admin/users/:id/activate", userHandler.ActivateUser)
		adminRoutes.PUT("/admin/users/:id/deactivate", userHandler.DeactivateUser)
		adminRoutes.PUT("/admin/users/:id/promote", adminHandler.PromoteToAdmin)
		adminRoutes.PUT("/admin/users/:id/demote", adminHandler.DemoteFromAdmin)

		adminRoutes.GET("/admin/stats", adminHandler.GetDashboardStats)
		adminRoutes.GET("/admin/requests", adminHandler.GetAllRequestsAdmin)

		adminRoutes.GET("/admin/disbursements", adminHandler.GetDisbursements)
		adminRoutes.POST("/admin/disbursements/:id/disburse", adminHandler.MarkDisbursed)
	}
}
