let pipes: HTMLElement[] = [];
let gameContainer: HTMLElement | null = null;

export function createPipe(x: number, y: number, width: number, isUpper: boolean): HTMLElement {
    const pipe = document.createElement('div');
    pipe.className = 'pipe';
    pipe.style.width = `${width}px`;
    pipe.style.left = `${x}px`;

    const pipeBody = document.createElement('div');
    pipeBody.className = isUpper ? 'pipe-upper' : 'pipe-lower';


    if (isUpper) {
        pipeBody.style.bottom = '-10px';
        pipeBody.style.height = `${y}px`;
    } else {
        pipeBody.style.top = '-10px';
        pipeBody.style.height = `${y}px`;
    }

    pipe.appendChild(pipeBody);
    return pipe;
}

export function updatePipes(pipePositions: number[], pipeStarts: number[], pipeGaps: number[], pipeWidth: number) {
    if (!gameContainer) {
        gameContainer = document.querySelector('.game-container');
        if (!gameContainer) return;
    }

    if (import.meta.env.VITE_DEBUG) {
        console.log('Pipe positions:', pipePositions);
        console.log('Pipe starts:', pipeStarts);
        console.log('Pipe gaps:', pipeGaps);
        console.log('Pipe width:', pipeWidth);
    }

    // Clear old pipes
    pipes.forEach(pipe => pipe.remove());
    pipes = [];

    // Create new pipes
    for (let i = 0; i < pipePositions.length; i++) {
        const pos = pipePositions[i];
        const start = pipeStarts[i];
        const gap = pipeGaps[i];

        // Create upper pipe
        const upperPipe = createPipe(pos, start, pipeWidth, true);
        gameContainer!.appendChild(upperPipe);
        pipes.push(upperPipe);

        // Create lower pipe (gap of 90px between pipes)
        const lowerPipeY = gameContainer!.clientHeight - start - gap - 112; // 112 is ground height
        const lowerPipe = createPipe(pos, lowerPipeY, pipeWidth, false);
        gameContainer!.appendChild(lowerPipe);
        pipes.push(lowerPipe);
    }
}
