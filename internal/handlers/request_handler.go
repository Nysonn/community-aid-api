package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"community-aid-api/internal/helpers"
	"community-aid-api/internal/middleware"
	"community-aid-api/internal/models"
	"community-aid-api/internal/services"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

type createRequestPayload struct {
	Title                      string   `json:"title"`
	Description                string   `json:"description"`
	Type                       string   `json:"type"`
	LocationName               *string  `json:"location_name"`
	LocationNameAlt            *string  `json:"locationName"`
	Latitude                   *float64 `json:"latitude"`
	Longitude                  *float64 `json:"longitude"`
	TargetAmount               *float64 `json:"target_amount"`
	TargetAmountAlt            *float64 `json:"targetAmount"`
	PaymentType                *string  `json:"payment_type"`
	PaymentTypeAlt             *string  `json:"paymentType"`
	BankAccountName            *string  `json:"bank_account_name"`
	BankAccountNameAlt         *string  `json:"bankAccountName"`
	BankAccountNumber          *string  `json:"bank_account_number"`
	BankAccountNumberAlt       *string  `json:"bankAccountNumber"`
	BankName                   *string  `json:"bank_name"`
	BankNameAlt                *string  `json:"bankName"`
	ReceivingMobileProvider    *string  `json:"receiving_mobile_provider"`
	ReceivingMobileProviderAlt *string  `json:"receivingMobileProvider"`
	ReceivingMobileNumber      *string  `json:"receiving_mobile_number"`
	ReceivingMobileNumberAlt   *string  `json:"receivingMobileNumber"`
}

func (p createRequestPayload) toInput() models.CreateRequestInput {
	return models.CreateRequestInput{
		Title:                   p.Title,
		Description:             p.Description,
		Type:                    p.Type,
		LocationName:            firstStringValue(p.LocationName, p.LocationNameAlt),
		Latitude:                p.Latitude,
		Longitude:               p.Longitude,
		TargetAmount:            firstFloatValue(p.TargetAmount, p.TargetAmountAlt),
		PaymentType:             firstStringPtr(p.PaymentType, p.PaymentTypeAlt),
		BankAccountName:         firstStringPtr(p.BankAccountName, p.BankAccountNameAlt),
		BankAccountNumber:       firstStringPtr(p.BankAccountNumber, p.BankAccountNumberAlt),
		BankName:                firstStringPtr(p.BankName, p.BankNameAlt),
		ReceivingMobileProvider: firstStringPtr(p.ReceivingMobileProvider, p.ReceivingMobileProviderAlt),
		ReceivingMobileNumber:   firstStringPtr(p.ReceivingMobileNumber, p.ReceivingMobileNumberAlt),
	}
}

type updateRequestPayload struct {
	Title                      *string  `json:"title"`
	Description                *string  `json:"description"`
	Status                     *string  `json:"status"`
	LocationName               *string  `json:"location_name"`
	LocationNameAlt            *string  `json:"locationName"`
	Latitude                   *float64 `json:"latitude"`
	Longitude                  *float64 `json:"longitude"`
	TargetAmount               *float64 `json:"target_amount"`
	TargetAmountAlt            *float64 `json:"targetAmount"`
	PaymentType                *string  `json:"payment_type"`
	PaymentTypeAlt             *string  `json:"paymentType"`
	BankAccountName            *string  `json:"bank_account_name"`
	BankAccountNameAlt         *string  `json:"bankAccountName"`
	BankAccountNumber          *string  `json:"bank_account_number"`
	BankAccountNumberAlt       *string  `json:"bankAccountNumber"`
	BankName                   *string  `json:"bank_name"`
	BankNameAlt                *string  `json:"bankName"`
	ReceivingMobileProvider    *string  `json:"receiving_mobile_provider"`
	ReceivingMobileProviderAlt *string  `json:"receivingMobileProvider"`
	ReceivingMobileNumber      *string  `json:"receiving_mobile_number"`
	ReceivingMobileNumberAlt   *string  `json:"receivingMobileNumber"`
}

