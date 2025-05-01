import { create } from "@bufbuild/protobuf";
import { sizeDelimitedEncode, sizeDelimitedDecodeStream } from "@bufbuild/protobuf/wire";
import * as engine from '../protos/game_engine/game_engine_pb';
import * as frameGen from '../protos/frame_gen/frame_gen_pb';
import { updateGameState } from '../game/state';
import { resetBird } from '../game/bird';
import { updateGameVisuals, hideJumpInstruction } from '../game/ui';

let gameWriter: WritableStreamDefaultWriter<any> | null = null;

export async function startGameTransport(jwt: string, gameId: string) {
    try {
        const url = import.meta.env.VITE_WEBTRANSPORT_URL + "?token=" + jwt + "&gameId=" + gameId;
        const transport = new WebTransport(url);

        await transport.ready;
        console.log('WebTransport connection established');

        const stream = await transport.createBidirectionalStream();
        gameWriter = stream.writable.getWriter();
        
        // Send initial input
        await sendGameInput(gameId);

        if (import.meta.env.VITE_DEBUG) {
            console.log('Initial input sent');
        }

        // Set up space bar event listener
        setupInputHandling(gameId);

        try {
            await handleGameStream(stream);
        } catch (e) {
            console.error('Error reading from stream:', e);
        } finally {
            cleanup(transport);
        }
    } catch (error) {
        console.error('Error with WebTransport:', error);
    }
}

async function sendGameInput(gameId: string) {
    if (!gameWriter) return;

    const inputReq = create(engine.GameEngineInputReqSchema, {
        gameId: gameId,
        key: engine.Key.SPACE
    });
    const inputBin = sizeDelimitedEncode(engine.GameEngineInputReqSchema, inputReq);
    await gameWriter.write(inputBin);
}

function setupInputHandling(gameId: string) {
    document.addEventListener('keydown', async (event) => {
        if (event.code === 'Space' && gameWriter) {
            event.preventDefault();
            hideJumpInstruction();
            await sendGameInput(gameId);
        }
    });
}

async function handleGameStream(stream: WebTransportBidirectionalStream) {
    for await (const msg of sizeDelimitedDecodeStream(frameGen.GenerateFrameReqSchema, stream.readable)) {
        if (import.meta.env.VITE_DEBUG) {
            console.log('Received game state update:', msg);
        }
        updateGameState(msg);
    }
}

function cleanup(transport: WebTransport) {
    if (gameWriter) {
        gameWriter.close();
        gameWriter = null;
    }
    
    resetBird();

    if (import.meta.env.VITE_DEBUG) {
        console.log('Stream closed');
    }

    transport.close();

    if (import.meta.env.VITE_DEBUG) {
        console.log('WebTransport session closed');
    }
}
