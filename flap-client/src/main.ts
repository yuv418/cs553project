import './style.css';
import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthService } from './protos/auth/auth_pb';
import * as engine from './protos/game_engine/game_engine_pb.js';
import * as frameGen from './protos/frame_gen/frame_gen_pb.js';
import { InitiatorService, StartGameReqSchema } from './protos/initiator/initiator_pb';
import { createClient } from "@connectrpc/connect";
import { create, Message } from "@bufbuild/protobuf"
import { sizeDelimitedEncode, sizeDelimitedDecodeStream } from "@bufbuild/protobuf/wire"
import { jwtDecode } from "jwt-decode";
import * as wkt from "@bufbuild/protobuf/wkt";

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
            const loginContainer = document.querySelector('.login-container');
            const playButton = document.getElementById('playButton');
            if (loginContainer instanceof HTMLElement && playButton instanceof HTMLElement) {
                loginContainer.innerHTML = `
                    <div class="welcome-message">
                        <h2>Welcome, ${username}!</h2>
                        <p>Click PLAY to start!</p>
                    </div>
                `;
                loginContainer.style.background = 'transparent';
                playButton.style.display = 'block';

                // Add click handler for the play button
                playButton.addEventListener('click', async () => {
                    const gameContainer = document.querySelector('.game-container');
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
                            const jumpInstruction = document.getElementById('jumpInstruction');
                            if (jumpInstruction instanceof HTMLElement) {
                                jumpInstruction.style.display = 'block';
                            }
                            playButton.style.display = 'none';
                            await connectToWebTransport(response.jwtToken, startGameResponse.gameId);
                        }
                    } catch (error) {
                        console.error('Failed to start game:', error);
                    }
                });
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
        })

        const inputBin = sizeDelimitedEncode(engine.GameEngineInputReqSchema, inputReq)

        await gameWriter.write(inputBin); console.log(`Sent: ${JSON.stringify(inputReq)}`); const gameContainer = document.querySelector('.game-container');
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
                    }, 100); // Remove after 100ms
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
                console.log('Space key pressed - sent input to server');
            }
        });

        try {
            for await (const msg of sizeDelimitedDecodeStream(frameGen.GenerateFrameReqSchema, stream.readable)) {
                console.log('Received game state update:', msg);

                // Show feedback message
                if (gameFeedback instanceof HTMLElement) {
                    gameFeedback.style.display = 'block';
                    gameFeedback.textContent = 'Game Update Received!';
                    setTimeout(() => {
                        gameFeedback.style.display = 'none';
                    }, 500); // Hide after 500ms
                }
            }
        } catch (e) {
            console.error('Error reading from stream:', e);
        } finally {
            if (gameWriter) {
                gameWriter.close();
                gameWriter = null;
            }
            console.log('Stream closed');

            // Close the WebTransport session
            transport.close();
            console.log('WebTransport session closed');
        }
    } catch (error) {
        console.error('Error with WebTransport:', error);
    }
};
