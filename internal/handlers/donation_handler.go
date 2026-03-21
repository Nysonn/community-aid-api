package handlers

import (
	"errors"
	"log"
	"net/http"

	"community-aid-api/internal/helpers"
	"community-aid-api/internal/models"
	"community-aid-api/internal/services"

	"github.com/gin-gonic/gin"
)

type DonationHandler struct {
	donationSvc *services.DonationService
	requestSvc  *services.RequestService
}

func NewDonationHandler(donationSvc *services.DonationService, requestSvc *services.RequestService) *DonationHandler {
	return &DonationHandler{donationSvc: donationSvc, requestSvc: requestSvc}
}

func (h *DonationHandler) CreateDonation(c *gin.Context) {
	var input models.CreateDonationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := helpers.ValidateStruct(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err := h.requestSvc.GetRequestByID(c.Request.Context(), input.RequestID)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "request not found")
		return
	}
	if err != nil {
		log.Printf("ERROR CreateDonation verify request %s: %v", input.RequestID, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	donation, err := h.donationSvc.CreateDonation(c.Request.Context(), input)
	if err != nil {
		log.Printf("ERROR CreateDonation save: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, donation)
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
