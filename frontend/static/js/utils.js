// Fonctions utilitaires
// Import de date-fns depuis CDN
import { format } from 'https://cdn.jsdelivr.net/npm/date-fns@3.0.0/+esm';
import { fr } from 'https://cdn.jsdelivr.net/npm/date-fns@3.0.0/locale/fr/+esm';

/**
 * Échappe les caractères HTML pour prévenir les attaques XSS
 * @param {string} text - Texte à échapper
 * @returns {string} Texte échappé
 */
export function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

/**
 * Formate une date de manière élégante en français avec date-fns
 * @param {string} dateString - Date au format ISO
 * @returns {string} Date formatée (ex: "22 août 2025 à 13h45")
 */
export function formatDate(dateString) {
    try {
        const date = new Date(dateString);
        
        // Format: "22 août 2025 à 13h45" avec date-fns
        return format(date, "d MMMM yyyy 'à' HH'h'mm", { 
            locale: fr 
        });
    } catch (error) {
        // Fallback: formatage manuel si date-fns échoue
        console.warn('Erreur date-fns, utilisation du formatage manuel:', error);
        const date = new Date(dateString);
        const day = date.getDate().toString().padStart(2, '0');
        const month = (date.getMonth() + 1).toString().padStart(2, '0');
        const year = date.getFullYear();
        const hours = date.getHours().toString().padStart(2, '0');
        const minutes = date.getMinutes().toString().padStart(2, '0');
        
        return `${day}/${month}/${year} à ${hours}h${minutes}`;
    }
}

