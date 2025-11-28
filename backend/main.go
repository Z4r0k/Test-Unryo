package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialiser la base de données
	initDB()
	defer db.Close()

	// Créer le routeur Gin
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

	// Démarrer le serveur
	log.Println("Serveur démarré sur le port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}
