package main

import "time"

// User is the public representation of an account (never includes the hash).
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Email holds the login identifier — a username or an email address.
type registerRequest struct {
	Email    string `json:"email" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type tokenResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type verifyResponse struct {
	Valid bool   `json:"valid"`
	Sub   string `json:"sub"`
	Email string `json:"email"`
}
