export const birdSprites = {
    up: '/assets/sprites/yellowbird-upflap.png',
    mid: '/assets/sprites/yellowbird-midflap.png',
    down: '/assets/sprites/yellowbird-downflap.png'
};

let lastBirdY: number | null = null;
let bird: HTMLElement | null = null;

// Preload bird sprites
Object.values(birdSprites).map(src => {
    const img = new Image();
    img.src = src;
    return img;
});

export function updateBirdPosition(y: number) {
    if (!bird) {
        bird = document.getElementById('bird');
        if (!bird) return;
    }
    
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

export function resetBird() {
    lastBirdY = null;
}
