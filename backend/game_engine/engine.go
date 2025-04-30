package engine

// https://stackoverflow.com/questions/29721449/how-can-i-print-to-stderr-in-go-without-using-log

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/yuv418/cs553project/backend/commondata"
	enginepb "github.com/yuv418/cs553project/backend/protos/game_engine"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	frameRate = 30
)

type IndividualGameState struct {
	birdY     float64                    // Bird's vertical position (Y-coordinate, pixels).
	gravity   float64                    // Gravity force per frame (pixels).
	flapForce float64                    // Upward force when flapping (negative).
	world     *worldgenpb.WorldGenerated // Slice of pipes for obstacles.
	frame     int                        // Frame counter for timing (e.g., pipe spawning).
	score     int                        // Player’s score (increments when passing pipes).
	playState string                     // Game state: "playing" or "gameOver".
	groundX   float64                    // Ground’s horizontal offset for scrolling (pixels).
}

type GameState struct {
	individualStateMap map[string]IndividualGameState
}

type SessionState struct {
	invidualSessionMap map[string]*bufio.Writer
}

var (
	GlobalSessionState = MakeSessionState()
	GlobalState        = MakeGameState()
	GlobalStateLock    = sync.Mutex{}
)

func MakeSessionState() *SessionState {
	state := &SessionState{
		invidualSessionMap: make(map[string]*bufio.Writer),
	}
	return state
}

func MakeGameState() *GameState {
	state := &GameState{}
	state.individualStateMap = make(map[string]IndividualGameState)

	return state
}

func StartGame(ctx *commondata.ReqCtx, req *enginepb.GameEngineStartReq) (*emptypb.Empty, error) {
	GlobalStateLock.Lock()

	GlobalState.individualStateMap[req.GameId] = IndividualGameState{
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

	// Start game loop

	GlobalStateLock.Unlock()

	return &emptypb.Empty{}, nil
}

func EstablishGameWebTransport(ctx *commondata.ReqCtx, transportWriter *bufio.Writer) error {

	// Acquire the WebTransport session for this username
	// https://gobyexample.com/timers
	// Somehow we want to pin this? Whatever

	log.Printf("EstablishGameWebTransport: user ID is %s\n", ctx.Username)

	go (func() {

		timer := time.NewTimer((1000 / frameRate) * time.Millisecond)
		for {
			<-timer.C

			GlobalStateLock.Lock()

			GlobalSessionState.invidualSessionMap[ctx.Username] = transportWriter

			GlobalStateLock.Unlock()
		}

	})()

	return nil
}

func HandleInput(ctx *commondata.ReqCtx, inp *enginepb.GameEngineInputReq) (*emptypb.Empty, error) {
	log.Printf("Username in HandleInput is %s\n", ctx.Username)

	switch inp.Key {
	case enginepb.Key_SPACE:
		GlobalStateLock.Lock()
		// Handle key space
		GlobalStateLock.Unlock()
		break
	default:
		fmt.Fprintf(os.Stderr, "invalid key in Key enum %d\n", inp.Key)
		break
	}

	return &emptypb.Empty{}, nil
}
