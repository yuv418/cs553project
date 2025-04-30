package main

import (
    "fmt"
    "log"
    "math/rand"
    "time"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Game struct: holds the game state and assets.
type Game struct {
    birdY         float64       // Bird's vertical position (Y-coordinate, pixels).
    gravity       float64       // Gravity force per frame (pixels).
    flapForce     float64       // Upward force when flapping (negative).
    pipes         []Pipe        // Slice of pipes for obstacles.
    frame         int           // Frame counter for timing (e.g., pipe spawning).
    score         int           // Player’s score (increments when passing pipes).
    state         string        // Game state: "playing" or "gameOver".
    groundX       float64       // Ground’s horizontal offset for scrolling (pixels).
    birdImg       *ebiten.Image // Bird image (e.g., bird-0.png).
    pipeImg       *ebiten.Image // Pipe image (e.g., pipe-green.png).
    groundImg     *ebiten.Image // Ground image (e.g., base.png).
    backgroundImg *ebiten.Image // Background image (e.g., background-day.png).
}

// Pipe struct: represents a pipe obstacle (top and bottom pair).
type Pipe struct {
    x      float64 // Pipe’s horizontal position (X-coordinate, pixels).
    gapY   float64 // Vertical center of the gap between pipes (pixels).
    scored bool    // Tracks if the pipe has been scored.
}

// NewGame: initializes a new game instance with default values and loaded images.
func NewGame() *Game {
    // Seed random number generator for pipe gaps and random flaps.
    rand.Seed(time.Now().UnixNano())

    // Load images.
    birdImg, _, err := ebitenutil.NewImageFromFile("bird.png")
    if err != nil {
        log.Fatal(err)
    }
    pipeImg, _, err := ebitenutil.NewImageFromFile("pipe.png")
    if err != nil {
        log.Fatal(err)
    }
    groundImg, _, err := ebitenutil.NewImageFromFile("ground.png")
    if err != nil {
        log.Fatal(err)
    }
    backgroundImg, _, err := ebitenutil.NewImageFromFile("background.png")
    if err != nil {
        log.Fatal(err)
    }

    // Initialize game state.
    return &Game{
        birdY:         200.0,
        gravity:       0.5,
        flapForce:     -10.0,
        pipes:         []Pipe{},
        frame:         0,
        score:         0,
        state:         "playing",
        groundX:       0.0,
        birdImg:       birdImg,
        pipeImg:       pipeImg,
        groundImg:     groundImg,
        backgroundImg: backgroundImg,
    }
}

// Update: runs each frame to update game logic (60 FPS by default).
func (g *Game) Update() error {
    // Handle game-over state: restart on R key press.
    if g.state == "gameOver" {
        if ebiten.IsKeyPressed(ebiten.KeyR) {
            *g = *NewGame()
        }
        return nil
    }

    // Increment frame counter.
    g.frame++

    // Random flap: 5% chance per frame (adjustable).
    if rand.Float64() < 0.05 {
        g.birdY += g.flapForce
    }
    // Apply gravity.
    g.birdY += g.gravity

    // Spawn a new pipe every 90 frames (~1.5 seconds at 60 FPS).
    if g.frame%90 == 0 {
        gapY := 150.0 + rand.Float64()*100.0 // Gap center between 150–250.
        g.pipes = append(g.pipes, Pipe{x: 400.0, gapY: gapY, scored: false})
    }

    // Move pipes left and check for scoring.
    for i := range g.pipes {
        g.pipes[i].x -= 2.0 // Move left 2 pixels/frame.
        // Score when bird (X=100) passes pipe (X=90).
        if g.pipes[i].x < 90.0 && !g.pipes[i].scored {
            g.score++
            g.pipes[i].scored = true
        }
    }

    // Remove off-screen pipes.
    newPipes := []Pipe{}
    for _, pipe := range g.pipes {
        if pipe.x >= -50.0 {
            newPipes = append(newPipes, pipe)
        }
    }
    g.pipes = newPipes

    // Move ground left.
    g.groundX -= 2.0
    if g.groundX <= -400.0 {
        g.groundX += 400.0
    }

    // Check collisions with pipes.
    for _, pipe := range g.pipes {
        if pipe.x <= 110.0 && pipe.x >= 90.0 {
            gapTop := pipe.gapY - 50.0
            gapBottom := pipe.gapY + 50.0
            if g.birdY < gapTop || g.birdY > gapBottom {
                g.state = "gameOver"
                return nil
            }
        }
    }

    // Check ground and ceiling collisions.
    if g.birdY < 0 || g.birdY > 360 {
        g.state = "gameOver"
        return nil
    }

    return nil
}

// Draw: renders the game visuals each frame.
func (g *Game) Draw(screen *ebiten.Image) {
    if g.state == "gameOver" {
        ebitenutil.DebugPrint(screen, fmt.Sprintf("Game Over\nScore: %d\nPress R to Restart", g.score))
        return
    }

    // Draw background (scaled to 400x400).
    backgroundOpts := &ebiten.DrawImageOptions{}
    backgroundOpts.GeoM.Scale(400.0/288.0, 400.0/512.0) // Scale 288x512 to 400x400.
    screen.DrawImage(g.backgroundImg, backgroundOpts)

    // Draw ground.
    groundOpts := &ebiten.DrawImageOptions{}
    groundOpts.GeoM.Translate(g.groundX, 360.0)
    screen.DrawImage(g.groundImg, groundOpts)
    groundOpts.GeoM.Reset()
    groundOpts.GeoM.Translate(g.groundX+400.0, 360.0)
    screen.DrawImage(g.groundImg, groundOpts)

    // Draw bird.
    birdOpts := &ebiten.DrawImageOptions{}
    birdOpts.GeoM.Translate(100.0, g.birdY)
    screen.DrawImage(g.birdImg, birdOpts)

    // Draw pipes.
    for _, pipe := range g.pipes {
        gapTop := pipe.gapY - 50.0
        gapBottom := pipe.gapY + 50.0

        // Top pipe (rotated 180 degrees).
        topOpts := &ebiten.DrawImageOptions{}
        topOpts.GeoM.Rotate(3.14159)
        topOpts.GeoM.Translate(pipe.x+52.0, gapTop)
        screen.DrawImage(g.pipeImg, topOpts)

        // Bottom pipe.
        bottomOpts := &ebiten.DrawImageOptions{}
        bottomOpts.GeoM.Translate(pipe.x, gapBottom)
        screen.DrawImage(g.pipeImg, bottomOpts)
    }

    // Draw live score.
    ebitenutil.DebugPrint(screen, fmt.Sprintf("Score: %d", g.score))
}

// Layout: sets the screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return 400, 400
}

func main() {
    game := NewGame()
    ebiten.SetWindowSize(400, 400)
    ebiten.SetWindowTitle("Flappy Bird World Generator")
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}