func (p updateRequestPayload) toInput() models.UpdateRequestInput {
	return models.UpdateRequestInput{
		Title:                   p.Title,
		Description:             p.Description,
		Status:                  p.Status,
		LocationName:            firstStringPtr(p.LocationName, p.LocationNameAlt),
		Latitude:                p.Latitude,
		Longitude:               p.Longitude,
		TargetAmount:            firstFloatValue(p.TargetAmount, p.TargetAmountAlt),
		PaymentType:             firstStringPtr(p.PaymentType, p.PaymentTypeAlt),
		BankAccountName:         firstStringPtr(p.BankAccountName, p.BankAccountNameAlt),
		BankAccountNumber:       firstStringPtr(p.BankAccountNumber, p.BankAccountNumberAlt),
		BankName:                firstStringPtr(p.BankName, p.BankNameAlt),
		ReceivingMobileProvider: firstStringPtr(p.ReceivingMobileProvider, p.ReceivingMobileProviderAlt),
		ReceivingMobileNumber:   firstStringPtr(p.ReceivingMobileNumber, p.ReceivingMobileNumberAlt),
	}
}

func firstStringValue(values ...*string) string {
	if value := firstStringPtr(values...); value != nil {
		return *value
	}
	return ""
}

func firstStringPtr(values ...*string) *string {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func firstFloatValue(values ...*float64) *float64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func formValue(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := c.PostForm(key); value != "" {
			return value
		}
	}
	return ""
}

