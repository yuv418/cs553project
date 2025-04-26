package main

// https://stackoverflow.com/questions/29721449/how-can-i-print-to-stderr-in-go-without-using-log

import (
	"fmt"
	"os"

	enginepb "github.com/yuv418/cs553project/backend/protos/game_engine"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
)

type GameState struct {
	birdY     float64                    // Bird's vertical position (Y-coordinate, pixels).
	gravity   float64                    // Gravity force per frame (pixels).
	flapForce float64                    // Upward force when flapping (negative).
	world     *worldgenpb.WorldGenerated // Slice of pipes for obstacles.
	frame     int                        // Frame counter for timing (e.g., pipe spawning).
	score     int                        // Player’s score (increments when passing pipes).
	playState string                     // Game state: "playing" or "gameOver".
	groundX   float64                    // Ground’s horizontal offset for scrolling (pixels).
}

func StartGame(req *enginepb.GameEngineStartReq, games map[int32]GameState) {
	games[req.GameId] = GameState{
		birdY:     0,
		gravity:   float64(req.ViewportWidth) / 10,
		flapForce: float64(req.ViewportHeight) / 10,
		world:     req.World,
		frame:     0,
		score:     0,
		playState: "playing",
		// TODO maybe remove this
		groundX: 0,
	}
}

func (state *GameState) HandleInput(inp *enginepb.GameEngineInputReq) {
	switch inp.Key {
	case enginepb.Key_SPACE:
		// Handle key space
		break
	default:
		fmt.Fprintf(os.Stderr, "invalid key in Key enum %d\n", inp.Key)
		break
	}
}
