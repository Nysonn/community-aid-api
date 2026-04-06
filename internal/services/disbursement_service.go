package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"community-aid-api/internal/models"
)

type DisbursementService struct {
	db *sql.DB
}

func NewDisbursementService(db *sql.DB) *DisbursementService {
	return &DisbursementService{db: db}
}

const disbursementCols = `
	d.id, d.offer_id, d.request_id, d.donor_name, d.donor_email, d.amount, d.recipient_name,
	d.payment_type, d.bank_account_name, d.bank_account_number, d.bank_name,
	d.receiving_mobile_provider, d.receiving_mobile_number,
	d.status, d.disbursed_at, d.created_at,
	r.title AS request_title`

func scanDisbursement(row interface{ Scan(...any) error }) (*models.Disbursement, error) {
	var d models.Disbursement
	var bankAccountName, bankAccountNumber, bankName sql.NullString
	var receivingMobileProvider, receivingMobileNumber sql.NullString
	var disbursedAt sql.NullTime
	err := row.Scan(
		&d.ID, &d.OfferID, &d.RequestID, &d.DonorName, &d.DonorEmail, &d.Amount, &d.RecipientName,
		&d.PaymentType, &bankAccountName, &bankAccountNumber, &bankName,
		&receivingMobileProvider, &receivingMobileNumber,
		&d.Status, &disbursedAt, &d.CreatedAt,
		&d.RequestTitle,
	)
	if err != nil {
		return nil, err
	}
	if bankAccountName.Valid {
		d.BankAccountName = &bankAccountName.String
	}
	if bankAccountNumber.Valid {
		d.BankAccountNumber = &bankAccountNumber.String
	}
	if bankName.Valid {
		d.BankName = &bankName.String
	}
	if receivingMobileProvider.Valid {
		d.ReceivingMobileProvider = &receivingMobileProvider.String
	}
	if receivingMobileNumber.Valid {
		d.ReceivingMobileNumber = &receivingMobileNumber.String
	}
	if disbursedAt.Valid {
		d.Disbursedat = &disbursedAt.Time
	}
	return &d, nil
}

// GetAllDisbursements returns all disbursements paginated, joined with request title.
func (s *DisbursementService) GetAllDisbursements(ctx context.Context, status string, page, pageSize int) ([]models.Disbursement, int, error) {
	var (
		countQuery string
		dataQuery  string
		args       []interface{}
	)

	offset := (page - 1) * pageSize

	if status != "" {
		countQuery = `SELECT COUNT(*) FROM disbursements d WHERE d.status = $1`
		dataQuery = `SELECT ` + disbursementCols + `
			FROM disbursements d
			JOIN emergency_requests r ON r.id = d.request_id
			WHERE d.status = $1
			ORDER BY d.created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{status, pageSize, offset}
	} else {
		countQuery = `SELECT COUNT(*) FROM disbursements d`
		dataQuery = `SELECT ` + disbursementCols + `
			FROM disbursements d
			JOIN emergency_requests r ON r.id = d.request_id
			ORDER BY d.created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{pageSize, offset}
	}

	var total int
	countArgs := args[:len(args)-2] // strip limit/offset for count
	if status != "" {
		countArgs = []interface{}{status}
	} else {
		countArgs = []interface{}{}
	}
	if err := s.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count disbursements: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("get disbursements: %w", err)
	}
	defer rows.Close()

	var results []models.Disbursement
	for rows.Next() {
		d, err := scanDisbursement(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan disbursement: %w", err)
		}
		results = append(results, *d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}
	if results == nil {
		results = []models.Disbursement{}
	}
	return results, total, nil
}

// GetDisbursementByID returns a single disbursement by ID.
func (s *DisbursementService) GetDisbursementByID(ctx context.Context, id string) (*models.Disbursement, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+disbursementCols+`
		 FROM disbursements d
		 JOIN emergency_requests r ON r.id = d.request_id
		 WHERE d.id = $1`, id,
	)
	d, err := scanDisbursement(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get disbursement by id: %w", err)
	}
	return d, nil
}

// MarkDisbursed marks a disbursement as disbursed and records the timestamp.
func (s *DisbursementService) MarkDisbursed(ctx context.Context, id string) (*models.Disbursement, error) {
	now := time.Now().UTC()
	row := s.db.QueryRowContext(ctx,
		`UPDATE disbursements d
		 SET status = 'disbursed', disbursed_at = $1
		 FROM emergency_requests r
		 WHERE r.id = d.request_id AND d.id = $2
		 RETURNING `+disbursementCols,
		now, id,
	)
	d, err := scanDisbursement(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("mark disbursed: %w", err)
	}
	return d, nil
}
