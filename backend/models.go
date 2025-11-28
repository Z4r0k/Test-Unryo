package main

import "time"

// User représente un usager
type User struct {
	ID            int       `json:"id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Email         string    `json:"email"`
	DateNaissance string    `json:"date_naissance"` // Format: YYYY-MM-DD
	Age           int       `json:"age"`            // Calculé à partir de date_naissance
	NiveauNatation string   `json:"niveau_natation"`
	CreatedAt     time.Time `json:"created_at"`
}

// UserRequest représente les données pour créer/modifier un usager
type UserRequest struct {
	FirstName     string `json:"first_name" binding:"required"`
	LastName      string `json:"last_name" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
	DateNaissance string `json:"date_naissance" binding:"required"`
	NiveauNatation string `json:"niveau_natation" binding:"required"`
}

// UsersResponse représente la réponse paginée
type UsersResponse struct {
	Users      []User `json:"users"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalPages int    `json:"total_pages"`
}

