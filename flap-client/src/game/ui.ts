import { AuthResponse } from '../protos/auth/auth_pb';
import { decodeToken, initiatorClient, isLoggedIn } from '../auth/auth';
import { startGameTransport } from '../network/transport';
import { setupLoginForm } from '../auth/form';

let gameContainer: HTMLElement | null = null;
let score = 0;

export function initializeUI() {
    gameContainer = document.querySelector('.game-container');
    if (isLoggedIn()) {
        const token = localStorage.getItem('auth_token');
        if (token) {
            const response = { jwtToken: token } as AuthResponse;
            showJumpInstruction(response);
        }
    } else {
        setupLoginForm();
    }
}

export function showJumpInstruction(response: AuthResponse) {
    const loginContainer = document.querySelector('.login-container');
    const gameContainer = document.querySelector('.game-container');
    const jumpInstruction = document.getElementById('jump-instruction');

    if (loginContainer instanceof HTMLElement && gameContainer instanceof HTMLElement) {
        loginContainer.style.display = 'none';
        jumpInstruction!.style.display = 'block';
        document.addEventListener('keydown', startGameIfSpacePressed);
    }
}

export function updateScore(newScore: number) {
    if (newScore !== score) {
        score = newScore;
        const scoreElement = document.getElementById('score');
        if (scoreElement) {
            scoreElement.textContent = score.toString();
        }
    }
}

export function hideJumpInstruction() {
    const jumpInstruction = document.getElementById('jump-instruction');
    if (jumpInstruction instanceof HTMLElement) {
        jumpInstruction.style.display = 'none';
    }
}

export function updateGameVisuals(isActive: boolean) {
    if (gameContainer instanceof HTMLElement) {
        gameContainer.classList.toggle('game-active', isActive);
    }
}

export function hideLoginContainer() {
    const loginContainer = document.querySelector('.login-container');
    if (loginContainer instanceof HTMLElement) {
        loginContainer.style.display = 'none';
    }
}

export async function startGame(response: AuthResponse) {
    if (!gameContainer) return;

    try {
        const startGameResponse = await initiatorClient.startGame({
            jwt: response.jwtToken,
            viewportWidth: gameContainer.clientWidth,
            viewportHeight: gameContainer.clientHeight
        }, {
            headers: {
                "Authorization": `Bearer ${response.jwtToken}`
            }
        });

        if (startGameResponse.gameId) {
            hideJumpInstruction();
            hideLoginContainer();
            await startGameTransport(response.jwtToken, startGameResponse.gameId);
        }
    } catch (error) {
        console.error('Failed to start game:', error);
    }
}

function startGameIfSpacePressed(event: KeyboardEvent) {
    if (event.code === 'Space') {
        event.preventDefault();
        const token = localStorage.getItem('auth_token');
        if (token) {
            const response = { jwtToken: token } as AuthResponse;
            document.removeEventListener('keydown', startGameIfSpacePressed);
            startGame(response);
        }
    }
}