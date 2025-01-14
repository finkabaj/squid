package types

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	DateOfBirth  time.Time `json:"date_of_birth"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthUser struct {
	User      User      `json:"user"`
	TokenPair TokenPair `json:"token_pair"`
}

type RegisterUser struct {
	Username    string    `json:"username" validate:"required,min=3,max=50"`
	FirstName   string    `json:"first_name" validate:"required,min=3,max=100"`
	LastName    string    `json:"last_name" validate:"required,min=3,max=100"`
	DateOfBirth time.Time `json:"date_of_birth" validate:"required,date_of_birth"`
	Email       string    `json:"email" validate:"required,email"`
	Password    string    `json:"password" validate:"required,min=8,max=50"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type DBCredentials struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}
