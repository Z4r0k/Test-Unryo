package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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

// initDB initialise la connexion à la base de données et crée les tables si nécessaire
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

