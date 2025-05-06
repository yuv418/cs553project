import { handleLogin } from './auth';
import { showJumpInstruction } from '../game/ui';

export function setupLoginForm() {

    const loginContainer = document.querySelector('.login-container') as HTMLElement;
    const loginForm = document.getElementById('login-form');

    if (!loginContainer) {
        console.error('Login container not found');
        return;
    }

    if (!loginForm) {
        console.error('Login form not found');
        return;
    }

    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();

        const username = (document.getElementById('username') as HTMLInputElement).value;
        const password = (document.getElementById('password') as HTMLInputElement).value;
        const errorDiv = document.getElementById('error-message');

        const response = await handleLogin(username, password);
        if (response?.jwtToken) {
            showJumpInstruction();
        } else {
            if (errorDiv) errorDiv.textContent = 'Authentication failed. Please try again.';
        }
    });

    loginContainer.style.display = 'block';
}
