// Application principale - Gestion des usagers

// Import de date-fns depuis CDN
import { format } from 'https://cdn.jsdelivr.net/npm/date-fns@3.0.0/+esm';
import { fr } from 'https://cdn.jsdelivr.net/npm/date-fns@3.0.0/locale/fr/+esm';

import { API_BASE_URL, DEFAULT_PAGE, DEFAULT_LIMIT, SEARCH_DEBOUNCE_MS, MESSAGE_DISPLAY_DURATION_MS } from './config.js';
import { escapeHtml } from './utils.js';

// Éléments DOM
const usersList = document.getElementById('usersList');
const userForm = document.getElementById('userForm');
const userFormElement = document.getElementById('userFormElement');
const addUserBtn = document.getElementById('addUserBtn');
const cancelBtn = document.getElementById('cancelBtn');
const formTitle = document.getElementById('formTitle');
const messageDiv = document.getElementById('message');
const searchInput = document.getElementById('searchInput');
const paginationDiv = document.getElementById('pagination');
const filterNiveau = document.getElementById('filterNiveau');
const filterAgeMin = document.getElementById('filterAgeMin');
const filterAgeMax = document.getElementById('filterAgeMax');
const clearFiltersBtn = document.getElementById('clearFilters');

// État de l'application
let editingUserId = null;
let currentPage = DEFAULT_PAGE;
let currentLimit = DEFAULT_LIMIT;
let currentSearch = '';
let currentFilterNiveau = '';
let currentFilterAgeMin = '';
let currentFilterAgeMax = '';
let totalPages = 1;
let searchTimeout = null;

// Initialisation au chargement de la page
document.addEventListener('DOMContentLoaded', () => {
    loadUsers();
    
    addUserBtn.addEventListener('click', () => {
        showForm();
    });
    
    cancelBtn.addEventListener('click', () => {
        hideForm();
    });
    
    userFormElement.addEventListener('submit', handleFormSubmit);
    
    // Recherche avec debounce
    searchInput.addEventListener('input', (e) => {
        clearTimeout(searchTimeout);
        currentSearch = e.target.value.trim();
        currentPage = DEFAULT_PAGE;
        searchTimeout = setTimeout(() => {
            loadUsers();
        }, SEARCH_DEBOUNCE_MS);
    });
    
    // Filtres
    filterNiveau.addEventListener('change', (e) => {
        currentFilterNiveau = e.target.value;
        currentPage = DEFAULT_PAGE;
        loadUsers();
    });
    
    filterAgeMin.addEventListener('input', (e) => {
        clearTimeout(searchTimeout);
        currentFilterAgeMin = e.target.value;
        currentPage = DEFAULT_PAGE;
        searchTimeout = setTimeout(() => {
            loadUsers();
        }, SEARCH_DEBOUNCE_MS);
    });
    
    filterAgeMax.addEventListener('input', (e) => {
        clearTimeout(searchTimeout);
        currentFilterAgeMax = e.target.value;
        currentPage = DEFAULT_PAGE;
        searchTimeout = setTimeout(() => {
            loadUsers();
        }, SEARCH_DEBOUNCE_MS);
    });
    
    clearFiltersBtn.addEventListener('click', () => {
        currentSearch = '';
        currentFilterNiveau = '';
        currentFilterAgeMin = '';
        currentFilterAgeMax = '';
        searchInput.value = '';
        filterNiveau.value = '';
        filterAgeMin.value = '';
        filterAgeMax.value = '';
        currentPage = DEFAULT_PAGE;
        loadUsers();
    });
});

