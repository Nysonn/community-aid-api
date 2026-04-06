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

// scanRequest scans a single row into an EmergencyRequest (no poster_name column).
func scanRequest(row interface{ Scan(...any) error }) (*models.EmergencyRequest, error) {
	return scanRequestFromRow(row, false)
}

const selectCols = `id, user_id, title, description, type, status, location_name, latitude, longitude, media_urls,
	target_amount, payment_type, bank_account_name, bank_account_number, bank_name,
	receiving_mobile_provider, receiving_mobile_number,
	COALESCE((SELECT SUM(amount) FROM donations WHERE request_id = emergency_requests.id), 0) AS amount_received,
	created_at, updated_at`

// selectColsAdmin adds the poster's full name via a LEFT JOIN with users.
const selectColsAdmin = `er.id, er.user_id, er.title, er.description, er.type, er.status,
	er.location_name, er.latitude, er.longitude, er.media_urls,
	er.target_amount, er.payment_type, er.bank_account_name, er.bank_account_number, er.bank_name,
	er.receiving_mobile_provider, er.receiving_mobile_number,
	COALESCE((SELECT SUM(amount) FROM donations WHERE request_id = er.id), 0) AS amount_received,
	er.created_at, er.updated_at,
	COALESCE(u.full_name, '') AS poster_name`

func scanRequestWithPoster(row interface{ Scan(...any) error }) (*models.EmergencyRequest, error) {
	r, err := scanRequestFromRow(row, true)
	return r, err
}

// scanRequestFromRow is the shared scanner; withPoster controls whether it reads the extra poster_name column.
func scanRequestFromRow(row interface{ Scan(...any) error }, withPoster bool) (*models.EmergencyRequest, error) {
	var r models.EmergencyRequest
	var targetAmount sql.NullFloat64
	var paymentType sql.NullString
	var bankAccountName, bankAccountNumber, bankName sql.NullString
	var receivingMobileProvider, receivingMobileNumber sql.NullString
	var lat, lng sql.NullFloat64

	dest := []any{
		&r.ID, &r.UserID, &r.Title, &r.Description, &r.Type, &r.Status,
		&r.LocationName, &lat, &lng, &r.MediaURLs,
		&targetAmount,
		&paymentType, &bankAccountName, &bankAccountNumber, &bankName,
		&receivingMobileProvider, &receivingMobileNumber,
		&r.AmountReceived,
		&r.CreatedAt, &r.UpdatedAt,
	}
	if withPoster {
		dest = append(dest, &r.PosterName)
	}

	if err := row.Scan(dest...); err != nil {
		return nil, err
	}
	if lat.Valid {
		r.Latitude = &lat.Float64
	}
	if lng.Valid {
		r.Longitude = &lng.Float64
	}
	if targetAmount.Valid {
		r.TargetAmount = &targetAmount.Float64
	}
	if paymentType.Valid {
		r.PaymentType = &paymentType.String
	}
	if bankAccountName.Valid {
		r.BankAccountName = &bankAccountName.String
	}
	if bankAccountNumber.Valid {
		r.BankAccountNumber = &bankAccountNumber.String
	}
	if bankName.Valid {
		r.BankName = &bankName.String
	}
	if receivingMobileProvider.Valid {
		r.ReceivingMobileProvider = &receivingMobileProvider.String
	}
	if receivingMobileNumber.Valid {
		r.ReceivingMobileNumber = &receivingMobileNumber.String
	}
	return &r, nil
}

