package models

import "time"

type User struct {
	ID              string    `json:"id"`
	ClerkID         string    `json:"clerk_id"`
	FullName        string    `json:"full_name"`
	Email           string    `json:"email"`
	PhoneNumber     string    `json:"phone_number"`
	Bio             *string   `json:"bio"`
	ProfileImageURL *string   `json:"profile_image_url"`
	Role            string    `json:"role"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateUserInput struct {
	ClerkID         string  `json:"clerk_id"          validate:"required"`
	FullName        string  `json:"full_name"         validate:"required"`
	Email           string  `json:"email"             validate:"required,email"`
	PhoneNumber     *string `json:"phone_number"`
	Role            string  `json:"role"`
	ProfileImageURL *string `json:"profile_image_url"`
}

type UpdateUserInput struct {
	FullName        *string `json:"full_name"`
	PhoneNumber     *string `json:"phone_number"`
	Bio             *string `json:"bio"`
	ProfileImageURL *string `json:"profile_image_url"`
}
