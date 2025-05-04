import { create } from "@bufbuild/protobuf";
import { sizeDelimitedEncode, sizeDelimitedDecodeStream } from "@bufbuild/protobuf/wire";
import * as engine from '../protos/game_engine/game_engine_pb';
import * as frameGen from '../protos/frame_gen/frame_gen_pb';
import * as musicPb from '../protos/music/music_pb';
import { updateGameState } from '../game/state';
import { resetBird } from '../game/bird';
import { hideJumpInstruction } from '../game/ui';
import { playSound } from "../game/music";

let gameWriter: WritableStreamDefaultWriter<any> | null = null;

// TODO typing
export async function startTransport(jwt: string, gameId: string, baseUrl: string, setupFn: any, cleanupFn: any, schema: any, handler: any) {
    try {
        const url = baseUrl + "?token=" + jwt + "&gameId=" + gameId;
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
        setupFn(gameId);

        try {
            await handleWTStream(jwt, stream, schema, handler);
        } catch (e) {
            console.error('Error reading from stream:', e);
        } finally {
            console.log("Closing WebTransport stream")
            cleanup(transport, cleanupFn);
        }
    } catch (error) {
        console.error('Error with WebTransport:', error);
    }
}

export async function startMusicTransport(jwt: string, gameId: string) {
    await startTransport(jwt,
                         gameId,
                         import.meta.env.VITE_WEBTRANSPORT_MUSIC_URL,
                         (_: string) => {},
                         () => {},
                         musicPb.PlayMusicRespSchema,
                         playSound)
}

export async function startGameTransport(jwt: string, gameId: string) {

    let eventListenerEvent = async (event: KeyboardEvent) => {
        if (event.code === 'Space' && gameWriter) {
            event.preventDefault();
            hideJumpInstruction();
            await sendGameInput(gameId);
        }
    }

    let setupInputHandling = (_: string) => {
        document.addEventListener('keydown', eventListenerEvent);
    };

    let cleanup = () => {
        // https://developer.mozilla.org/en-US/docs/Web/API/EventTarget/removeEventListener#matching_event_listeners_for_removal
        document.removeEventListener('keydown', eventListenerEvent)
    }

    await startTransport(jwt,
                         gameId,
                         import.meta.env.VITE_WEBTRANSPORT_GAME_URL,
                         setupInputHandling,
                         cleanup,
                         frameGen.GenerateFrameReqSchema,
                         updateGameState)
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

async function handleWTStream(jwt: string, stream: WebTransportBidirectionalStream, schema: any, msgHandler: any) {

    // @ts-ignore
    for await (const msg of sizeDelimitedDecodeStream(schema, stream.readable)) {
        if (import.meta.env.VITE_DEBUG) {
            console.log('Received game state update:', msg);
        }
        msgHandler(jwt, msg);
    }
}

function cleanup(transport: WebTransport, cleanupFn: any) {
    if (gameWriter) {
        gameWriter.close();
        gameWriter = null;
    }
    
    resetBird();

    if (import.meta.env.VITE_DEBUG) {
        console.log('Stream closed');
    }

    if (!transport.closed) {
        transport.close();
    }

    cleanupFn()

    if (import.meta.env.VITE_DEBUG) {
        console.log('WebTransport session closed');
    }
}
