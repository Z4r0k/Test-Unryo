package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// getUsers liste tous les usagers avec pagination, recherche et filtres
// GET /api/users
func getUsers(c *gin.Context) {
	// Paramètres de pagination
	page := 1
	limit := 10
	search := c.Query("search")

	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := (page - 1) * limit

	// Filtres par colonne
	filterNiveau := c.Query("filter_niveau")
	filterAgeMin := c.Query("filter_age_min")
	filterAgeMax := c.Query("filter_age_max")

	// Construire la requête avec recherche et filtres
	var query string
	var args []interface{}
	var whereConditions []string
	var whereArgs []interface{}

	// Recherche globale
	if search != "" {
		whereConditions = append(whereConditions, "(first_name LIKE ? OR last_name LIKE ? OR email LIKE ?)")
		searchPattern := "%" + search + "%"
		whereArgs = append(whereArgs, searchPattern, searchPattern, searchPattern)
	}

	// Filtre par niveau
	if filterNiveau != "" {
		whereConditions = append(whereConditions, "niveau_natation = ?")
		whereArgs = append(whereArgs, filterNiveau)
	}

	// Filtre par âge (calculé à partir de date_naissance)
	if filterAgeMin != "" || filterAgeMax != "" {
		// Pour SQLite, on calcule l'âge approximativement
		// On utilise une sous-requête pour calculer l'âge
		ageCondition := `(
			(julianday('now') - julianday(date_naissance)) / 365.25
		)`
		if filterAgeMin != "" {
			if minAge, err := strconv.Atoi(filterAgeMin); err == nil {
				whereConditions = append(whereConditions, ageCondition+" >= ?")
				whereArgs = append(whereArgs, float64(minAge))
			}
		}
		if filterAgeMax != "" {
			if maxAge, err := strconv.Atoi(filterAgeMax); err == nil {
				whereConditions = append(whereConditions, ageCondition+" <= ?")
				whereArgs = append(whereArgs, float64(maxAge))
			}
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	query = `SELECT id, first_name, last_name, email, date_naissance, niveau_natation, created_at 
		FROM users 
		` + whereClause + `
		ORDER BY id DESC 
		LIMIT ? OFFSET ?`
	
	args = append(whereArgs, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var dateNaissance sql.NullString
		var niveauNatation sql.NullString
		err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &dateNaissance, &niveauNatation, &u.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		u.DateNaissance = dateNaissance.String
		u.NiveauNatation = niveauNatation.String
		u.Age = calculateAge(u.DateNaissance)
		users = append(users, u)
	}

	// Compter le total (avec ou sans recherche/filtres)
	var total int
	countQuery := `SELECT COUNT(*) FROM users ` + whereClause
	err = db.QueryRow(countQuery, whereArgs...).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// S'assurer qu'on retourne toujours un tableau, même vide
	if users == nil {
		users = []User{}
	}

	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	response := UsersResponse{
		Users:      users,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// getUserByID récupère un usager par son ID
// GET /api/users/:id
func getUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	var u User
	var dateNaissance sql.NullString
	var niveauNatation sql.NullString
	err = db.QueryRow("SELECT id, first_name, last_name, email, date_naissance, niveau_natation, created_at FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &dateNaissance, &niveauNatation, &u.CreatedAt)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usager non trouvé"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u.DateNaissance = dateNaissance.String
	u.NiveauNatation = niveauNatation.String
	u.Age = calculateAge(u.DateNaissance)

	c.JSON(http.StatusOK, u)
}

// createUser crée un nouvel usager
// POST /api/users
func createUser(c *gin.Context) {
	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("INSERT INTO users (first_name, last_name, email, date_naissance, niveau_natation) VALUES (?, ?, ?, ?, ?)",
		req.FirstName, req.LastName, req.Email, req.DateNaissance, req.NiveauNatation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	var u User
	var dateNaissance sql.NullString
	var niveauNatation sql.NullString
	err = db.QueryRow("SELECT id, first_name, last_name, email, date_naissance, niveau_natation, created_at FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &dateNaissance, &niveauNatation, &u.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u.DateNaissance = dateNaissance.String
	u.NiveauNatation = niveauNatation.String
	u.Age = calculateAge(u.DateNaissance)

	c.JSON(http.StatusCreated, u)
}

// updateUser modifie un usager existant
// PUT /api/users/:id
func updateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := db.Exec("UPDATE users SET first_name = ?, last_name = ?, email = ?, date_naissance = ?, niveau_natation = ? WHERE id = ?",
		req.FirstName, req.LastName, req.Email, req.DateNaissance, req.NiveauNatation, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usager non trouvé"})
		return
	}

	var u User
	var dateNaissance sql.NullString
	var niveauNatation sql.NullString
	err = db.QueryRow("SELECT id, first_name, last_name, email, date_naissance, niveau_natation, created_at FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &dateNaissance, &niveauNatation, &u.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u.DateNaissance = dateNaissance.String
	u.NiveauNatation = niveauNatation.String
	u.Age = calculateAge(u.DateNaissance)

	c.JSON(http.StatusOK, u)
}

// deleteUser supprime un usager
// DELETE /api/users/:id
func deleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID invalide"})
		return
	}

	result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usager non trouvé"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usager supprimé avec succès"})
}

