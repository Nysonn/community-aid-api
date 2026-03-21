package handlers

import (
	"errors"
	"log"
	"net/http"

	"community-aid-api/internal/helpers"
	"community-aid-api/internal/middleware"
	"community-aid-api/internal/models"
	"community-aid-api/internal/services"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc *services.UserService
	cld     *cloudinary.Cloudinary
}

func NewUserHandler(userSvc *services.UserService, cld *cloudinary.Cloudinary) *UserHandler {
	return &UserHandler{userSvc: userSvc, cld: cld}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	clerkID := c.GetString(middleware.ContextKeyClerkID)

	user, err := h.userSvc.GetUserByClerkID(c.Request.Context(), clerkID)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Printf("ERROR GetMe clerkID=%s: %v", clerkID, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, user)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	var input models.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.userSvc.UpdateUser(c.Request.Context(), userID, input)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Printf("ERROR UpdateMe user=%s: %v", userID, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, updated)
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	if err := c.Request.ParseMultipartForm(8 << 20); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "failed to parse form data")
		return
	}

	file, fileHeader, err := c.Request.FormFile("avatar")
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer file.Close()

	result, err := h.cld.Upload.Upload(c.Request.Context(), file, uploader.UploadParams{
		Folder:   "community-aid/avatars",
		PublicID: userID,
	})
	if err != nil {
		log.Printf("ERROR UploadAvatar cloudinary user=%s file=%s: %v", userID, fileHeader.Filename, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "failed to upload avatar")
		return
	}

	updated, err := h.userSvc.UploadUserAvatar(c.Request.Context(), userID, result.SecureURL)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Printf("ERROR UploadAvatar save user=%s: %v", userID, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, updated)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	page, pageSize := helpers.ParsePagination(c)

	users, total, err := h.userSvc.GetAllUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		log.Printf("ERROR GetAllUsers: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.PaginatedResponse(c, http.StatusOK, users, total, page, pageSize)
}

func (h *UserHandler) ActivateUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userSvc.SetUserActiveStatus(c.Request.Context(), id, true)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Printf("ERROR ActivateUser %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, user)
}

func (h *UserHandler) DeactivateUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userSvc.SetUserActiveStatus(c.Request.Context(), id, false)
	if errors.Is(err, services.ErrNotFound) {
		helpers.ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		log.Printf("ERROR DeactivateUser %s: %v", id, err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusOK, user)
}
