# Tests Unitaires

## Exécution des tests

Les tests nécessitent CGO pour fonctionner avec SQLite.:

### Dans Docker 

```bash
docker run --rm -v ${PWD}/backend:/app -w /app golang:1.21 sh -c "CGO_ENABLED=1 go test -v"
```

## Tests disponibles

Les tests couvrent :

1. **TestCalculateAge** - Test du calcul de l'âge à partir d'une date de naissance
2. **TestCreateUser** - Test de création d'un usager
3. **TestCreateUserInvalidData** - Test de validation des données
4. **TestGetUsers** - Test de récupération de la liste des usagers
5. **TestGetUsersWithPagination** - Test de la pagination
6. **TestGetUsersWithSearch** - Test de la recherche
7. **TestGetUsersWithFilterNiveau** - Test du filtrage par niveau
8. **TestGetUsersWithAgeFilter** - Test du filtrage par âge
9. **TestGetUserByID** - Test de récupération d'un usager par ID
10. **TestGetUserByIDNotFound** - Test de gestion d'erreur (usager non trouvé)
11. **TestUpdateUser** - Test de mise à jour d'un usager
12. **TestDeleteUser** - Test de suppression d'un usager
13. **TestDeleteUserNotFound** - Test de gestion d'erreur (suppression)

## Structure des tests

- Utilise une base de données SQLite en mémoire (`:memory:`) pour chaque test
- Utilise `testify/assert` pour les assertions
- Chaque test est isolé et indépendant

