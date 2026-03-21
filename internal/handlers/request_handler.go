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

	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "failed to parse form data")
		return
	}

	input := models.CreateRequestInput{
		Title:        c.PostForm("title"),
		Description:  c.PostForm("description"),
		Type:         c.PostForm("type"),
		LocationName: c.PostForm("location_name"),
	}
	if v := c.PostForm("latitude"); v != "" {
		if lat, err := strconv.ParseFloat(v, 64); err == nil {
			input.Latitude = &lat
		}
	}
	if v := c.PostForm("longitude"); v != "" {
		if lon, err := strconv.ParseFloat(v, 64); err == nil {
			input.Longitude = &lon
		}
	}

	if err := helpers.ValidateStruct(&input); err != nil {
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

	var input models.UpdateRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
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
