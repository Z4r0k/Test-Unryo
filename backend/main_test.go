package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Créer une base de données temporaire pour les tests
	dbPath := ":memory:" // Base de données en mémoire
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Erreur lors de l'ouverture de la base de données de test: %v", err)
	}

	// Créer la table
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
		t.Fatalf("Erreur lors de la création de la table de test: %v", err)
	}

	return db
}

func setupRouter(testDB *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(setupCORS())

	api := r.Group("/api")
	{
		api.GET("/users", func(c *gin.Context) {
			originalDB := db
			db = testDB
			getUsers(c)
			db = originalDB
		})
		api.GET("/users/:id", func(c *gin.Context) {
			originalDB := db
			db = testDB
			getUserByID(c)
			db = originalDB
		})
		api.POST("/users", func(c *gin.Context) {
			originalDB := db
			db = testDB
			createUser(c)
			db = originalDB
		})
		api.PUT("/users/:id", func(c *gin.Context) {
			originalDB := db
			db = testDB
			updateUser(c)
			db = originalDB
		})
		api.DELETE("/users/:id", func(c *gin.Context) {
			originalDB := db
			db = testDB
			deleteUser(c)
			db = originalDB
		})
	}

	return r
}

func TestCalculateAge(t *testing.T) {
	tests := []struct {
		name           string
		dateNaissance  string
		expectedAge    int
		expectedError  bool
	}{
		{
			name:          "Date valide - 10 ans",
			dateNaissance: time.Now().AddDate(-10, 0, 0).Format("2006-01-02"),
			expectedAge:   10,
			expectedError: false,
		},
		{
			name:          "Date valide - 5 ans",
			dateNaissance: time.Now().AddDate(-5, -6, 0).Format("2006-01-02"),
			expectedAge:   5,
			expectedError: false,
		},
		{
			name:          "Date invalide",
			dateNaissance: "invalid-date",
			expectedAge:   0,
			expectedError: true,
		},
		{
			name:          "Date vide",
			dateNaissance: "",
			expectedAge:   0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			age := calculateAge(tt.dateNaissance)
			if !tt.expectedError {
				// Pour les dates valides, on accepte une différence de 1 an due au calcul
				assert.GreaterOrEqual(t, age, tt.expectedAge-1)
				assert.LessOrEqual(t, age, tt.expectedAge+1)
			} else {
				assert.Equal(t, 0, age)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	r := setupRouter(testDB)

	// Test création d'un usager valide
	userData := UserRequest{
		FirstName:     "Jean",
		LastName:      "Dupont",
		Email:         "jean.dupont@test.com",
		DateNaissance: "2010-05-15",
		NiveauNatation: "NAGEUR 3",
	}

	jsonData, _ := json.Marshal(userData)
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var user User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, userData.FirstName, user.FirstName)
	assert.Equal(t, userData.LastName, user.LastName)
	assert.Equal(t, userData.Email, user.Email)
	assert.Equal(t, userData.DateNaissance, user.DateNaissance)
	assert.Equal(t, userData.NiveauNatation, user.NiveauNatation)
	assert.Greater(t, user.ID, 0)
	assert.Greater(t, user.Age, 0)
}

func TestCreateUserInvalidData(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	r := setupRouter(testDB)

	// Test avec données invalides (email manquant)
	userData := UserRequest{
		FirstName:     "Jean",
		LastName:      "Dupont",
		Email:         "", // Email manquant
		DateNaissance: "2010-05-15",
		NiveauNatation: "NAGEUR 3",
	}

	jsonData, _ := json.Marshal(userData)
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUsers(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Insérer des données de test
	testDB.Exec(`INSERT INTO users (first_name, last_name, email, date_naissance, niveau_natation) 
		VALUES ('Jean', 'Dupont', 'jean@test.com', '2010-05-15', 'NAGEUR 3'),
		       ('Marie', 'Martin', 'marie@test.com', '2012-03-20', 'PRÉSCOLAIRE 2')`)

	r := setupRouter(testDB)

	req, _ := http.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response UsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, 2, len(response.Users))
}

func TestGetUsersWithPagination(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Insérer 15 usagers
	for i := 0; i < 15; i++ {
		testDB.Exec(`INSERT INTO users (first_name, last_name, email, date_naissance, niveau_natation) 
			VALUES (?, ?, ?, '2010-05-15', 'NAGEUR 3')`, 
			"User", "Test", "user"+strconv.Itoa(i)+"@test.com")
	}

	r := setupRouter(testDB)

	// Test pagination page 1
	req, _ := http.NewRequest("GET", "/api/users?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response UsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 15, response.Total)
	assert.Equal(t, 10, len(response.Users))
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 2, response.TotalPages)
}

func TestGetUsersWithSearch(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	testDB.Exec(`INSERT INTO users (first_name, last_name, email, date_naissance, niveau_natation) 
		VALUES ('Jean', 'Dupont', 'jean@test.com', '2010-05-15', 'NAGEUR 3'),
		       ('Marie', 'Martin', 'marie@test.com', '2012-03-20', 'PRÉSCOLAIRE 2'),
		       ('Jean', 'Bernard', 'jean.bernard@test.com', '2011-07-10', 'NAGEUR 2')`)

	r := setupRouter(testDB)

	req, _ := http.NewRequest("GET", "/api/users?search=Jean", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response UsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Total) // 2 usagers avec "Jean"
}

func TestGetUsersWithFilterNiveau(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	testDB.Exec(`INSERT INTO users (first_name, last_name, email, date_naissance, niveau_natation) 
		VALUES ('Jean', 'Dupont', 'jean@test.com', '2010-05-15', 'NAGEUR 3'),
		       ('Marie', 'Martin', 'marie@test.com', '2012-03-20', 'PRÉSCOLAIRE 2'),
		       ('Luc', 'Bernard', 'luc@test.com', '2011-07-10', 'NAGEUR 3')`)

	r := setupRouter(testDB)

	req, _ := http.NewRequest("GET", "/api/users?filter_niveau=NAGEUR 3", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response UsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response.Total) // 2 usagers avec NAGEUR 3
}

func TestGetUserByID(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Insérer un usager
	result, _ := testDB.Exec(`INSERT INTO users (first_name, last_name, email, date_naissance, niveau_natation) 
		VALUES ('Jean', 'Dupont', 'jean@test.com', '2010-05-15', 'NAGEUR 3')`)
	userID, _ := result.LastInsertId()

	r := setupRouter(testDB)

	req, _ := http.NewRequest("GET", "/api/users/"+strconv.FormatInt(userID, 10), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var user User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, "Jean", user.FirstName)
	assert.Equal(t, "Dupont", user.LastName)
}

func TestGetUserByIDNotFound(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	r := setupRouter(testDB)

	req, _ := http.NewRequest("GET", "/api/users/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateUser(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Insérer un usager
	result, _ := testDB.Exec(`INSERT INTO users (first_name, last_name, email, date_naissance, niveau_natation) 
		VALUES ('Jean', 'Dupont', 'jean@test.com', '2010-05-15', 'NAGEUR 3')`)
	userID, _ := result.LastInsertId()

	r := setupRouter(testDB)

	// Mettre à jour l'usager
	userData := UserRequest{
		FirstName:     "Jean",
		LastName:      "Martin",
		Email:         "jean.martin@test.com",
		DateNaissance: "2010-05-15",
		NiveauNatation: "NAGEUR 4",
	}

	jsonData, _ := json.Marshal(userData)
	req, _ := http.NewRequest("PUT", "/api/users/"+strconv.FormatInt(userID, 10), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var user User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, "Martin", user.LastName)
	assert.Equal(t, "jean.martin@test.com", user.Email)
	assert.Equal(t, "NAGEUR 4", user.NiveauNatation)
}

func TestDeleteUser(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Insérer un usager
	result, _ := testDB.Exec(`INSERT INTO users (first_name, last_name, email, date_naissance, niveau_natation) 
		VALUES ('Jean', 'Dupont', 'jean@test.com', '2010-05-15', 'NAGEUR 3')`)
	userID, _ := result.LastInsertId()

	r := setupRouter(testDB)

	req, _ := http.NewRequest("DELETE", "/api/users/"+strconv.FormatInt(userID, 10), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Vérifier que l'usager a été supprimé
	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&count)
	assert.Equal(t, 0, count)
}

func TestDeleteUserNotFound(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	r := setupRouter(testDB)

	req, _ := http.NewRequest("DELETE", "/api/users/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUsersWithAgeFilter(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	// Insérer des usagers avec différents âges
	testDB.Exec(`INSERT INTO users (first_name, last_name, email, date_naissance, niveau_natation) 
		VALUES ('Jean', 'Dupont', 'jean@test.com', '2010-05-15', 'NAGEUR 3'),
		       ('Marie', 'Martin', 'marie@test.com', '2015-03-20', 'PRÉSCOLAIRE 2'),
		       ('Luc', 'Bernard', 'luc@test.com', '2018-07-10', 'PARENT ET ENFANT 2')`)

	r := setupRouter(testDB)

	// Filtrer par âge minimum 5 ans
	req, _ := http.NewRequest("GET", "/api/users?filter_age_min=5", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response UsersResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	// Devrait retourner les usagers de 5 ans et plus
	assert.GreaterOrEqual(t, response.Total, 1)
}

