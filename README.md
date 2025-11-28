# Gestion des Usagers - API REST et Application Web

Application complète permettant de gérer des usagers (CRUD) avec une API REST en Go et une interface web client-side en JavaScript.

## Technologies Utilisées

### Backend (Go)
- **Go 1.21** : Langage moderne, performant et avec une excellente gestion de la concurrence
- **Gin Framework** : Framework web léger et rapide pour Go, idéal pour les APIs REST
- **SQLite** : Base de données embarquée, parfaite pour un MVP (pas besoin de serveur de base de données séparé)
- **go-sqlite3** : Driver SQLite pour Go

**Justification du choix Go :**
- **Alignement avec la stack technique de l'entreprise** :  Unryo utilise déjà Go pour son backend, alors je voulais montrer que j’étais capable de programmer en Go.
- Performance élevée
- Compilation en binaire unique, facile à déployer
- Excellente gestion de la concurrence native
- Écosystème mature pour les APIs REST
- Pas de dépendances runtime (une fois compilé)

### Frontend (JavaScript)
- **JavaScript Vanilla** : Pas de framework lourd, code simple et maintenable pour un MVP
- **HTML5/CSS3** : Interface moderne et responsive
- **Fetch API** : Communication avec l'API REST

**Justification du choix JavaScript Vanilla :**
- Pas de dépendances externes à gérer
- Taille minimale
- Facile à comprendre et maintenir
- Parfait pour un MVP
- Compatible avec tous les navigateurs modernes

### Base de données
- **SQLite** : Base de données embarquée
  - Pas besoin de serveur séparé
- Simple à déployer
- Parfaite pour un MVP
- Facilement migrable vers PostgreSQL/MySQL si nécessaire

### Technologies considérées mais écartées

#### Backend
- **Node.js/Express** : Écarté car Go offre de meilleures performances et une meilleure gestion de la concurrence
- **Python/Flask/FastAPI** : Écarté car Go compile en binaire unique, plus facile à déployer
- **PostgreSQL/MySQL** : Écarté car SQLite est suffisant pour un MVP et simplifie le déploiement

#### Frontend
- **React/Vue/Angular** : Écartés car trop lourds pour un MVP simple. JavaScript vanilla est suffisant
- **TypeScript** : Écarté pour simplifier (mais pourrait être ajouté facilement)

## Fonctionnalités

-  Liste tous les usagers
-  Création d'un nouvel usager
-  Modification d'un usager existant
-  Suppression d'un usager
-  Interface web moderne et responsive
-  Validation des données (côté client et serveur)
-  Gestion des erreurs

## Structure du Projet

```
.
├── backend/
│   ├── main.go          # Point d'entrée et configuration des routes
│   ├── models.go        # Structures de données (User, UserRequest, etc.)
│   ├── database.go      # Gestion de la base de données
│   ├── handlers.go      # Handlers HTTP (CRUD)
│   ├── middleware.go    # Middleware (CORS)
│   ├── main_test.go     # Tests unitaires
│   ├── go.mod           # Dépendances Go
│   ├── go.sum           # Checksums des dépendances
│   └── Dockerfile       # Image Docker pour le backend
├── frontend/
│   ├── index.html       # Page principale
│   └── static/
│       ├── app.js       # Logique JavaScript
│       └── styles.css   # Styles CSS
├── docker-compose.yml   # Orchestration Docker
└── README.md            # Documentation
```

##  Installation et Déploiement

