package engine

// https://stackoverflow.com/questions/29721449/how-can-i-print-to-stderr-in-go-without-using-log

import (
	"bufio"
	"fmt"
	"log"
	"math"
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
	gravity      = 0.25
	flapStrength = 4.6
	maxPipeSpeed = 5
	birdX        = 50
)

type IndividualGameState struct {
	birdY        float64                    // Bird's vertical position (Y-coordinate, pixels).
	birdVelocity float64                    // Bird velocity
	flapForce    float64                    // Upward force when flapping (negative).
	world        *worldgenpb.WorldGenerated // Slice of pipes for obstacles.
	frame        int32                      // Frame counter for timing (e.g., pipe spawning).
	score        int32                      // Player’s score (increments when passing pipes).
	playState    PlayState                  // Game state: "ready," "play," "over".
	groundX      float64                    // Ground’s horizontal offset for scrolling (pixels).
	pipeSpeed    float64
	// Full height/Y
	pipeWindowX     float64
	pipeWindowWidth float64
	pipesToRender   int
	pipeStartOffset float64
	pipeStarts      []float64
	pipePositions   []float64
	pipeGaps        []float64
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

	// TODO: Validate that the game ID doesn't already exist.
	game := &IndividualGameState{
		birdY:        200,
		birdVelocity: 0,
		flapForce:    float64(req.ViewportHeight) / 10,
		world:        req.World,
		frame:        0,
		score:        0,
		playState:    Ready,
		// TODO maybe remove this
		groundX:         0,
		pipeSpeed:       2,
		pipeWindowX:     0,
		pipeWindowWidth: float64(req.ViewportWidth),
		pipesToRender:   int(req.ViewportWidth) / (pipeWidth + int(req.World.PipeSpacing)),
		pipeStartOffset: float64(req.World.PipeSpacing) * 1.5,
	}

	GlobalState.individualStateMap[req.GameId] = game

	GlobalStateLock.Unlock()

	return &emptypb.Empty{}, nil
}

func EstablishGameWebTransport(ctx *commondata.ReqCtx, transportWriter *bufio.Writer) error {

	// Acquire the WebTransport session for this username
	// https://gobyexample.com/timers
	// Somehow we want to pin this? Whatever

	log.Printf("EstablishGameWebTransport: user ID is %s game ID is %s\n", ctx.Username, ctx.GameId)

	// https://stackoverflow.com/questions/16466320/is-there-a-way-to-do-repetitive-tasks-at-intervals

	gameId := ctx.GameId

	go (func() {

		timer := time.NewTicker((1000 / frameRate) * time.Millisecond)
		quit := make(chan struct{})
		GlobalStateLock.Lock()
		pipesToRender := GlobalState.individualStateMap[gameId].pipesToRender
		GlobalStateLock.Unlock()

		frameUpdate := &framegenpb.GenerateFrameReq{
			GameId:    gameId,
			PipeWidth: pipeWidth,
			BirdPosition: &framegenpb.Pos{
				X: birdX,
			},
			PipePositions: make([]float64, pipesToRender, pipesToRender),
			PipeStarts:    make([]float64, pipesToRender, pipesToRender),
			PipeGaps:      make([]float64, pipesToRender, pipesToRender),
		}

		for {
			select {
			case <-timer.C:
				// Get the game ID corresponding to everything

				// TODO: Should we just lock to get the statePtr here and then
				// lock again to write the statePtr, or what?
				GlobalStateLock.Lock()
				statePtr := GlobalState.individualStateMap[gameId]
				GlobalStateLock.Unlock()

				statePtr.birdVelocity += gravity
				statePtr.birdY += statePtr.birdVelocity

				// TODO check collisions

				// Advance the pipe window
				statePtr.pipeWindowX += statePtr.pipeSpeed

				advanceAmt := pipeWidth + statePtr.world.PipeSpacing
				// Render the pipes
				for i := range statePtr.pipesToRender {
					// Find the closest pipe to
					// pipeWindowX + (i*advanceAmt)

					closestPipe := 0

					if statePtr.pipeWindowX > statePtr.world.PipeSpacing {
						adj := ((statePtr.pipeWindowX + (float64(i) * advanceAmt)) - statePtr.world.PipeSpacing)
						closestPipe = int(math.Ceil(adj / advanceAmt))
					}

					closestPipePos := (float64(closestPipe) * advanceAmt) + statePtr.pipeStartOffset

					frameUpdate.PipePositions[i] = closestPipePos - statePtr.pipeWindowX
					frameUpdate.PipeGaps[i] = statePtr.world.PipeSpecs[i].GapHeight
					frameUpdate.PipeStarts[i] = statePtr.world.PipeSpecs[i].GapStart

				}

				statePtr.score++
				// Increase difficulty slightly
				if statePtr.score%5 == 0 && statePtr.pipeSpeed < maxPipeSpeed {
					statePtr.pipeSpeed += 0.5
				}

				frameUpdate.Score = statePtr.score
				frameUpdate.BirdPosition.Y = statePtr.birdY

				common.WebTransportSendBuf(transportWriter, frameUpdate)
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
