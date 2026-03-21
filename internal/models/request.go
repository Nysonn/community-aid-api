package models

import (
	"time"

	"github.com/lib/pq"
)

type EmergencyRequest struct {
	ID           string         `json:"id"`
	UserID       string         `json:"user_id"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Type         string         `json:"type"`
	Status       string         `json:"status"`
	LocationName string         `json:"location_name"`
	Latitude     *float64       `json:"latitude"`
	Longitude    *float64       `json:"longitude"`
	MediaURLs    pq.StringArray `json:"media_urls"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type CreateRequestInput struct {
	Title        string   `json:"title"         validate:"required"`
	Description  string   `json:"description"   validate:"required"`
	Type         string   `json:"type"          validate:"required,oneof=medical food rescue shelter"`
	LocationName string   `json:"location_name" validate:"required"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

type UpdateRequestInput struct {
	Title        *string  `json:"title"`
	Description  *string  `json:"description"`
	Status       *string  `json:"status"`
	LocationName *string  `json:"location_name"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}

type RequestFilters struct {
	Type         *string `json:"type"`
	Status       *string `json:"status"`
	LocationName *string `json:"location_name"`
}
