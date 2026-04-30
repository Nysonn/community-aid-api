package models

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

type Offer struct {
	ID                      string    `json:"id"`
	RequestID               string    `json:"request_id"`
	ResponderName           string    `json:"responder_name"`
	ResponderContact        string    `json:"responder_contact"`
	OfferType               string    `json:"offer_type"`
	Status                  string    `json:"status"`
	ExpertiseDetails        *string   `json:"expertise_details,omitempty"`
	VehicleType             *string   `json:"vehicle_type,omitempty"`
	DonationAmount          *float64  `json:"donation_amount,omitempty"`
	DonorEmail              *string   `json:"donor_email,omitempty"`
	PaymentMethod           *string   `json:"payment_method,omitempty"`
	MobileMoneyProvider     *string   `json:"mobile_money_provider,omitempty"`
	MobileMoneyNumberMasked *string   `json:"mobile_money_number_masked,omitempty"`
	CardLast4               *string   `json:"card_last4,omitempty"`
	CardExpiryMonth         *int      `json:"card_expiry_month,omitempty"`
	CardExpiryYear          *int      `json:"card_expiry_year,omitempty"`
	CardholderName          *string   `json:"cardholder_name,omitempty"`
	Latitude                *float64  `json:"latitude"`
	Longitude               *float64  `json:"longitude"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type CreateOfferInput struct {
	RequestID           string   `json:"request_id" validate:"required"`
	ResponderName       string   `json:"responder_name" validate:"required"`
	ResponderContact    string   `json:"responder_contact" validate:"required"`
	OfferType           string   `json:"offer_type" validate:"required"`
	ExpertiseDetails    *string  `json:"expertise_details,omitempty"`
	VehicleType         *string  `json:"vehicle_type,omitempty"`
	DonationAmount      *float64 `json:"donation_amount,omitempty" validate:"omitempty,gt=0"`
	DonorEmail          *string  `json:"donor_email,omitempty"`
	PaymentMethod       *string  `json:"payment_method,omitempty" validate:"omitempty,oneof=mobile_money visa"`
	MobileMoneyProvider *string  `json:"mobile_money_provider,omitempty" validate:"omitempty,oneof=airtel_money mtn_momo"`
	MobileMoneyNumber   *string  `json:"mobile_money_number,omitempty"`
	CardNumber          *string  `json:"card_number,omitempty"`
	CardCVC             *string  `json:"card_cvc,omitempty"`
	CardExpiryMonth     *int     `json:"card_expiry_month,omitempty"`
	CardExpiryYear      *int     `json:"card_expiry_year,omitempty"`
	CardholderName      *string  `json:"cardholder_name,omitempty"`
	Latitude            *float64 `json:"latitude"`
	Longitude           *float64 `json:"longitude"`
}

type UpdateOfferStatusInput struct {
	Status string `json:"status" validate:"required,oneof=pending accepted fulfilled"`
}

var digitsOnlyRegex = regexp.MustCompile(`\D`)

func (input *CreateOfferInput) Normalize() {
	input.RequestID = strings.TrimSpace(input.RequestID)
	input.ResponderName = strings.TrimSpace(input.ResponderName)
	input.ResponderContact = strings.TrimSpace(input.ResponderContact)
	trimStringPtr(&input.ExpertiseDetails)
	trimStringPtr(&input.VehicleType)
	trimStringPtr(&input.DonorEmail)
	trimStringPtr(&input.PaymentMethod)
	trimStringPtr(&input.MobileMoneyProvider)
	trimStringPtr(&input.MobileMoneyNumber)
	trimStringPtr(&input.CardNumber)
	trimStringPtr(&input.CardCVC)
	trimStringPtr(&input.CardholderName)
}

func (input *CreateOfferInput) ValidateBusinessRules() error {
	if !isValidContact(input.ResponderContact) {
		return fmt.Errorf("responder_contact must be a valid email address or phone number")
	}

	switch input.OfferType {
	case "expertise":
		if isBlankPtr(input.ExpertiseDetails) {
			return fmt.Errorf("expertise_details is required when offer_type is expertise")
		}
	case "transport":
		if isBlankPtr(input.VehicleType) {
			return fmt.Errorf("vehicle_type is required when offer_type is transport")
		}
	case "donation":
		if input.DonationAmount == nil || *input.DonationAmount <= 0 {
			return fmt.Errorf("donation_amount is required when offer_type is donation")
		}
		if isBlankPtr(input.DonorEmail) {
			return fmt.Errorf("donor_email is required when offer_type is donation")
		}
		if !isValidEmail(*input.DonorEmail) {
			return fmt.Errorf("donor_email must be a valid email address")
		}
		if input.PaymentMethod == nil {
			return fmt.Errorf("payment_method is required when offer_type is donation")
		}

		switch *input.PaymentMethod {
		case "mobile_money":
			if input.MobileMoneyProvider == nil {
				return fmt.Errorf("mobile_money_provider is required when payment_method is mobile_money")
			}
			if input.MobileMoneyNumber == nil {
				return fmt.Errorf("mobile_money_number is required when payment_method is mobile_money")
			}
			if !isValidPhoneNumber(*input.MobileMoneyNumber) {
				return fmt.Errorf("mobile_money_number must be a valid phone number")
			}
		case "visa":
			if input.CardNumber == nil {
				return fmt.Errorf("card_number is required when payment_method is visa")
			}
			if input.CardCVC == nil {
				return fmt.Errorf("card_cvc is required when payment_method is visa")
			}
			if input.CardExpiryMonth == nil || *input.CardExpiryMonth < 1 || *input.CardExpiryMonth > 12 {
				return fmt.Errorf("card_expiry_month must be between 1 and 12")
			}
			if input.CardExpiryYear == nil || *input.CardExpiryYear < time.Now().Year() {
				return fmt.Errorf("card_expiry_year must be this year or later")
			}
			if input.CardholderName == nil {
				return fmt.Errorf("cardholder_name is required when payment_method is visa")
			}
			if !isValidCardNumber(*input.CardNumber) {
				return fmt.Errorf("card_number must be 12 to 19 digits")
			}
			if !isValidCardCVC(*input.CardCVC) {
				return fmt.Errorf("card_cvc must be 3 or 4 digits")
			}
		}
	}

	return nil
}

func (input CreateOfferInput) MaskedMobileMoneyNumber() *string {
	if input.MobileMoneyNumber == nil {
		return nil
	}
	digits := digitsOnly(*input.MobileMoneyNumber)
	if len(digits) < 4 {
		return nil
	}
	masked := "****" + digits[len(digits)-4:]
	return &masked
}

func (input CreateOfferInput) CardNumberLast4() *string {
	if input.CardNumber == nil {
		return nil
	}
	digits := digitsOnly(*input.CardNumber)
	if len(digits) < 4 {
		return nil
	}
	last4 := digits[len(digits)-4:]
	return &last4
}

func trimStringPtr(value **string) {
	if *value == nil {
		return
	}
	trimmed := strings.TrimSpace(**value)
	if trimmed == "" {
		*value = nil
		return
	}
	*value = &trimmed
}

func isBlankPtr(value *string) bool {
	return value == nil || strings.TrimSpace(*value) == ""
}

func isValidContact(value string) bool {
	return isValidEmail(value) || isValidPhoneNumber(value)
}

func isValidEmail(value string) bool {
	_, err := mail.ParseAddress(value)
	return err == nil
}

func isValidPhoneNumber(value string) bool {
	digits := digitsOnly(value)
	return len(digits) >= 9 && len(digits) <= 15
}

func isValidCardNumber(value string) bool {
	digits := digitsOnly(value)
	return len(digits) >= 12 && len(digits) <= 19
}

func isValidCardCVC(value string) bool {
	digits := digitsOnly(value)
	return len(digits) == 3 || len(digits) == 4
}

func digitsOnly(value string) string {
	return digitsOnlyRegex.ReplaceAllString(value, "")
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
