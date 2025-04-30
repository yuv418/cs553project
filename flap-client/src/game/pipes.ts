let pipes: HTMLElement[] = [];
let gameContainer: HTMLElement | null = null;

export function createPipe(x: number, y: number, isUpper: boolean): HTMLElement {
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

export interface PipePosition {
    x: number;
    y: number;
}

export function updatePipes(pipePositions: PipePosition[]) {
    if (!gameContainer) {
        gameContainer = document.querySelector('.game-container');
        if (!gameContainer) return;
    }

    // Clear old pipes
    pipes.forEach(pipe => pipe.remove());
    pipes = [];

    // Create new pipes
    pipePositions.forEach(pos => {
        // Create upper pipe
        const upperPipe = createPipe(pos.x, pos.y, true);
        gameContainer!.appendChild(upperPipe);
        pipes.push(upperPipe);

        // Create lower pipe (gap of 90px between pipes)
        const lowerPipeY = gameContainer!.clientHeight - pos.y - 90 - 112; // 112 is ground height
        const lowerPipe = createPipe(pos.x, lowerPipeY, false);
        gameContainer!.appendChild(lowerPipe);
        pipes.push(lowerPipe);
    });
}
