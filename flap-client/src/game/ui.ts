import { AuthResponse } from '../protos/auth/auth_pb';
import { decodeToken, initiatorClient, isLoggedIn } from '../auth/auth';
import { startGameTransport, startMusicTransport } from '../network/transport';
import { setupLoginForm } from '../auth/form';
import { birdSprites, getBirdSize } from './bird';
import { scoreClient } from './score';

let gameContainer: HTMLElement | null = null;
let score = 0;

export function initializeUI() {
    gameContainer = document.querySelector('.game-container');
    hideGameOverScreen();
    if (isLoggedIn()) {
        const token = localStorage.getItem('auth_token');
        if (token) {
            showJumpInstruction();
        }
    } else {
        setupLoginForm();
    }
}

export function showJumpInstruction() {
    const loginContainer = document.querySelector('.login-container');
    const gameContainer = document.querySelector('.game-container');
    const jumpInstruction = document.getElementById('jump-instruction');

    if (loginContainer instanceof HTMLElement && gameContainer instanceof HTMLElement) {
        loginContainer.style.display = 'none';
        jumpInstruction!.style.display = 'block';
        document.addEventListener('keydown', startGameIfSpacePressed);
    }
}

export function showGameOverScreen(jwt: string) {
    // Get local score history and global score history

    scoreClient.getScores({}, { headers: {
        "Authorization": `Bearer ${jwt}`
    }}).then((resp) => {
        const gameOverScreen = document.getElementById('game-over');
        if (gameOverScreen instanceof HTMLElement) {
            gameOverScreen.style.display = 'block';
            const scoreElement = document.getElementById('final-score');
            if (scoreElement) {
                scoreElement.textContent = score.toString();
            }

            const bestScoreElement = document.getElementById("best-score");
            if (bestScoreElement) {
                bestScoreElement.textContent = resp.entries.sort((a, b) => b.score - a.score)[0].score.toString()
            }

            // Populate leaderboard
            const leaderboardElement = document.getElementById("global-leaderboard")
            if (leaderboardElement) {
                leaderboardElement.replaceChildren()

                let el = document.createElement("p")
                el.textContent = "Leaderboard"
                leaderboardElement.appendChild(el)

                for (var ent of resp.globalEntries) {
                    let el = document.createElement("p")
                    el.textContent = `${ent.username} ${ent.score}`

                    leaderboardElement.appendChild(el)
                }
            }

            const pastScoresElement = document.getElementById("past-scores")
            if (pastScoresElement) {
                pastScoresElement.replaceChildren()

                let el = document.createElement("p")
                el.textContent = "Score History"
                pastScoresElement.appendChild(el)

                for (var ent of resp.entries) {
                    let el = document.createElement("p")
                    el.textContent = `${ent.username} ${ent.score}`

                    pastScoresElement.appendChild(el)
                }
            }

            const toggleBtn = document.getElementById("toggle-stats-button")
            toggleBtn?.addEventListener("click", (_: MouseEvent) => {
                if (toggleBtn.textContent === "View Score History") {
                    console.log("Here")
                    toggleBtn.textContent = "View Leaderboard"
                    if (leaderboardElement && pastScoresElement) {
                        leaderboardElement.style.display = "none"
                        pastScoresElement.style.display = "inline-block"
                    }
                } else {
                    toggleBtn.textContent = "View Score History"
                    if (leaderboardElement && pastScoresElement) {
                        leaderboardElement.style.display = "inline-block"
                        pastScoresElement.style.display = "none"
                    }
                }
            })

            const restartButton = document.getElementById('restart-button');
            if (restartButton instanceof HTMLButtonElement) {
                restartButton.addEventListener('click', () => {
                    initializeUI();
                }, { once: true });
                document.addEventListener('keydown', (event) => {
                    if (event.code === 'Space') {
                        event.preventDefault();
                        restartButton.click();
                    }
                }, { once: true });
            }
        }
    })

}

export function hideGameOverScreen() {
    const gameOverScreen = document.getElementById('game-over');
    if (gameOverScreen instanceof HTMLElement) {
        gameOverScreen.style.display = 'none';
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

export function hideLoginContainer() {
    const loginContainer = document.querySelector('.login-container');
    if (loginContainer instanceof HTMLElement) {
        loginContainer.style.display = 'none';
    }
}

export async function startGame(response: AuthResponse) {
    if (!gameContainer) return;

    const birdSize = getBirdSize();

    try {
        const startGameResponse = await initiatorClient.startGame({
            jwt: response.jwtToken,
            viewportWidth: gameContainer.clientWidth,
            viewportHeight: gameContainer.clientHeight,
            birdHeight: birdSize.height,
            birdWidth: birdSize.width
        }, {
            headers: {
                "Authorization": `Bearer ${response.jwtToken}`
            }
        });

        if (startGameResponse.gameId) {
            hideJumpInstruction();
            hideLoginContainer();

            // This blocks!
            setTimeout(async () => { await startGameTransport(response.jwtToken, startGameResponse.gameId) });
            await startMusicTransport(response.jwtToken, startGameResponse.gameId);
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
