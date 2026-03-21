package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"community-aid-api/internal/models"

	"github.com/lib/pq"
)

type RequestService struct {
	db *sql.DB
}

func NewRequestService(db *sql.DB) *RequestService {
	return &RequestService{db: db}
}

// scanRequest scans a single row into an EmergencyRequest.
func scanRequest(row interface{ Scan(...any) error }) (*models.EmergencyRequest, error) {
	var r models.EmergencyRequest
	err := row.Scan(
		&r.ID, &r.UserID, &r.Title, &r.Description, &r.Type, &r.Status,
		&r.LocationName, &r.Latitude, &r.Longitude, &r.MediaURLs,
		&r.CreatedAt, &r.UpdatedAt,
	)
	return &r, err
}

const selectCols = `id, user_id, title, description, type, status, location_name, latitude, longitude, media_urls, created_at, updated_at`

func (s *RequestService) CreateRequest(ctx context.Context, userID string, input models.CreateRequestInput, mediaURLs []string) (*models.EmergencyRequest, error) {
	if mediaURLs == nil {
		mediaURLs = []string{}
	}
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO emergency_requests
			(user_id, title, description, type, status, location_name, latitude, longitude, media_urls)
		 VALUES ($1, $2, $3, $4, 'pending', $5, $6, $7, $8)
		 RETURNING `+selectCols,
		userID, input.Title, input.Description, input.Type,
		input.LocationName, input.Latitude, input.Longitude, pq.Array(mediaURLs),
	)
	r, err := scanRequest(row)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return r, nil
}

// GetAllRequests returns approved requests matching the filters, paginated.
// Returns the matching rows, total row count (before pagination), and any error.
func (s *RequestService) GetAllRequests(ctx context.Context, filters models.RequestFilters, page, pageSize int) ([]models.EmergencyRequest, int, error) {
	args := []interface{}{}
	argIdx := 1

	var statusCond string
	if filters.Status != nil {
		statusCond = fmt.Sprintf("status = $%d", argIdx)
		args = append(args, *filters.Status)
		argIdx++
	} else {
		statusCond = "status = 'approved'"
	}
	conditions := []string{statusCond}

	if filters.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *filters.Type)
		argIdx++
	}
	if filters.LocationName != nil {
		conditions = append(conditions, fmt.Sprintf("location_name ILIKE $%d", argIdx))
		args = append(args, "%"+*filters.LocationName+"%")
		argIdx++
	}
	where := strings.Join(conditions, " AND ")

	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM emergency_requests WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count requests: %w", err)
	}

	offset := (page - 1) * pageSize
	dataArgs := append(append([]interface{}{}, args...), pageSize, offset)
	dataQuery := fmt.Sprintf(
		`SELECT `+selectCols+` FROM emergency_requests WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	results, err := s.queryRequests(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// GetAllRequestsAdmin returns all requests regardless of status, paginated.
func (s *RequestService) GetAllRequestsAdmin(ctx context.Context, filters models.RequestFilters, page, pageSize int) ([]models.EmergencyRequest, int, error) {
	args := []interface{}{}
	argIdx := 1
	conditions := []string{}

	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filters.Status)
		argIdx++
	}
	if filters.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *filters.Type)
		argIdx++
	}
	if filters.LocationName != nil {
		conditions = append(conditions, fmt.Sprintf("location_name ILIKE $%d", argIdx))
		args = append(args, "%"+*filters.LocationName+"%")
		argIdx++
	}

	baseQuery := `FROM emergency_requests`
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) `+baseQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin requests: %w", err)
	}

	offset := (page - 1) * pageSize
	dataArgs := append(append([]interface{}{}, args...), pageSize, offset)
	dataQuery := fmt.Sprintf(
		`SELECT `+selectCols+` `+baseQuery+` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		argIdx, argIdx+1,
	)
	results, err := s.queryRequests(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (s *RequestService) GetRequestByID(ctx context.Context, id string) (*models.EmergencyRequest, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT `+selectCols+` FROM emergency_requests WHERE id = $1`, id,
	)
	r, err := scanRequest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get request by id: %w", err)
	}
	return r, nil
}

func (s *RequestService) GetRequestsByUserID(ctx context.Context, userID string) ([]models.EmergencyRequest, error) {
	return s.queryRequests(ctx,
		`SELECT `+selectCols+` FROM emergency_requests WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
}

func (s *RequestService) UpdateRequest(ctx context.Context, id string, input models.UpdateRequestInput) (*models.EmergencyRequest, error) {
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if input.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *input.Title)
		argIdx++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *input.Description)
		argIdx++
	}
	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *input.Status)
		argIdx++
	}
	if input.LocationName != nil {
		setClauses = append(setClauses, fmt.Sprintf("location_name = $%d", argIdx))
		args = append(args, *input.LocationName)
		argIdx++
	}
	if input.Latitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("latitude = $%d", argIdx))
		args = append(args, *input.Latitude)
		argIdx++
	}
	if input.Longitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("longitude = $%d", argIdx))
		args = append(args, *input.Longitude)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE emergency_requests SET %s WHERE id = $%d RETURNING `+selectCols,
		strings.Join(setClauses, ", "), argIdx,
	)

	row := s.db.QueryRowContext(ctx, query, args...)
	r, err := scanRequest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}
	return r, nil
}

func (s *RequestService) DeleteRequest(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM emergency_requests WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete request: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *RequestService) queryRequests(ctx context.Context, query string, args ...interface{}) ([]models.EmergencyRequest, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query requests: %w", err)
	}
	defer rows.Close()

	var results []models.EmergencyRequest
	for rows.Next() {
		r, err := scanRequest(rows)
		if err != nil {
			return nil, fmt.Errorf("scan request: %w", err)
		}
		results = append(results, *r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	if results == nil {
		results = []models.EmergencyRequest{}
	}
	return results, nil
}
