import http from 'k6/http';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// Initialisation du générateur aléatoire

export const options = {
    stages: [
        { duration: '1m', target: 100 },  // 10 utilisateurs pendant 1 minute
    ],
};

export default function () {
    // Génération aléatoire d'une adresse e-mail
    const randomEmail = `user_${uuidv4()}@example.com`;

    // Corps de la requête
    const payload = JSON.stringify({
        first_name: 'Jean',
        last_name: 'Test',
        email: randomEmail,
        password: 'Testtest99@',
        role: 'client',
    });

    // En-têtes de la requête
    const headers = {
        'Content-Type': 'application/json',
    };

    // Requête POST
    const response = http.post('http://dev.goodfood.ovh/auth/register', payload, { headers });

    // Vérifications
    check(response, {
        'Status is 201': (r) => r.status === 201,
        'Response contains id': (r) => JSON.parse(r.body).id !== undefined,
    });
}