### Prérequis
- Docker et Docker Compose installés
- Aucun autre outil requis (pas de comptes payants, pas d'outils externes)

### Démarrage rapide

1. **Cloner ou télécharger le projet**

2. **Construire et démarrer les conteneurs :**
```bash
   docker-compose up --build
```

3. **Accéder à l'application :**
   - Ouvrir un navigateur à l'adresse : `http://localhost:8080`

### Commandes utiles

```bash
# Démarrer en arrière-plan
docker-compose up -d

# Voir les logs
docker-compose logs -f

# Arrêter les conteneurs
docker-compose down

# Reconstruire après modification
docker-compose up --build
```

## API REST

L'API est disponible à l'adresse `http://localhost:8080/api/users`

### Endpoints

#### GET /api/users
Liste tous les usagers avec pagination, recherche et filtres

**Paramètres de requête (optionnels) :**
- `page` : Numéro de page (défaut: 1)
- `limit` : Nombre d'usagers par page (défaut: 10, max: 100)
- `search` : Recherche dans prénom, nom ou email
- `filter_niveau` : Filtrer par niveau de natation
- `filter_age_min` : Âge minimum
- `filter_age_max` : Âge maximum

**Exemples :**
- `GET /api/users` - Première page, 10 usagers
- `GET /api/users?page=2&limit=20` - Page 2, 20 usagers par page
- `GET /api/users?search=Jean` - Recherche "Jean"
- `GET /api/users?filter_niveau=NAGEUR 3` - Filtrer par niveau
- `GET /api/users?filter_age_min=5&filter_age_max=10` - Filtrer par âge

**Réponse :**
```json
{
  "users": [
    {
      "id": 1,
      "first_name": "Jean",
      "last_name": "Dupont",
      "email": "jean.dupont@example.com",
      "date_naissance": "2010-05-15",
      "age": 14,
      "niveau_natation": "NAGEUR 3",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "limit": 10,
  "total_pages": 5
}
```

#### GET /api/users/:id
Récupère un usager par son ID

**Réponse :**
```json
{
  "id": 1,
  "first_name": "Jean",
  "last_name": "Dupont",
  "email": "jean.dupont@example.com",
  "date_naissance": "2010-05-15",
  "age": 14,
  "niveau_natation": "NAGEUR 3",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### POST /api/users
Crée un nouvel usager

**Corps de la requête :**
```json
{
  "first_name": "Jean",
  "last_name": "Dupont",
  "email": "jean.dupont@example.com",
  "date_naissance": "2010-05-15",
  "niveau_natation": "NAGEUR 3"
}
```

**Réponse :** Retourne l'usager créé avec son ID et son âge calculé

#### PUT /api/users/:id
Modifie un usager existant

**Corps de la requête :**
```json
{
  "first_name": "Jean",
  "last_name": "Martin",
  "email": "jean.martin@example.com",
  "date_naissance": "2010-05-15",
  "niveau_natation": "NAGEUR 4"
}
```

**Réponse :** Retourne l'usager mis à jour

#### DELETE /api/users/:id
Supprime un usager

**Réponse :**
```json
{
  "message": "Usager supprimé avec succès"
}
```

### Frontend

Le frontend est servi directement par le backend Go. Ouvrir `http://localhost:8080` dans un navigateur.

## Notes

- La base de données SQLite est créée automatiquement au premier démarrage
- Les données sont persistées dans le volume Docker `./backend/data`
- CORS est activé pour permettre les requêtes depuis le frontend

## Tests Unitaires

Des tests unitaires sont disponibles dans le dossier `backend/`. Pour les exécuter :

### Dans Docker (recommandé)
```bash
# Reconstruire avec l'image de build
docker run --rm -v ${PWD}/backend:/app -w /app golang:1.21 sh -c "CGO_ENABLED=1 go test -v"
```

Les tests couvrent :
- Calcul de l'âge
- Création, lecture, mise à jour, suppression d'usagers
- Pagination
- Recherche
- Filtrage par niveau et par âge
- Gestion des erreurs

Voir `backend/README_TESTS.md` pour plus de détails.

## Améliorations Possibles

- Authentification et autorisation (JWT)
- Tests d'intégration end-to-end
- Migration vers PostgreSQL pour la production
- Validation plus poussée (unicité de l'email, Date, etc.)
- Logging structuré
- Rate limiting

## Architecture et Considérations Futures

### Architecture Actuelle

```
┌─────────────────────────────────────────────────────────────┐
│                        CLIENT (Navigateur)                  │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Frontend JavaScript (Vanilla)                       │   │
│  │  - index.html                                        │   │
│  │  - app.js (Logique client-side)                      │   │
│  │  - styles.css                                        │   │
│  └──────────────────────────────────────────────────────┘   │
└───────────────────────┬─────────────────────────────────────┘
                        │ HTTP/REST API
                        │ (JSON)
┌───────────────────────▼─────────────────────────────────────┐
│                    BACKEND (Go)                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  API REST (Gin Framework)                            │   │
│  │  - GET    /api/users (avec pagination/filtres)       │   │
│  │  - GET    /api/users/:id                             │   │
│  │  - POST   /api/users                                 │   │
│  │  - PUT    /api/users/:id                             │   │
│  │  - DELETE /api/users/:id                             │   │
│  └───────────────────────┬──────────────────────────────┘   │
│                          │                                  │
│  ┌───────────────────────▼──────────────────────────────┐   │
│  │  Couche Données (SQLite)                             │   │
│  │  - users.db (fichier local)                          │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Composants et Rôles

#### Frontend (Client-Side)
- **Rôle** : Interface utilisateur, affichage et interaction
- **Technologies** : HTML5, CSS3, JavaScript Vanilla
- **Responsabilités** :
  - Affichage de la liste des usagers
  - Formulaire de création/édition
  - Gestion de la pagination côté client
  - Filtrage et recherche
  - Communication avec l'API REST via Fetch API

#### Backend (API REST)
- **Rôle** : Logique et gestion des données
- **Technologies** : Go 1.21, Gin Framework
- **Responsabilités** :
  - Validation des données
  - Gestion CRUD des usagers
  - Pagination et filtrage côté serveur
  - Calcul de l'âge à partir de la date de naissance
  - Gestion des erreurs et codes HTTP appropriés

#### Base de Données
- **Rôle** : Persistance des données
- **Technologie** : SQLite
- **Responsabilités** :
  - Stockage des informations des usagers
  - Intégrité des données (contraintes UNIQUE sur email)
  - Requêtes optimisées avec index

### Composants Non Implémentés (Considérations Futures)

#### 1. Authentification et Autorisation
- **Composant manquant** : Système d'authentification (JWT, OAuth2, Clerk)
- **Rôle prévu** : Sécuriser l'API, gérer les sessions utilisateur
- **Impact** : Actuellement, l'API est publique et accessible à tous
- **Implémentation suggérée** :
  - Middleware d'authentification JWT
  - Système de rôles (admin, utilisateur, etc.)
  - Refresh tokens pour la sécurité

#### 2. Cache
- **Composant manquant** : Système de cache (Redis, Memcached)
- **Rôle prévu** : Réduire la charge sur la base de données
- **Impact** : Chaque requête interroge directement SQLite
- **Implémentation suggérée** :
  - Cache des listes paginées
  - Cache des résultats de recherche fréquents
  - Invalidation intelligente du cache

#### 3. Logging et Monitoring
- **Composant manquant** : Système de logs structurés et monitoring
- **Rôle prévu** : Traçabilité, débogage, alertes
- **Impact** : Difficile de diagnostiquer les problèmes en production
- **Implémentation suggérée** :
  - Logging structuré (JSON) avec niveaux
  - Intégration avec des outils comme ELK Stack, Grafana
  - Métriques (Prometheus) et alertes

#### 4. Rate Limiting
- **Composant manquant** : Limitation du taux de requêtes
- **Rôle prévu** : Protection contre les abus et DDoS
- **Impact** : L'API peut être surchargée par des requêtes malveillantes
- **Implémentation suggérée** :
  - Rate limiting par IP (ex: 100 req/min)
  - Rate limiting par utilisateur authentifié
  - Utilisation de middleware comme `golang.org/x/time/rate`

#### 5. Base de Données Production
- **Composant manquant** : Base de données relationnelle (PostgreSQL/MySQL)
- **Rôle prévu** : Scalabilité, transactions, réplication
- **Impact** : SQLite ne convient pas pour la production à grande échelle
- **Implémentation suggérée** :
  - Migration vers PostgreSQL
  - Pool de connexions
  - Réplication maître-esclave pour haute disponibilité

#### 6. Tests d'Intégration
- **Composant manquant** : Tests end-to-end
- **Rôle prévu** : Validation du comportement complet de l'application
- **Impact** : Seuls les tests unitaires sont présents
- **Implémentation suggérée** :
  - Tests d'intégration avec base de données de test
  - Tests E2E avec Selenium/Playwright
  - Tests de charge (k6, Apache JMeter)

#### 7. CI/CD
- **Composant manquant** : Pipeline d'intégration et déploiement continu
- **Rôle prévu** : Automatisation des tests et déploiements
- **Impact** : Déploiement manuel, risque d'erreurs
- **Implémentation suggérée** :
  - GitHub Actions / GitLab CI
  - Tests automatiques à chaque commit
  - Déploiement automatique après validation

## Plan d'Attaque : Vulnérabilités et Points Faibles

### 1. Sécurité

#### Vulnérabilité : API Publique Sans Authentification
- **Risque** : Accès non autorisé, modification/suppression de données
- **Impact** : Critique
- **Solution** :
  - Implémenter JWT avec expiration
  - Middleware d'authentification sur toutes les routes
  - HTTPS obligatoire en production

#### Vulnérabilité : Pas de Validation Côté Serveur Avancée
- **Risque** : Injection SQL, XSS via données malformées
- **Impact** : Élevé
- **Solution** :
  - Validation stricte des entrées (format email, dates, etc.)
  - Sanitization des données
  - Utilisation de requêtes préparées (déjà fait avec SQLite)

### 2. Performance

#### Point Faible : Pas de Cache
- **Risque** : Charge élevée sur la base de données
- **Impact** : Moyen
- **Solution** :
  - Implémenter Redis pour le cache
  - Cache des requêtes fréquentes (listes paginées)

#### Point Faible : SQLite en Production
- **Risque** : Limitations de concurrence, pas de réplication
- **Impact** : Élevé pour la scalabilité
- **Solution** :
  - Migrer vers PostgreSQL
  - Pool de connexions
  - Index optimisés

#### Point Faible : Pas de Compression
- **Risque** : Bande passante inutilement utilisée
- **Impact** : Faible
- **Solution** :
  - Activer gzip compression dans Gin
  - Compression des réponses JSON

### 3. Disponibilité

#### Point Faible : Point de Défaillance Unique
- **Risque** : Si le serveur tombe, tout le service est indisponible
- **Impact** : Critique
- **Solution** :
  - Load balancer avec plusieurs instances
  - Health checks automatiques
  - Auto-scaling

#### Point Faible : Pas de Backup Automatique
- **Risque** : Perte de données en cas de problème
- **Impact** : Critique
- **Solution** :
  - Backups automatiques quotidiens
  - Réplication de base de données
  - Stockage des backups hors site


## 📄 Licence

Ce projet est un MVP de démonstration pour la compagnie Unryo.
