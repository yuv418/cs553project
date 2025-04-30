import './style.css';
import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthResponse, AuthService } from './protos/auth/auth_pb';
import { InitiatorService } from './protos/initiator/initiator_pb';
import { createClient } from "@connectrpc/connect";
import { create } from "@bufbuild/protobuf"
import { sizeDelimitedEncode, sizeDelimitedDecodeStream } from "@bufbuild/protobuf/wire"
import { jwtDecode } from "jwt-decode";

import * as engine from './protos/game_engine/game_engine_pb.js';
import * as frameGen from './protos/frame_gen/frame_gen_pb.js';
import { env } from 'process';

interface JWTPayload {
    sub?: string;
    username?: string;
    exp: number;
}

const authTransport = createConnectTransport({
    baseUrl: import.meta.env.VITE_AUTH_SERVICE_URL,
    useBinaryFormat: true,
});

const initiatorTransport = createConnectTransport({
    baseUrl: import.meta.env.VITE_INITIATOR_SERVICE_URL,
    useBinaryFormat: true,
});

const authClient = createClient(AuthService, authTransport);
const initiatorClient = createClient(InitiatorService, initiatorTransport);

document.getElementById('loginForm')?.addEventListener('submit', async (e) => {
    e.preventDefault();

    const username = (document.getElementById('username') as HTMLInputElement).value;
    const password = (document.getElementById('password') as HTMLInputElement).value;
    const errorDiv = document.getElementById('error-message');

    try {
        const response = await authClient.authenticate({
            username,
            password
        });
        if (response.jwtToken) {
            // Store the JWT token
            localStorage.setItem('auth_token', response.jwtToken);

            // Decode the JWT token
            const decoded = jwtDecode<JWTPayload>(response.jwtToken);
            const username = decoded.sub || decoded.username;
            const gameContainer = document.querySelector('.game-container');
            const loginContainer = document.querySelector('.login-container');
            if (loginContainer instanceof HTMLElement && gameContainer instanceof HTMLElement) {
                loginContainer.innerHTML = `
                    <div class="welcome-message">
                        <h2>Back for more, ${username}?</h2>
                        <p>Tap to start</p>
                    </div>
                `;
                loginContainer.style.background = 'transparent';

                gameContainer.addEventListener('click', startGame.bind(null, response), { once: true });
            }
        } else {
            if (errorDiv) errorDiv.textContent = 'Authentication failed';
        }
    } catch (error) {
        console.error('Authentication error:', error);
        if (errorDiv) errorDiv.textContent = 'Authentication failed. Please try again.';
    }
});

let gameWriter: WritableStreamDefaultWriter<any> | null = null;

// Game state
let bird: HTMLElement | null = null;
let pipes: HTMLElement[] = [];
let gameContainer: HTMLElement | null = null;
let score = 0;
let lastBirdY: number | null = null;

const birdSprites = {
    up: '/assets/sprites/yellowbird-upflap.png',
    mid: '/assets/sprites/yellowbird-midflap.png',
    down: '/assets/sprites/yellowbird-downflap.png'
};

// Preload bird sprites (we love performance)
Object.values(birdSprites).map(src => {
    const img = new Image();
    img.src = src;
    return img;
});

let animationFrameId: number;

async function startGame(response: AuthResponse) {
    const loginContainer = document.querySelector('.login-container') as HTMLElement;
    const gameContainer = document.querySelector('.game-container') as HTMLElement;

    if (!gameContainer || !loginContainer) return;

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
            const jumpInstruction = document.getElementById('jumpInstruction');
            if (jumpInstruction instanceof HTMLElement) {
                jumpInstruction.style.display = 'block';
            }
            loginContainer.style.display = 'none';
            await connectToWebTransport(response.jwtToken, startGameResponse.gameId);
        }
    } catch (error) {
        console.error('Failed to start game:', error);
    }
}

function animateBird() {
    if (!bird) return;
    const birdY = parseFloat(bird.style.top) || 0;
    if (lastBirdY !== null) {
        if (birdY < lastBirdY) {
            // Bird is moving up
            bird.style.backgroundImage = `url(${birdSprites.up})`;
        } else if (birdY > lastBirdY) {
            // Bird is moving down
            bird.style.backgroundImage = `url(${birdSprites.down})`;
        }
    }
    lastBirdY = birdY;
    animationFrameId = requestAnimationFrame(animateBird);
}

function updateBirdPosition(y: number) {
    if (!bird) return;
    
    // Determine vertical movement
    let spriteToUse = birdSprites.mid;
    if (lastBirdY !== null) {
        if (y < lastBirdY - 2) { // Moving up
            spriteToUse = birdSprites.up;
        } else if (y > lastBirdY + 2) { // Moving down
            spriteToUse = birdSprites.down;
        }
    }
    
    // Update position and sprite
    bird.style.top = `${y}px`;
    bird.style.backgroundImage = `url(${spriteToUse})`;
    
    // Add a slight rotation based on vertical movement
    const rotation = lastBirdY !== null ? Math.min(Math.max(-20, (y - lastBirdY) * 2), 20) : 0;
    bird.style.transform = `rotate(${rotation}deg)`;
    
    lastBirdY = y;
}

