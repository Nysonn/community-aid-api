package services

import (
	"context"
	"database/sql"
	"fmt"

	"community-aid-api/internal/models"
)

type DonationService struct {
	db *sql.DB
}

func NewDonationService(db *sql.DB) *DonationService {
	return &DonationService{db: db}
}

const donationCols = `id, request_id, donor_name, amount, date, created_at`

func scanDonation(row interface{ Scan(...any) error }) (*models.Donation, error) {
	var d models.Donation
	err := row.Scan(&d.ID, &d.RequestID, &d.DonorName, &d.Amount, &d.Date, &d.CreatedAt)
	return &d, err
}

func (s *DonationService) CreateDonation(ctx context.Context, input models.CreateDonationInput) (*models.Donation, error) {
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO donations (request_id, donor_name, amount, date)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+donationCols,
		input.RequestID, input.DonorName, input.Amount, input.Date,
	)
	d, err := scanDonation(row)
	if err != nil {
		return nil, fmt.Errorf("create donation: %w", err)
	}
	return d, nil
}

func (s *DonationService) GetDonationsByRequestID(ctx context.Context, requestID string) ([]models.Donation, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+donationCols+` FROM donations WHERE request_id = $1 ORDER BY date DESC`,
		requestID,
	)
	if err != nil {
		return nil, fmt.Errorf("get donations by request id: %w", err)
	}
	defer rows.Close()

	var results []models.Donation
	for rows.Next() {
		d, err := scanDonation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan donation: %w", err)
		}
		results = append(results, *d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	if results == nil {
		results = []models.Donation{}
	}
	return results, nil
}

// GetAllDonationsAdmin returns all donations joined with their request title, paginated.
// Returns rows, total count, and error.
func (s *DonationService) GetAllDonationsAdmin(ctx context.Context, page, pageSize int) ([]models.DonationWithRequest, int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM donations`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count donations: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx,
		`SELECT d.id, d.request_id, d.donor_name, d.amount, d.date, d.created_at,
		        r.title AS request_title
		 FROM donations d
		 JOIN emergency_requests r ON r.id = d.request_id
		 ORDER BY d.date DESC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("get all donations admin: %w", err)
	}
	defer rows.Close()

	var results []models.DonationWithRequest
	for rows.Next() {
		var dwr models.DonationWithRequest
		if err := rows.Scan(
			&dwr.ID, &dwr.RequestID, &dwr.DonorName, &dwr.Amount, &dwr.Date, &dwr.CreatedAt,
			&dwr.RequestTitle,
		); err != nil {
			return nil, 0, fmt.Errorf("scan donation with request: %w", err)
		}
		results = append(results, dwr)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}
	if results == nil {
		results = []models.DonationWithRequest{}
	}
	return results, total, nil
}
