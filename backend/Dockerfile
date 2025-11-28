# Build stage
FROM golang:1.21 AS builder

WORKDIR /app

# Installer les dépendances pour la compilation CGO
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copier les fichiers de dépendances
COPY backend/go.mod ./
COPY backend/go.sum* ./

# Copier le code source (nécessaire pour go mod tidy)
COPY backend/ .

# Télécharger les dépendances et générer go.sum si nécessaire
RUN go mod download && go mod tidy

# Construire l'application
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Runtime stage
FROM debian:bookworm-slim

# Installer SQLite et ca-certificates
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copier le binaire depuis le builder
COPY --from=builder /app/main .

# Créer un répertoire pour la base de données
RUN mkdir -p /app/data

# Copier le frontend
COPY frontend ./frontend

EXPOSE 8080

CMD ["./main"]
