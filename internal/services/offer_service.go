package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"community-aid-api/internal/models"
)

type OfferService struct {
	db *sql.DB
}

func NewOfferService(db *sql.DB) *OfferService {
	return &OfferService{db: db}
}

const offerCols = `id, request_id, responder_name, responder_contact, offer_type, status,
	expertise_details, vehicle_type, donation_amount, donor_email, payment_method, mobile_money_provider,
	mobile_money_number_masked, card_last4, card_expiry_month, card_expiry_year, cardholder_name,
	latitude, longitude, created_at, updated_at`

func scanOffer(row interface{ Scan(...any) error }) (*models.Offer, error) {
	var o models.Offer
	var expertiseDetails sql.NullString
	var vehicleType sql.NullString
	var donationAmount sql.NullFloat64
	var donorEmail sql.NullString
	var paymentMethod sql.NullString
	var mobileMoneyProvider sql.NullString
	var mobileMoneyNumberMasked sql.NullString
	var cardLast4 sql.NullString
	var cardExpiryMonth sql.NullInt64
	var cardExpiryYear sql.NullInt64
	var cardholderName sql.NullString
	var latitude sql.NullFloat64
	var longitude sql.NullFloat64
	err := row.Scan(
		&o.ID, &o.RequestID, &o.ResponderName, &o.ResponderContact,
		&o.OfferType, &o.Status,
		&expertiseDetails, &vehicleType, &donationAmount, &donorEmail, &paymentMethod, &mobileMoneyProvider,
		&mobileMoneyNumberMasked, &cardLast4, &cardExpiryMonth, &cardExpiryYear, &cardholderName,
		&latitude, &longitude,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if expertiseDetails.Valid {
		o.ExpertiseDetails = &expertiseDetails.String
	}
	if vehicleType.Valid {
		o.VehicleType = &vehicleType.String
	}
	if donationAmount.Valid {
		o.DonationAmount = &donationAmount.Float64
	}
	if donorEmail.Valid {
		o.DonorEmail = &donorEmail.String
	}
	if paymentMethod.Valid {
		o.PaymentMethod = &paymentMethod.String
	}
	if mobileMoneyProvider.Valid {
		o.MobileMoneyProvider = &mobileMoneyProvider.String
	}
	if mobileMoneyNumberMasked.Valid {
		o.MobileMoneyNumberMasked = &mobileMoneyNumberMasked.String
	}
	if cardLast4.Valid {
		o.CardLast4 = &cardLast4.String
	}
	if cardExpiryMonth.Valid {
		month := int(cardExpiryMonth.Int64)
		o.CardExpiryMonth = &month
	}
	if cardExpiryYear.Valid {
		year := int(cardExpiryYear.Int64)
		o.CardExpiryYear = &year
	}
	if cardholderName.Valid {
		o.CardholderName = &cardholderName.String
	}
	if latitude.Valid {
		o.Latitude = &latitude.Float64
	}
	if longitude.Valid {
		o.Longitude = &longitude.Float64
	}
	return &o, err
}

func (s *OfferService) CreateOffer(ctx context.Context, input models.CreateOfferInput, req *models.EmergencyRequest, recipientName string) (*models.Offer, error) {
	if input.OfferType == "donation" {
		if err := validateDonationDisbursementDetails(req); err != nil {
			return nil, err
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create offer tx: %w", err)
	}
	defer tx.Rollback()

	// Donation offers are immediately accepted (money already sent to admin).
	initialStatus := "pending"
	if input.OfferType == "donation" {
		initialStatus = "accepted"
	}

	row := tx.QueryRowContext(ctx,
		`INSERT INTO offers (
			request_id, responder_name, responder_contact, offer_type, status,
			expertise_details, vehicle_type, donation_amount, donor_email, payment_method, mobile_money_provider,
			mobile_money_number_masked, card_last4, card_expiry_month, card_expiry_year, cardholder_name,
			latitude, longitude
		)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		 RETURNING `+offerCols,
		input.RequestID,
		input.ResponderName,
		input.ResponderContact,
		input.OfferType,
		initialStatus,
		input.ExpertiseDetails,
		input.VehicleType,
		input.DonationAmount,
		input.DonorEmail,
		input.PaymentMethod,
		input.MobileMoneyProvider,
		input.MaskedMobileMoneyNumber(),
		input.CardNumberLast4(),
		input.CardExpiryMonth,
		input.CardExpiryYear,
		input.CardholderName,
		input.Latitude,
		input.Longitude,
	)
	o, err := scanOffer(row)
	if err != nil {
		return nil, fmt.Errorf("create offer: %w", err)
	}

	if input.OfferType == "donation" && input.DonationAmount != nil {
		// Record the donation
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO donations (request_id, donor_name, amount, date)
			 VALUES ($1, $2, $3, $4)`,
			input.RequestID,
			input.ResponderName,
			*input.DonationAmount,
			time.Now().UTC().Format("2006-01-02"),
		); err != nil {
			return nil, fmt.Errorf("create linked donation: %w", err)
		}

		// Create a pending disbursement for the admin to process
		donorEmail := ""
		if input.DonorEmail != nil {
			donorEmail = *input.DonorEmail
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO disbursements (
				offer_id, request_id, donor_name, donor_email, amount, recipient_name,
				payment_type, bank_account_name, bank_account_number, bank_name,
				receiving_mobile_provider, receiving_mobile_number
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			o.ID, input.RequestID,
			input.ResponderName, donorEmail, *input.DonationAmount, recipientName,
			nullableStr(req.PaymentType),
			nullableStr(req.BankAccountName),
			nullableStr(req.BankAccountNumber),
			nullableStr(req.BankName),
			nullableStr(req.ReceivingMobileProvider),
			nullableStr(req.ReceivingMobileNumber),
		); err != nil {
			return nil, fmt.Errorf("create disbursement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create offer: %w", err)
	}
	return o, nil
}

func nullableStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func validateDonationDisbursementDetails(req *models.EmergencyRequest) error {
	if req == nil {
		return &BadRequestError{Message: "request not found"}
	}

	if req.PaymentType == nil {
		return &BadRequestError{Message: "this request is not configured to receive donations"}
	}

	switch strings.TrimSpace(*req.PaymentType) {
	case "bank":
		if isBlankStringPtr(req.BankAccountName) || isBlankStringPtr(req.BankAccountNumber) || isBlankStringPtr(req.BankName) {
			return &BadRequestError{Message: "this request is missing bank disbursement details"}
		}
	case "mobile_money":
		if isBlankStringPtr(req.ReceivingMobileProvider) || isBlankStringPtr(req.ReceivingMobileNumber) {
			return &BadRequestError{Message: "this request is missing mobile money disbursement details"}
		}
	default:
		return &BadRequestError{Message: "this request has an invalid payment configuration"}
	}

	return nil
}

func isBlankStringPtr(value *string) bool {
	return value == nil || strings.TrimSpace(*value) == ""
}

func (s *OfferService) GetOffersByRequestID(ctx context.Context, requestID string) ([]models.Offer, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+offerCols+` FROM offers WHERE request_id = $1 ORDER BY created_at ASC`,
		requestID,
	)
	if err != nil {
		return nil, fmt.Errorf("get offers by request id: %w", err)
	}
	defer rows.Close()
	return scanOffers(rows)
}

func (s *OfferService) GetOfferByID(ctx context.Context, id string) (*models.Offer, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+offerCols+` FROM offers WHERE id = $1`, id,
	)
	o, err := scanOffer(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get offer by id: %w", err)
	}
	return o, nil
}

func (s *OfferService) UpdateOfferStatus(ctx context.Context, id, status string) (*models.Offer, error) {
	row := s.db.QueryRowContext(ctx,
		`UPDATE offers SET status = $1, updated_at = NOW() WHERE id = $2 RETURNING `+offerCols,
		status, id,
	)
	o, err := scanOffer(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update offer status: %w", err)
	}
	return o, nil
}

// GetAllOffersAdmin returns all offers paginated. Returns rows, total count, and error.
func (s *OfferService) GetAllOffersAdmin(ctx context.Context, page, pageSize int) ([]models.Offer, int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM offers`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count offers: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+offerCols+` FROM offers ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("get all offers admin: %w", err)
	}
	defer rows.Close()

	results, err := scanOffers(rows)
	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func scanOffers(rows *sql.Rows) ([]models.Offer, error) {
	var results []models.Offer
	for rows.Next() {
		o, err := scanOffer(rows)
		if err != nil {
			return nil, fmt.Errorf("scan offer: %w", err)
		}
		results = append(results, *o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	if results == nil {
		results = []models.Offer{}
	}
	return results, nil
}
