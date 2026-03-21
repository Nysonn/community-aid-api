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

type AuthHandler struct {
	userSvc *services.UserService
}

func NewAuthHandler(userSvc *services.UserService) *AuthHandler {
	return &AuthHandler{userSvc: userSvc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input models.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := helpers.ValidateStruct(&input); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	existing, err := h.userSvc.GetUserByClerkID(c.Request.Context(), input.ClerkID)
	if err != nil && !errors.Is(err, services.ErrNotFound) {
		log.Printf("ERROR Register GetUserByClerkID: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}
	if existing != nil {
		helpers.SuccessResponse(c, http.StatusOK, existing)
		return
	}

	user, err := h.userSvc.CreateUser(c.Request.Context(), input)
	if err != nil {
		log.Printf("ERROR Register CreateUser: %v", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	helpers.SuccessResponse(c, http.StatusCreated, user)
}