func (s *RequestService) CreateRequest(ctx context.Context, userID string, input models.CreateRequestInput, mediaURLs []string) (*models.EmergencyRequest, error) {
	if mediaURLs == nil {
		mediaURLs = []string{}
	}
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO emergency_requests
			(user_id, title, description, type, status, location_name, latitude, longitude, media_urls,
			 target_amount, payment_type, bank_account_name, bank_account_number, bank_name,
			 receiving_mobile_provider, receiving_mobile_number)
		 VALUES ($1, $2, $3, $4, 'pending', $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		 RETURNING `+selectCols,
		userID, input.Title, input.Description, input.Type,
		input.LocationName, input.Latitude, input.Longitude, pq.Array(mediaURLs),
		input.TargetAmount, input.PaymentType,
		input.BankAccountName, input.BankAccountNumber, input.BankName,
		input.ReceivingMobileProvider, input.ReceivingMobileNumber,
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

// GetAllRequestsAdmin returns all requests regardless of status, paginated, with poster name.
func (s *RequestService) GetAllRequestsAdmin(ctx context.Context, filters models.RequestFilters, page, pageSize int) ([]models.EmergencyRequest, int, error) {
	args := []interface{}{}
	argIdx := 1
	conditions := []string{}

	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("er.status = $%d", argIdx))
		args = append(args, *filters.Status)
		argIdx++
	}
	if filters.Type != nil {
		conditions = append(conditions, fmt.Sprintf("er.type = $%d", argIdx))
		args = append(args, *filters.Type)
		argIdx++
	}
	if filters.LocationName != nil {
		conditions = append(conditions, fmt.Sprintf("er.location_name ILIKE $%d", argIdx))
		args = append(args, "%"+*filters.LocationName+"%")
		argIdx++
	}

	baseQuery := `FROM emergency_requests er LEFT JOIN users u ON u.id = er.user_id`
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
		`SELECT `+selectColsAdmin+` `+baseQuery+` ORDER BY er.created_at DESC LIMIT $%d OFFSET $%d`,
		argIdx, argIdx+1,
	)

	rows, err := s.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query admin requests: %w", err)
	}
	defer rows.Close()

	var results []models.EmergencyRequest
	for rows.Next() {
		r, err := scanRequestWithPoster(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan admin request: %w", err)
		}
		results = append(results, *r)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}
	if results == nil {
		results = []models.EmergencyRequest{}
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
	if input.TargetAmount != nil {
		setClauses = append(setClauses, fmt.Sprintf("target_amount = $%d", argIdx))
		args = append(args, *input.TargetAmount)
		argIdx++
	}
	if input.PaymentType != nil {
		setClauses = append(setClauses, fmt.Sprintf("payment_type = $%d", argIdx))
		args = append(args, *input.PaymentType)
		argIdx++
	}
	if input.BankAccountName != nil {
		setClauses = append(setClauses, fmt.Sprintf("bank_account_name = $%d", argIdx))
		args = append(args, *input.BankAccountName)
		argIdx++
	}
	if input.BankAccountNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("bank_account_number = $%d", argIdx))
		args = append(args, *input.BankAccountNumber)
		argIdx++
	}
	if input.BankName != nil {
		setClauses = append(setClauses, fmt.Sprintf("bank_name = $%d", argIdx))
		args = append(args, *input.BankName)
		argIdx++
	}
	if input.ReceivingMobileProvider != nil {
		setClauses = append(setClauses, fmt.Sprintf("receiving_mobile_provider = $%d", argIdx))
		args = append(args, *input.ReceivingMobileProvider)
		argIdx++
	}
	if input.ReceivingMobileNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("receiving_mobile_number = $%d", argIdx))
		args = append(args, *input.ReceivingMobileNumber)
		argIdx++
	}
	if input.Longitude != nil {
		setClauses = append(setClauses, fmt.Sprintf("longitude = $%d", argIdx))
		args = append(args, *input.Longitude)
		argIdx++
	}
	if input.TargetAmount != nil {
		setClauses = append(setClauses, fmt.Sprintf("target_amount = $%d", argIdx))
		args = append(args, *input.TargetAmount)
		argIdx++
	}
	if input.PaymentType != nil {
		setClauses = append(setClauses, fmt.Sprintf("payment_type = $%d", argIdx))
		args = append(args, *input.PaymentType)
		argIdx++
	}
	if input.BankAccountName != nil {
		setClauses = append(setClauses, fmt.Sprintf("bank_account_name = $%d", argIdx))
		args = append(args, *input.BankAccountName)
		argIdx++
	}
	if input.BankAccountNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("bank_account_number = $%d", argIdx))
		args = append(args, *input.BankAccountNumber)
		argIdx++
	}
	if input.BankName != nil {
		setClauses = append(setClauses, fmt.Sprintf("bank_name = $%d", argIdx))
		args = append(args, *input.BankName)
		argIdx++
	}
	if input.ReceivingMobileProvider != nil {
		setClauses = append(setClauses, fmt.Sprintf("receiving_mobile_provider = $%d", argIdx))
		args = append(args, *input.ReceivingMobileProvider)
		argIdx++
	}
	if input.ReceivingMobileNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("receiving_mobile_number = $%d", argIdx))
		args = append(args, *input.ReceivingMobileNumber)
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
