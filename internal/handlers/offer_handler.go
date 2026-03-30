package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"community-aid-api/internal/helpers"
	"community-aid-api/internal/middleware"
	"community-aid-api/internal/models"
	"community-aid-api/internal/services"

	"github.com/gin-gonic/gin"
)

type OfferHandler struct {
	offerSvc   *services.OfferService
	requestSvc *services.RequestService
	emailSvc   *services.EmailService
	db         *sql.DB
}

func NewOfferHandler(
	offerSvc *services.OfferService,
	requestSvc *services.RequestService,
	emailSvc *services.EmailService,
	db *sql.DB,
) *OfferHandler {
	return &OfferHandler{offerSvc: offerSvc, requestSvc: requestSvc, emailSvc: emailSvc, db: db}
}

func (h *OfferHandler) CreateOffer(c *gin.Context) {
	var input models.CreateOfferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}
	input.Normalize()
	if err := helpers.ValidateStruct(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := input.ValidateBusinessRules(); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	req, err := h.requestSvc.GetRequestByID(c.Request.Context(), input.RequestID)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "request not found")
		return
	}
	if err != nil {
		log.Printf("ERROR CreateOffer verify request %s: %v", input.RequestID, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	if req.Status != "approved" {
		helpers.ErrorResponse(c, http.StatusBadRequest, "offers can only be made on approved requests")
		return
	}

	offer, err := h.offerSvc.CreateOffer(c.Request.Context(), input)
	if err != nil {
		log.Printf("ERROR CreateOffer save: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	go h.notifyRequestOwner(req.UserID, req.Title)
	helpers.SuccessResponse(c, http.StatusCreated, offer)
}

func (h *OfferHandler) GetOffersByRequestID(c *gin.Context) {
	requestID := c.Param("request_id")

	offers, err := h.offerSvc.GetOffersByRequestID(c.Request.Context(), requestID)
	if err != nil {
		log.Printf("ERROR GetOffersByRequestID %s: %v", requestID, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, offers)
}

func (h *OfferHandler) UpdateOfferStatus(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)
	userRole := c.GetString(middleware.ContextKeyUserRole)

	offer, err := h.offerSvc.GetOfferByID(c.Request.Context(), id)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "offer not found")
		return
	}
	if err != nil {
		log.Printf("ERROR UpdateOfferStatus fetch %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	req, err := h.requestSvc.GetRequestByID(c.Request.Context(), offer.RequestID)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "associated request not found")
		return
	}
	if err != nil {
		log.Printf("ERROR UpdateOfferStatus fetch request %s: %v", offer.RequestID, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	if userRole != "admin" && req.UserID != userID {
		helpers.ErrorResponse(c, http.StatusForbidden, "you do not have permission to update this offer")
		return
	}

	var input models.UpdateOfferStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := helpers.ValidateStruct(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.offerSvc.UpdateOfferStatus(c.Request.Context(), id, input.Status)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "offer not found")
		return
	}
	if err != nil {
		log.Printf("ERROR UpdateOfferStatus save %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, updated)
}

func (h *OfferHandler) GetAllOffersAdmin(c *gin.Context) {
	page, pageSize := helpers.ParsePagination(c)

	offers, total, err := h.offerSvc.GetAllOffersAdmin(c.Request.Context(), page, pageSize)
	if err != nil {
		log.Printf("ERROR GetAllOffersAdmin: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.PaginatedResponse(c, http.StatusOK, offers, total, page, pageSize)
}

func (h *OfferHandler) notifyRequestOwner(userID, requestTitle string) {
	var email string
	if err := h.db.QueryRowContext(context.Background(), `SELECT email FROM users WHERE id = $1`, userID).Scan(&email); err != nil {
		log.Printf("warn: could not fetch owner email for user %s: %v", userID, err)
		return
	}
	if err := h.emailSvc.SendOfferNotificationEmail(email, requestTitle); err != nil {
		log.Printf("warn: failed to send offer notification to %s: %v", email, err)
	}
}
