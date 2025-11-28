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


