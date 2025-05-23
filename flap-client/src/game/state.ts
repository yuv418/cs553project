import type { GenerateFrameReq } from '../protos/frame_gen/frame_gen_pb';
import { updateBirdPosition } from './bird';
import { updatePipes } from './pipes';
import { showGameOverScreen, updateScore } from './ui';
import { logReceiveTime } from '../latencyLogger';

export function updateGameState(jwt: string, frame: GenerateFrameReq) {
    window.firstFrameReceived = true

    if (import.meta.env.VITE_LOG_LATENCY) {
        logReceiveTime('frame');
    }
    // Update game state
    if (frame.gameOver) {
        showGameOverScreen(jwt);
        return;
    }

    // Update bird position
    if (frame.birdPosition) {
        updateBirdPosition(frame.birdPosition.y);
    }

    // Update score
    if (frame.score !== undefined) {
        updateScore(frame.score);
    }

    // Update pipes
    if (frame.pipePositions) {
        updatePipes(frame.pipePositions, frame.pipeStarts, frame.pipeGaps, frame.pipeWidth);
    }
}
