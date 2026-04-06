package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"community-aid-api/internal/models"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

const userCols = `id, clerk_id, full_name, email, phone_number, bio, profile_image_url, role, is_active, created_at, updated_at`

func scanUser(row interface{ Scan(...any) error }) (*models.User, error) {
	var u models.User
	err := row.Scan(
		&u.ID, &u.ClerkID, &u.FullName, &u.Email, &u.PhoneNumber,
		&u.Bio, &u.ProfileImageURL, &u.Role, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	return &u, err
}

func (s *UserService) CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	role := "community_member"
	if input.Role != "" {
		role = input.Role
	}

	phone := ""
	if input.PhoneNumber != nil {
		phone = *input.PhoneNumber
	}

	u, err := scanUser(s.db.QueryRowContext(ctx,
		`INSERT INTO users (clerk_id, full_name, email, phone_number, role, profile_image_url)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (clerk_id) DO NOTHING
		 RETURNING `+userCols,
		input.ClerkID, input.FullName, input.Email, phone, role, input.ProfileImageURL,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return s.GetUserByClerkID(ctx, input.ClerkID)
	}
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (s *UserService) GetUserByClerkID(ctx context.Context, clerkID string) (*models.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		`SELECT `+userCols+` FROM users WHERE clerk_id = $1`, clerkID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by clerk id: %w", err)
	}
	return u, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		`SELECT `+userCols+` FROM users WHERE id = $1`, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		`SELECT `+userCols+` FROM users WHERE email = $1`, email,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// GetAllUsers returns all users paginated. Returns rows, total count, and error.
func (s *UserService) GetAllUsers(ctx context.Context, page, pageSize int) ([]models.User, int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+userCols+` FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("get all users: %w", err)
	}
	defer rows.Close()

	var results []models.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		results = append(results, *u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}
	if results == nil {
		results = []models.User{}
	}
	return results, total, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, input models.UpdateUserInput) (*models.User, error) {
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if input.FullName != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", argIdx))
		args = append(args, *input.FullName)
		argIdx++
	}
	if input.PhoneNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone_number = $%d", argIdx))
		args = append(args, *input.PhoneNumber)
		argIdx++
	}
	if input.Bio != nil {
		setClauses = append(setClauses, fmt.Sprintf("bio = $%d", argIdx))
		args = append(args, *input.Bio)
		argIdx++
	}
	if input.ProfileImageURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("profile_image_url = $%d", argIdx))
		args = append(args, *input.ProfileImageURL)
		argIdx++
	}

	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE users SET %s WHERE id = $%d RETURNING `+userCols,
		strings.Join(setClauses, ", "), argIdx,
	)

	u, err := scanUser(s.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return u, nil
}

func (s *UserService) SetUserActiveStatus(ctx context.Context, id string, isActive bool) (*models.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		`UPDATE users SET is_active = $1, updated_at = NOW() WHERE id = $2 RETURNING `+userCols,
		isActive, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("set user active status: %w", err)
	}
	return u, nil
}

func (s *UserService) SetUserRole(ctx context.Context, id, role string) (*models.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		`UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2 RETURNING `+userCols,
		role, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("set user role: %w", err)
	}
	return u, nil
}

func (s *UserService) UploadUserAvatar(ctx context.Context, id, imageURL string) (*models.User, error) {
	u, err := scanUser(s.db.QueryRowContext(ctx,
		`UPDATE users SET profile_image_url = $1, updated_at = NOW() WHERE id = $2 RETURNING `+userCols,
		imageURL, id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("upload user avatar: %w", err)
	}
	return u, nil
}
