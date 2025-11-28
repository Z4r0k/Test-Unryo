package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

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

var db *sql.DB

// calculateAge calcule l'âge à partir d'une date de naissance (format YYYY-MM-DD)
func calculateAge(dateNaissance string) int {
	if dateNaissance == "" {
		return 0
	}
	birthDate, err := time.Parse("2006-01-02", dateNaissance)
	if err != nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	return age
}

func initDB() {
	var err error
	// Utiliser un chemin dans /app/data pour la persistance dans Docker
	dbPath := "./data/users.db"
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		os.MkdirAll("./data", 0755)
	}
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Erreur lors de l'ouverture de la base de données:", err)
	}

	// Créer la table si elle n'existe pas
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		date_naissance TEXT NOT NULL,
		niveau_natation TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Erreur lors de la création de la table:", err)
	}

	// Ajouter les nouvelles colonnes si elles n'existent pas (migration)
	db.Exec("ALTER TABLE users ADD COLUMN date_naissance TEXT")
	db.Exec("ALTER TABLE users ADD COLUMN niveau_natation TEXT")
	
	// Fonction pour calculer l'âge
	db.Exec(`
		CREATE TRIGGER IF NOT EXISTS calculate_age 
		AFTER INSERT ON users
		BEGIN
			-- L'âge sera calculé côté application
		END;
	`)
}

func setupCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// UsersResponse représente la réponse paginée
type UsersResponse struct {
	Users      []User `json:"users"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalPages int    `json:"total_pages"`
}

// GET /api/users - Liste tous les usagers avec pagination et recherche
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

// GET /api/users/:id - Récupère un usager par ID
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

// POST /api/users - Crée un nouvel usager
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

// PUT /api/users/:id - Modifie un usager existant
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

// DELETE /api/users/:id - Supprime un usager
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

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()
	r.Use(setupCORS())

	// Routes API
	api := r.Group("/api")
	{
		api.GET("/users", getUsers)
		api.GET("/users/:id", getUserByID)
		api.POST("/users", createUser)
		api.PUT("/users/:id", updateUser)
		api.DELETE("/users/:id", deleteUser)
	}

	// Servir les fichiers statiques du frontend
	r.Static("/static", "./frontend/static")
	r.StaticFile("/", "./frontend/index.html")

	log.Println("Serveur démarré sur le port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}
