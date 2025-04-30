package engine

// https://stackoverflow.com/questions/29721449/how-can-i-print-to-stderr-in-go-without-using-log

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/yuv418/cs553project/backend/common"
	"github.com/yuv418/cs553project/backend/commondata"
	framegenpb "github.com/yuv418/cs553project/backend/protos/frame_gen"
	enginepb "github.com/yuv418/cs553project/backend/protos/game_engine"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	frameRate = 30
)

type PlayState int8

const (
	Ready PlayState = iota
	Play
	Over
)

const (
	groundHeight = 112
	pipeWidth    = 52
	pipeGap      = 90 // Gap between upper and lower pipe
	gravity      = 0.25
	flapStrength = 4.6
)

type IndividualGameState struct {
	birdY        float64                    // Bird's vertical position (Y-coordinate, pixels).
	birdVelocity float64                    // Bird velocity
	flapForce    float64                    // Upward force when flapping (negative).
	world        *worldgenpb.WorldGenerated // Slice of pipes for obstacles.
	frame        int                        // Frame counter for timing (e.g., pipe spawning).
	score        int                        // Player’s score (increments when passing pipes).
	playState    PlayState                  // Game state: "ready," "play," "over".
	groundX      float64                    // Ground’s horizontal offset for scrolling (pixels).
}

type GameState struct {
	individualStateMap map[string]*IndividualGameState
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
	state.individualStateMap = make(map[string]*IndividualGameState)

	return state
}

func StartGame(ctx *commondata.ReqCtx, req *enginepb.GameEngineStartReq) (*emptypb.Empty, error) {
	GlobalStateLock.Lock()

	GlobalState.individualStateMap[req.GameId] = &IndividualGameState{
		birdY:        200,
		birdVelocity: 0,
		flapForce:    float64(req.ViewportHeight) / 10,
		world:        req.World,
		frame:        0,
		score:        0,
		playState:    Ready,
		// TODO maybe remove this
		groundX: 0,
	}

	GlobalStateLock.Unlock()

	return &emptypb.Empty{}, nil
}

func EstablishGameWebTransport(ctx *commondata.ReqCtx, transportWriter *bufio.Writer) error {

	// Acquire the WebTransport session for this username
	// https://gobyexample.com/timers
	// Somehow we want to pin this? Whatever

	log.Printf("EstablishGameWebTransport: user ID is %s game ID is %s\n", ctx.Username, ctx.GameId)

	// https://stackoverflow.com/questions/16466320/is-there-a-way-to-do-repetitive-tasks-at-intervals

	go (func() {

		timer := time.NewTicker((1000 / frameRate) * time.Millisecond)
		quit := make(chan struct{})

		for {
			select {
			case <-timer.C:
				// Get the game ID corresponding to everything

				common.WebTransportSendBuf(transportWriter, &framegenpb.GenerateFrameReq{})

			case <-quit:
				timer.Stop()
				return
			}

		}

	})()

	return nil
}

// This is a webtransport function, so returning nil will not send anything
func HandleInput(ctx *commondata.ReqCtx, inp *enginepb.GameEngineInputReq) (*emptypb.Empty, error) {
	log.Printf("Username in HandleInput is %s game ID is %s\n", ctx.Username, ctx.GameId)

	switch inp.Key {
	case enginepb.Key_SPACE:
		GlobalStateLock.Lock()
		statePtr := GlobalState.individualStateMap[ctx.GameId]

		if statePtr.playState == Ready {
			statePtr.playState = Play
		} else if statePtr.playState == Play {
			statePtr.birdVelocity = -flapStrength
		}

		GlobalStateLock.Unlock()
		break
	default:
		fmt.Fprintf(os.Stderr, "invalid key in Key enum %d\n", inp.Key)
		break
	}

	return nil, nil
}