// Charger les usagers avec pagination et recherche
async function loadUsers() {
    try {
        // Construire l'URL avec les paramètres
        const params = new URLSearchParams();
        params.append('page', currentPage.toString());
        params.append('limit', currentLimit.toString());
        if (currentSearch) {
            params.append('search', currentSearch);
        }
        if (currentFilterNiveau) {
            params.append('filter_niveau', currentFilterNiveau);
        }
        if (currentFilterAgeMin) {
            params.append('filter_age_min', currentFilterAgeMin);
        }
        if (currentFilterAgeMax) {
            params.append('filter_age_max', currentFilterAgeMax);
        }
        
        const url = `${API_BASE_URL}?${params.toString()}`;
        const response = await fetch(url);
        const contentType = response.headers.get("content-type");
        
        if (!response.ok) {
            let errorMessage = 'Erreur lors du chargement';
            if (contentType && contentType.includes("application/json")) {
                try {
                    const errorData = await response.json();
                    errorMessage = errorData.error || errorMessage;
                } catch (e) {
                    // Ignore JSON parse error
                }
            } else {
                errorMessage = `Erreur ${response.status}: ${response.statusText}`;
            }
            throw new Error(errorMessage);
        }
        
        if (!contentType || !contentType.includes("application/json")) {
            throw new Error('Réponse non-JSON reçue du serveur');
        }
        
        const data = await response.json();
        currentPage = data.page;
        totalPages = data.total_pages;
        displayUsers(data.users);
        displayPagination(data);
    } catch (error) {
        showMessage('Erreur lors du chargement des usagers: ' + error.message, 'error');
        usersList.innerHTML = '<p class="error">Erreur lors du chargement des usagers</p>';
        paginationDiv.style.display = 'none';
    }
}

// Afficher les usagers
function displayUsers(users) {
    // Vérifier que users est un tableau
    if (users === null || users === undefined) {
        users = [];
    }
    
    if (!Array.isArray(users)) {
        usersList.innerHTML = '<p class="error">Format de données invalide</p>';
        paginationDiv.style.display = 'none';
        return;
    }
    
    if (users.length === 0) {
        if (currentSearch) {
            usersList.innerHTML = '<p class="empty">Aucun usager trouvé pour votre recherche</p>';
        } else {
            usersList.innerHTML = '<p class="empty">Aucun usager enregistré</p>';
        }
        paginationDiv.style.display = 'none';
        return;
    }
    
    usersList.innerHTML = users.map(user => `
        <div class="user-card">
            <div class="user-info">
                <h3>${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}</h3>
                <p class="email">${escapeHtml(user.email)}</p>
                <p class="age">Âge: ${user.age || 'N/A'} ans</p>
                <p class="niveau">Niveau: ${escapeHtml(user.niveau_natation || 'N/A')}</p>
                <p class="date">Créé le ${format(new Date(user.created_at), "d MMMM yyyy 'à' HH'h'mm", { locale: fr })}</p>
            </div>
            <div class="user-actions">
                <button class="btn btn-edit" onclick="editUser(${user.id})" aria-label="Modifier ${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}">Modifier</button>
                <button class="btn btn-delete" onclick="deleteUser(${user.id})" aria-label="Supprimer ${escapeHtml(user.first_name)} ${escapeHtml(user.last_name)}">Supprimer</button>
            </div>
        </div>
    `).join('');
}

// Afficher la pagination
function displayPagination(data) {
    if (data.total_pages <= 1) {
        paginationDiv.style.display = 'none';
        return;
    }
    
    paginationDiv.style.display = 'flex';
    
    const start = Math.max(1, data.page - 2);
    const end = Math.min(data.total_pages, data.page + 2);
    
    let html = '<div class="pagination-info">';
    const startItem = (data.page - 1) * data.limit + 1;
    const endItem = Math.min(data.page * data.limit, data.total);
    html += `Affichage de ${startItem} à ${endItem} sur ${data.total} usager${data.total > 1 ? 's' : ''}`;
    html += '</div>';
    
    html += '<div class="pagination-controls">';
    
    // Bouton précédent
    if (data.page > 1) {
        html += `<button class="btn-pagination" onclick="goToPage(${data.page - 1})">Précédent</button>`;
    } else {
        html += `<button class="btn-pagination" disabled>Précédent</button>`;
    }
    
    // Première page
    if (start > 1) {
        html += `<button class="btn-pagination ${data.page === 1 ? 'active' : ''}" onclick="goToPage(1)">1</button>`;
        if (start > 2) {
            html += '<span class="pagination-ellipsis">...</span>';
        }
    }
    
    // Pages autour de la page actuelle
    for (let i = start; i <= end; i++) {
        html += `<button class="btn-pagination ${data.page === i ? 'active' : ''}" onclick="goToPage(${i})">${i}</button>`;
    }
    
    // Dernière page
    if (end < data.total_pages) {
        if (end < data.total_pages - 1) {
            html += '<span class="pagination-ellipsis">...</span>';
        }
        html += `<button class="btn-pagination ${data.page === data.total_pages ? 'active' : ''}" onclick="goToPage(${data.total_pages})">${data.total_pages}</button>`;
    }
    
    // Bouton suivant
    if (data.page < data.total_pages) {
        html += `<button class="btn-pagination" onclick="goToPage(${data.page + 1})">Suivant</button>`;
    } else {
        html += `<button class="btn-pagination" disabled>Suivant</button>`;
    }
    
    html += '</div>';
    
    paginationDiv.innerHTML = html;
}