func parseOptionalFormFloat(c *gin.Context, keys ...string) *float64 {
	value := formValue(c, keys...)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

type RequestHandler struct {
	service  *services.RequestService
	emailSvc *services.EmailService
	cld      *cloudinary.Cloudinary
	db       *sql.DB
}

func NewRequestHandler(
	service *services.RequestService,
	emailSvc *services.EmailService,
	cld *cloudinary.Cloudinary,
	db *sql.DB,
) *RequestHandler {
	return &RequestHandler{service: service, emailSvc: emailSvc, cld: cld, db: db}
}

func (h *RequestHandler) CreateRequest(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	var input models.CreateRequestInput
	if c.ContentType() == "application/json" {
		var payload createRequestPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			helpers.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
			return
		}
		input = payload.toInput()
	} else {
		if c.ContentType() == "multipart/form-data" {
			if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
				helpers.ErrorResponse(c, http.StatusBadRequest, "failed to parse form data")
				return
			}
		} else if err := c.Request.ParseForm(); err != nil {
			helpers.ErrorResponse(c, http.StatusBadRequest, "failed to parse form data")
			return
		}

		input = models.CreateRequestInput{
			Title:                   formValue(c, "title"),
			Description:             formValue(c, "description"),
			Type:                    formValue(c, "type"),
			LocationName:            formValue(c, "location_name", "locationName"),
			TargetAmount:            parseOptionalFormFloat(c, "target_amount", "targetAmount"),
			PaymentType:             stringPtrFromForm(c, "payment_type", "paymentType"),
			BankAccountName:         stringPtrFromForm(c, "bank_account_name", "bankAccountName"),
			BankAccountNumber:       stringPtrFromForm(c, "bank_account_number", "bankAccountNumber"),
			BankName:                stringPtrFromForm(c, "bank_name", "bankName"),
			ReceivingMobileProvider: stringPtrFromForm(c, "receiving_mobile_provider", "receivingMobileProvider"),
			ReceivingMobileNumber:   stringPtrFromForm(c, "receiving_mobile_number", "receivingMobileNumber"),
			Latitude:                parseOptionalFormFloat(c, "latitude"),
			Longitude:               parseOptionalFormFloat(c, "longitude"),
		}
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

	var mediaURLs []string
	if form := c.Request.MultipartForm; form != nil {
		files := form.File["media"]
		if len(files) > 5 {
			files = files[:5]
		}
		for _, fh := range files {
			file, err := fh.Open()
			if err != nil {
				log.Printf("warn: could not open uploaded file %s: %v", fh.Filename, err)
				continue
			}
			result, err := h.cld.Upload.Upload(c.Request.Context(), file, uploader.UploadParams{
				Folder: "community-aid/requests",
			})
			file.Close()
			if err != nil {
				log.Printf("warn: cloudinary upload failed for %s: %v", fh.Filename, err)
				continue
			}
			mediaURLs = append(mediaURLs, result.SecureURL)
		}
	}

	req, err := h.service.CreateRequest(c.Request.Context(), userID, input, mediaURLs)
	if err != nil {
		log.Printf("ERROR CreateRequest: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, req)
}

func (h *RequestHandler) GetAllRequests(c *gin.Context) {
	filters := models.RequestFilters{}
	if v := c.Query("type"); v != "" {
		filters.Type = &v
	}
	if v := c.Query("status"); v != "" {
		filters.Status = &v
	}
	if v := c.Query("location_name"); v != "" {
		filters.LocationName = &v
	}

	page, pageSize := helpers.ParsePagination(c)

	requests, total, err := h.service.GetAllRequests(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		log.Printf("ERROR GetAllRequests: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.PaginatedResponse(c, http.StatusOK, requests, total, page, pageSize)
}

func (h *RequestHandler) GetRequestByID(c *gin.Context) {
	id := c.Param("id")

	req, err := h.service.GetRequestByID(c.Request.Context(), id)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "request not found")
		return
	}
	if err != nil {
		log.Printf("ERROR GetRequestByID %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, req)
}

func (h *RequestHandler) GetMyRequests(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	requests, err := h.service.GetRequestsByUserID(c.Request.Context(), userID)
	if err != nil {
		log.Printf("ERROR GetMyRequests user=%s: %v", userID, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, requests)
}

func (h *RequestHandler) UpdateRequest(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString(middleware.ContextKeyUserID)
	userRole := c.GetString(middleware.ContextKeyUserRole)

	existing, err := h.service.GetRequestByID(c.Request.Context(), id)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "request not found")
		return
	}
	if err != nil {
		log.Printf("ERROR UpdateRequest fetch %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	if userRole != "admin" && existing.UserID != userID {
		helpers.ErrorResponse(c, http.StatusForbidden, "you do not have permission to update this request")
		return
	}

	var payload updateRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}
	input := payload.toInput()
	input.Normalize()
	if err := helpers.ValidateStruct(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := input.ValidatePayoutBusinessRules(existing); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.service.UpdateRequest(c.Request.Context(), id, input)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "request not found")
		return
	}
	if err != nil {
		log.Printf("ERROR UpdateRequest save %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, updated)
}

func stringPtrFromForm(c *gin.Context, keys ...string) *string {
	value := formValue(c, keys...)
	if value == "" {
		return nil
	}
	return &value
}

func (h *RequestHandler) DeleteRequest(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteRequest(c.Request.Context(), id); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			helpers.ErrorResponse(c, http.StatusNotFound, "request not found")
			return
		}
		log.Printf("ERROR DeleteRequest %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *RequestHandler) ApproveRequest(c *gin.Context) {
	id := c.Param("id")
	status := "approved"

	existing, err := h.service.GetRequestByID(c.Request.Context(), id)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "request not found")
		return
	}
	if err != nil {
		log.Printf("ERROR ApproveRequest fetch %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}
	if err := existing.ValidateBusinessRules(); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.service.UpdateRequest(c.Request.Context(), id, models.UpdateRequestInput{Status: &status})
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "request not found")
		return
	}
	if err != nil {
		log.Printf("ERROR ApproveRequest %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	go h.notifyOwner(updated.UserID, updated.Title, true)
	helpers.SuccessResponse(c, http.StatusOK, updated)
}

func (h *RequestHandler) RejectRequest(c *gin.Context) {
	id := c.Param("id")
	status := "rejected"

	updated, err := h.service.UpdateRequest(c.Request.Context(), id, models.UpdateRequestInput{Status: &status})
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "request not found")
		return
	}
	if err != nil {
		log.Printf("ERROR RejectRequest %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	go h.notifyOwner(updated.UserID, updated.Title, false)
	helpers.SuccessResponse(c, http.StatusOK, updated)
}

func (h *RequestHandler) notifyOwner(userID, requestTitle string, approved bool) {
	var email string
	if err := h.db.QueryRowContext(context.Background(), `SELECT email FROM users WHERE id = $1`, userID).Scan(&email); err != nil {
		log.Printf("warn: could not fetch owner email for user %s: %v", userID, err)
		return
	}
	if approved {
		if err := h.emailSvc.SendRequestApprovedEmail(email, requestTitle); err != nil {
			log.Printf("warn: failed to send approval email to %s: %v", email, err)
		}
	} else {
		if err := h.emailSvc.SendRequestRejectedEmail(email, requestTitle); err != nil {
			log.Printf("warn: failed to send rejection email to %s: %v", email, err)
		}
	}
}