function createPipe(x: number, y: number, isUpper: boolean): HTMLElement {
    const pipe = document.createElement('div');
    pipe.className = 'pipe';
    pipe.style.left = `${x}px`;

    const pipeBody = document.createElement('div');
    pipeBody.className = isUpper ? 'pipe-upper' : 'pipe-lower';

    if (isUpper) {
        pipeBody.style.bottom = '0';
        pipeBody.style.height = `${y}px`;
    } else {
        pipeBody.style.top = '0';
        pipeBody.style.height = `${y}px`;
    }

    pipe.appendChild(pipeBody);
    return pipe;
}

interface PipePosition {
    x: number;
    y: number;
}

function updateGameState(frame: frameGen.GenerateFrameReq) {
    if (!gameContainer) {
        gameContainer = document.querySelector('.game-container');
        bird = document.getElementById('bird');
    }

    if (!gameContainer || !bird) return;

    // Update bird position
    if (frame.birdPosition) {
        updateBirdPosition(frame.birdPosition.y);
    }

    // Update score
    if (frame.score !== undefined && frame.score !== score) {
        score = frame.score;
        const scoreElement = document.getElementById('score');
        if (scoreElement) {
            scoreElement.textContent = score.toString();
        }
    }

    // Clear old pipes
    pipes.forEach(pipe => pipe.remove());
    pipes = [];

    // Create new pipes
    if (frame.pipePositions) {
        frame.pipePositions.forEach((pos: PipePosition) => {
            if (!gameContainer) return;
            // Create upper pipe
            const upperPipe = createPipe(pos.x, pos.y, true);
            gameContainer.appendChild(upperPipe);
            pipes.push(upperPipe);

            // Create lower pipe (gap of 90px between pipes)
            const lowerPipeY = gameContainer.clientHeight - pos.y - 90 - 112; // 112 is ground height
            const lowerPipe = createPipe(pos.x, lowerPipeY, false);
            gameContainer.appendChild(lowerPipe);
            pipes.push(lowerPipe);
        });
    }
}

export const connectToWebTransport = async (jwt: string, gameId: string) => {
    try {
        const url = import.meta.env.VITE_WEBTRANSPORT_URL + "?token=" + jwt;
        const transport = new WebTransport(url);

        await transport.ready;
        console.log('WebTransport connection established');

        const stream = await transport.createBidirectionalStream();
        gameWriter = stream.writable.getWriter();
        const inputReq = create(engine.GameEngineInputReqSchema, {
            gameId: gameId,
            key: engine.Key.SPACE
        });

        const inputBin = sizeDelimitedEncode(engine.GameEngineInputReqSchema, inputReq);

        await gameWriter.write(inputBin);

        if (import.meta.env.VITE_DEBUG) {
            console.log('Initial input sent:', inputReq);
        }

        const gameContainer = document.querySelector('.game-container');
        const gameFeedback = document.getElementById('gameFeedback');

        // Add space bar event listener
        document.addEventListener('keydown', async (event) => {
            if (event.code === 'Space' && gameWriter) {
                event.preventDefault(); // Prevent page scrolling

                // Visual feedback for space press
                if (gameContainer instanceof HTMLElement) {
                    gameContainer.classList.add('game-active');
                    setTimeout(() => {
                        gameContainer.classList.remove('game-active');
                    }, 100);
                }

                // Hide the jump instruction after first jump
                const jumpInstruction = document.getElementById('jumpInstruction');
                if (jumpInstruction instanceof HTMLElement) {
                    jumpInstruction.style.display = 'none';
                }

                const inputReq = create(engine.GameEngineInputReqSchema, {
                    gameId: gameId,
                    key: engine.Key.SPACE
                });
                const inputBin = sizeDelimitedEncode(engine.GameEngineInputReqSchema, inputReq);
                await gameWriter.write(inputBin);
            }
        });

        try {
            for await (const msg of sizeDelimitedDecodeStream(frameGen.GenerateFrameReqSchema, stream.readable)) {
                if (import.meta.env.VITE_DEBUG) {
                    console.log('Received game state update:', msg);
                }
                updateGameState(msg);
            }
        } catch (e) {
            console.error('Error reading from stream:', e);
        } finally {
            if (gameWriter) {
                gameWriter.close();
                gameWriter = null;
            }            // Reset bird state
            lastBirdY = null;

            if (import.meta.env.VITE_DEBUG) {
                console.log('Stream closed');
            }

            // Close the WebTransport session
            transport.close();

            if (import.meta.env.VITE_DEBUG) {
                console.log('WebTransport session closed');
            }
        }
    } catch (error) {
        console.error('Error with WebTransport:', error);
    }
};
