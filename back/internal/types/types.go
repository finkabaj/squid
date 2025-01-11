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

type RegisterUser struct {
	Username    string    `json:"username" validate:"required,min=3,max=50"`
	FirstName   string    `json:"first_name" validate:"required,min=3,max=100"`
	LastName    string    `json:"last_name" validate:"required,min=3,max=100"`
	DateOfBirth time.Time `json:"date_of_birth" validate:"required,date"`
	Email       string    `json:"email" validate:"required,email"`
	Password    string    `json:"password" validate:"required,min=8,max=100"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type JWT struct {
	UserID    string
	Email     string
	Username  string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TokenHash string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type DBCredentials struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}
