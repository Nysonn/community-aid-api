package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"community-aid-api/internal/models"
)

type OfferService struct {
	db *sql.DB
}

func NewOfferService(db *sql.DB) *OfferService {
	return &OfferService{db: db}
}

const offerCols = `id, request_id, responder_name, responder_contact, offer_type, status, latitude, longitude, created_at, updated_at`

func scanOffer(row interface{ Scan(...any) error }) (*models.Offer, error) {
	var o models.Offer
	err := row.Scan(
		&o.ID, &o.RequestID, &o.ResponderName, &o.ResponderContact,
		&o.OfferType, &o.Status, &o.Latitude, &o.Longitude,
		&o.CreatedAt, &o.UpdatedAt,
	)
	return &o, err
}

func (s *OfferService) CreateOffer(ctx context.Context, input models.CreateOfferInput) (*models.Offer, error) {
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO offers (request_id, responder_name, responder_contact, offer_type, status, latitude, longitude)
		 VALUES ($1, $2, $3, $4, 'pending', $5, $6)
		 RETURNING `+offerCols,
		input.RequestID, input.ResponderName, input.ResponderContact, input.OfferType,
		input.Latitude, input.Longitude,
	)
	o, err := scanOffer(row)
	if err != nil {
		return nil, fmt.Errorf("create offer: %w", err)
	}
	return o, nil
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