// Aller à une page spécifique
window.goToPage = function(page) {
    if (page >= 1 && page <= totalPages && page !== currentPage) {
        currentPage = page;
        loadUsers();
        // Scroll vers le haut de la liste
        usersList.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
};

// Afficher le formulaire
function showForm(user = null) {
    editingUserId = user ? user.id : null;
    formTitle.textContent = user ? 'Modifier un usager' : 'Ajouter un usager';
    
    if (user) {
        document.getElementById('userId').value = user.id;
        document.getElementById('firstName').value = user.first_name;
        document.getElementById('lastName').value = user.last_name;
        document.getElementById('email').value = user.email;
        document.getElementById('dateNaissance').value = user.date_naissance || '';
        document.getElementById('niveauNatation').value = user.niveau_natation || '';
    } else {
        userFormElement.reset();
    }
    
    userForm.style.display = 'block';
    userForm.scrollIntoView({ behavior: 'smooth' });
}

// Masquer le formulaire
function hideForm() {
    userForm.style.display = 'none';
    userFormElement.reset();
    editingUserId = null;
}

// Gérer la soumission du formulaire
async function handleFormSubmit(e) {
    e.preventDefault();
    
    const userData = {
        first_name: document.getElementById('firstName').value.trim(),
        last_name: document.getElementById('lastName').value.trim(),
        email: document.getElementById('email').value.trim(),
        date_naissance: document.getElementById('dateNaissance').value,
        niveau_natation: document.getElementById('niveauNatation').value
    };
    
    try {
        const url = editingUserId 
            ? `${API_BASE_URL}/${editingUserId}`
            : API_BASE_URL;
        
        const method = editingUserId ? 'PUT' : 'POST';
        
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(userData)
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Erreur lors de l\'enregistrement');
        }
        
        showMessage(
            editingUserId 
                ? 'Usager modifié avec succès' 
                : 'Usager créé avec succès',
            'success'
        );
        
        hideForm();
        currentPage = DEFAULT_PAGE; // Reset à la première page après création/modification
        loadUsers();
    } catch (error) {
        showMessage('Erreur: ' + error.message, 'error');
    }
}

// Modifier un usager
window.editUser = async function(id) {
    try {
        const response = await fetch(`${API_BASE_URL}/${id}`);
        if (!response.ok) throw new Error('Erreur lors du chargement');
        
        const user = await response.json();
        showForm(user);
    } catch (error) {
        showMessage('Erreur lors du chargement de l\'usager: ' + error.message, 'error');
    }
};

// Supprimer un usager
window.deleteUser = async function(id) {
    if (!confirm('Êtes-vous sûr de vouloir supprimer cet usager ?')) {
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}/${id}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Erreur lors de la suppression');
        }
        
        showMessage('Usager supprimé avec succès', 'success');
        currentPage = DEFAULT_PAGE; // Reset à la première page après suppression
        loadUsers();
    } catch (error) {
        showMessage('Erreur: ' + error.message, 'error');
    }
};

// Afficher un message
function showMessage(text, type) {
    messageDiv.textContent = text;
    messageDiv.className = `message ${type}`;
    messageDiv.style.display = 'block';
    
    setTimeout(() => {
        messageDiv.style.display = 'none';
    }, MESSAGE_DISPLAY_DURATION_MS);
}

