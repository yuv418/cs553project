import type { GenerateFrameReq } from '../protos/frame_gen/frame_gen_pb';
import { updateBirdPosition } from './bird';
import { updatePipes } from './pipes';
import { updateScore } from './ui';

export function updateGameState(frame: GenerateFrameReq) {
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
        updatePipes(frame.pipePositions);
    }
}
