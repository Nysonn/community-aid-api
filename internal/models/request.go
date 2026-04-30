package models

import (
	"fmt"
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
	PaymentType             *string   `json:"payment_type"`
	BankAccountName         *string   `json:"bank_account_name"`
	BankAccountNumber       *string   `json:"bank_account_number"`
	BankName                *string   `json:"bank_name"`
	ReceivingMobileProvider *string   `json:"receiving_mobile_provider"`
	ReceivingMobileNumber   *string   `json:"receiving_mobile_number"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	PosterName   string `json:"poster_name,omitempty"`
	PosterPhone  string `json:"poster_phone,omitempty"`
	PosterEmail  string `json:"poster_email,omitempty"`
}

type CreateRequestInput struct {
	Title        string   `json:"title"         validate:"required"`
	Description  string   `json:"description"   validate:"required"`
	Type         string   `json:"type"          validate:"required"`
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
	// Payment receiving details
	TargetAmount            *float64 `json:"target_amount"             validate:"omitempty,gt=0"`
	PaymentType             *string  `json:"payment_type"              validate:"omitempty,oneof=bank mobile_money"`
	BankAccountName         *string  `json:"bank_account_name"`
	BankAccountNumber       *string  `json:"bank_account_number"`
	BankName                *string  `json:"bank_name"`
	ReceivingMobileProvider *string  `json:"receiving_mobile_provider" validate:"omitempty,oneof=mtn_momo airtel_money"`
	ReceivingMobileNumber   *string  `json:"receiving_mobile_number"`
}

func (input *UpdateRequestInput) Normalize() {
	trimStringPtr(&input.Title)
	trimStringPtr(&input.Description)
	trimStringPtr(&input.Status)
	trimStringPtr(&input.LocationName)
	trimStringPtr(&input.PaymentType)
	trimStringPtr(&input.BankAccountName)
	trimStringPtr(&input.BankAccountNumber)
	trimStringPtr(&input.BankName)
	trimStringPtr(&input.ReceivingMobileProvider)
	trimStringPtr(&input.ReceivingMobileNumber)
}

type RequestFilters struct {
	Type         *string `json:"type"`
	Status       *string `json:"status"`
	LocationName *string `json:"location_name"`
}

func (input *CreateRequestInput) Normalize() {
	input.Title = trimAndCollapseWhitespace(input.Title)
	input.Description = trimAndCollapseWhitespace(input.Description)
	input.Type = trimAndCollapseWhitespace(input.Type)
	input.LocationName = trimAndCollapseWhitespace(input.LocationName)
	trimStringPtr(&input.PaymentType)
	trimStringPtr(&input.BankAccountName)
	trimStringPtr(&input.BankAccountNumber)
	trimStringPtr(&input.BankName)
	trimStringPtr(&input.ReceivingMobileProvider)
	trimStringPtr(&input.ReceivingMobileNumber)
}

func (input CreateRequestInput) ValidateBusinessRules() error {
	return validateFundingConfiguration(
		input.TargetAmount,
		input.PaymentType,
		input.BankAccountName,
		input.BankAccountNumber,
		input.BankName,
		input.ReceivingMobileProvider,
		input.ReceivingMobileNumber,
	)
}

func (input UpdateRequestInput) ValidatePayoutBusinessRules(existing *EmergencyRequest) error {
	// Merge: input fields take precedence over existing
	targetAmount := existing.TargetAmount
	if input.TargetAmount != nil {
		targetAmount = input.TargetAmount
	}
	paymentType := existing.PaymentType
	if input.PaymentType != nil {
		paymentType = input.PaymentType
	}
	bankAccountName := existing.BankAccountName
	if input.BankAccountName != nil {
		bankAccountName = input.BankAccountName
	}
	bankAccountNumber := existing.BankAccountNumber
	if input.BankAccountNumber != nil {
		bankAccountNumber = input.BankAccountNumber
	}
	bankName := existing.BankName
	if input.BankName != nil {
		bankName = input.BankName
	}
	receivingMobileProvider := existing.ReceivingMobileProvider
	if input.ReceivingMobileProvider != nil {
		receivingMobileProvider = input.ReceivingMobileProvider
	}
	receivingMobileNumber := existing.ReceivingMobileNumber
	if input.ReceivingMobileNumber != nil {
		receivingMobileNumber = input.ReceivingMobileNumber
	}
	return validateFundingConfiguration(
		targetAmount,
		paymentType,
		bankAccountName,
		bankAccountNumber,
		bankName,
		receivingMobileProvider,
		receivingMobileNumber,
	)
}

func (r EmergencyRequest) ValidateBusinessRules() error {
	return validateFundingConfiguration(
		r.TargetAmount,
		r.PaymentType,
		r.BankAccountName,
		r.BankAccountNumber,
		r.BankName,
		r.ReceivingMobileProvider,
		r.ReceivingMobileNumber,
	)
}

func validateFundingConfiguration(
	targetAmount *float64,
	paymentType *string,
	bankAccountName *string,
	bankAccountNumber *string,
	bankName *string,
	receivingMobileProvider *string,
	receivingMobileNumber *string,
) error {
	hasFundingFields := targetAmount != nil || paymentType != nil ||
		!isBlankPtr(bankAccountName) || !isBlankPtr(bankAccountNumber) || !isBlankPtr(bankName) ||
		!isBlankPtr(receivingMobileProvider) || !isBlankPtr(receivingMobileNumber)

	if !hasFundingFields {
		return nil
	}

	if targetAmount == nil || *targetAmount <= 0 {
		return fmt.Errorf("target_amount is required when a request accepts donations")
	}

	if paymentType == nil {
		return fmt.Errorf("payment_type is required when a request accepts donations")
	}

	switch *paymentType {
	case "bank":
		if isBlankPtr(bankAccountName) || isBlankPtr(bankAccountNumber) || isBlankPtr(bankName) {
			return fmt.Errorf("bank_account_name, bank_account_number, and bank_name are required when payment_type is bank")
		}
	case "mobile_money":
		if isBlankPtr(receivingMobileProvider) || isBlankPtr(receivingMobileNumber) {
			return fmt.Errorf("receiving_mobile_provider and receiving_mobile_number are required when payment_type is mobile_money")
		}
	default:
		return fmt.Errorf("payment_type must be one of bank mobile_money")
	}

	return nil
}

func trimAndCollapseWhitespace(value string) string {
	return normalizeWhitespace(value)
}
