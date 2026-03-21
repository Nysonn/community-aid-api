package models

import "time"

type Offer struct {
	ID               string    `json:"id"`
	RequestID        string    `json:"request_id"`
	ResponderName    string    `json:"responder_name"`
	ResponderContact string    `json:"responder_contact"`
	OfferType        string    `json:"offer_type"`
	Status           string    `json:"status"`
	Latitude         *float64  `json:"latitude"`
	Longitude        *float64  `json:"longitude"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateOfferInput struct {
	RequestID        string   `json:"request_id"        validate:"required"`
	ResponderName    string   `json:"responder_name"    validate:"required"`
	ResponderContact string   `json:"responder_contact" validate:"required"`
	OfferType        string   `json:"offer_type"        validate:"required,oneof=transport donation expertise"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
}

type UpdateOfferStatusInput struct {
	Status string `json:"status" validate:"required,oneof=pending accepted fulfilled"`
}
