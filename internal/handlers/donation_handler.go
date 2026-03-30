package handlers

import (
	"log"
	"net/http"

	"community-aid-api/internal/helpers"
	"community-aid-api/internal/services"

	"github.com/gin-gonic/gin"
)

type DonationHandler struct {
	donationSvc *services.DonationService
}

func NewDonationHandler(donationSvc *services.DonationService) *DonationHandler {
	return &DonationHandler{donationSvc: donationSvc}
}

func (h *DonationHandler) GetDonationsByRequestID(c *gin.Context) {
	requestID := c.Param("request_id")

	donations, err := h.donationSvc.GetDonationsByRequestID(c.Request.Context(), requestID)
	if err != nil {
		log.Printf("ERROR GetDonationsByRequestID %s: %v", requestID, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, donations)
}

func (h *DonationHandler) GetAllDonationsAdmin(c *gin.Context) {
	page, pageSize := helpers.ParsePagination(c)

	donations, total, err := h.donationSvc.GetAllDonationsAdmin(c.Request.Context(), page, pageSize)
	if err != nil {
		log.Printf("ERROR GetAllDonationsAdmin: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.PaginatedResponse(c, http.StatusOK, donations, total, page, pageSize)
}
