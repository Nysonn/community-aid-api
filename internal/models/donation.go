package models

import "time"

type Donation struct {
	ID        string    `json:"id"`
	RequestID string    `json:"request_id"`
	DonorName string    `json:"donor_name"`
	Amount    float64   `json:"amount"`
	Date      string    `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateDonationInput struct {
	RequestID string  `json:"request_id" validate:"required,uuid"`
	DonorName string  `json:"donor_name" validate:"required"`
	Amount    float64 `json:"amount"     validate:"required,gt=0"`
	Date      string  `json:"date"       validate:"required"`
}

type DonationWithRequest struct {
	Donation
	RequestTitle string `json:"request_title"`
}
