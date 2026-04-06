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
	// Fundraising
	TargetAmount   *float64 `json:"target_amount"`
	AmountReceived float64  `json:"amount_received"`
	// Payment receiving details (where donors send money)
	PaymentType              *string `json:"payment_type"`
	BankAccountName          *string `json:"bank_account_name"`
	BankAccountNumber        *string `json:"bank_account_number"`
	BankName                 *string `json:"bank_name"`
	ReceivingMobileProvider  *string `json:"receiving_mobile_provider"`
	ReceivingMobileNumber    *string `json:"receiving_mobile_number"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	PosterName string    `json:"poster_name,omitempty"`
}

type CreateRequestInput struct {
	Title        string   `json:"title"         validate:"required"`
	Description  string   `json:"description"   validate:"required"`
	Type         string   `json:"type"          validate:"required,oneof=medical food rescue shelter"`
	LocationName string   `json:"location_name" validate:"required"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	// Fundraising
	TargetAmount *float64 `json:"target_amount" validate:"omitempty,gt=0"`
	// Payment receiving details
	PaymentType             *string `json:"payment_type"              validate:"omitempty,oneof=bank mobile_money"`
	BankAccountName         *string `json:"bank_account_name"`
	BankAccountNumber       *string `json:"bank_account_number"`
	BankName                *string `json:"bank_name"`
	ReceivingMobileProvider *string `json:"receiving_mobile_provider" validate:"omitempty,oneof=mtn_momo airtel_money"`
	ReceivingMobileNumber   *string `json:"receiving_mobile_number"`
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
