package models

import "time"

type Disbursement struct {
	ID                      string     `json:"id"`
	OfferID                 string     `json:"offer_id"`
	RequestID               string     `json:"request_id"`
	DonorName               string     `json:"donor_name"`
	DonorEmail              string     `json:"donor_email"`
	Amount                  float64    `json:"amount"`
	RecipientName           string     `json:"recipient_name"`
	PaymentType             string     `json:"payment_type"`
	BankAccountName         *string    `json:"bank_account_name"`
	BankAccountNumber       *string    `json:"bank_account_number"`
	BankName                *string    `json:"bank_name"`
	ReceivingMobileProvider *string    `json:"receiving_mobile_provider"`
	ReceivingMobileNumber   *string    `json:"receiving_mobile_number"`
	Status                  string     `json:"status"`
	Disbursedat             *time.Time `json:"disbursed_at"`
	CreatedAt               time.Time  `json:"created_at"`
	// Joined fields for display
	RequestTitle string `json:"request_title,omitempty"`
}
