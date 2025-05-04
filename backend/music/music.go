package music

import (
	"fmt"
	"log"
	"sync"

	"embed"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/quic-go/webtransport-go"
	"github.com/yuv418/cs553project/backend/common"
	"github.com/yuv418/cs553project/backend/commondata"
	musicpb "github.com/yuv418/cs553project/backend/protos/music"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	//go:embed audio/wing.ogg
	//go:embed audio/point.ogg
	//go:embed audio/hit.ogg
	musicFiles      embed.FS
	MusicServer     = newMusicServer()
	MusicServerLock = sync.Mutex{}
)

type musicServer struct {
	soundFiles   map[musicpb.SoundEffect][]byte // Maps SoundEffect to WAV file paths
	transportMap map[string]*commondata.WebTransportHandle
}

// newMusicServer initializes the server with audio context and sound file mappings
func newMusicServer() *musicServer {

	// TODO: efficientize
	wingBin, err := musicFiles.ReadFile("audio/wing.ogg")
	if err != nil {
		log.Panic("couldn't read wing.ogg")
	}
	pointBin, err := musicFiles.ReadFile("audio/point.ogg")
	if err != nil {
		log.Panic("couldn't read audio/point.ogg")
	}
	hitBin, err := musicFiles.ReadFile("audio/hit.ogg")
	if err != nil {
		log.Panic("couldn't read audio/hitBin.ogg")
	}

	// Define mapping of SoundEffect enums to WAV file paths
	soundFiles := map[musicpb.SoundEffect][]byte{
		musicpb.SoundEffect_JUMP:            wingBin,  // Flap sound
		musicpb.SoundEffect_SCORE_INCREASED: pointBin, // Score increment sound
		musicpb.SoundEffect_DIE:             hitBin,   // Collision sound
	}

	return &musicServer{
		soundFiles:   soundFiles,
		transportMap: make(map[string]*commondata.WebTransportHandle),
	}
}

// loadSound reads a WAV file and creates a new audio.Player
func EstablishMusicWebTransport(ctx *commondata.ReqCtx, handle *commondata.WebTransportHandle) error {

	// Acquire the WebTransport session for this username
	// https://gobyexample.com/timers
	// Somehow we want to pin this? Whatever

	log.Printf("EstablishMusicWebTransport: user ID is %s game ID is %s\n", ctx.Username, ctx.GameId)

	// https://stackoverflow.com/questions/16466320/is-there-a-way-to-do-repetitive-tasks-at-intervals

	MusicServerLock.Lock()

	MusicServer.transportMap[ctx.GameId] = handle

	MusicServerLock.Unlock()

	return nil
}

// PlayMusic implements the PlayMusic RPC to play a sound effect
func PlayMusic(ctx *commondata.ReqCtx, req *musicpb.PlayMusicReq) (*empty.Empty, error) {
	// Log incoming request for debugging

	MusicServerLock.Lock()
	defer MusicServerLock.Unlock()

	log.Printf("Received PlayMusic request: game_id=%s, effect=%v, MusicServer=%v", req.GameId, req.Effect, MusicServer.transportMap)

	for k, _ := range MusicServer.transportMap {
		log.Printf("game is %s want %s", k, ctx.GameId)
		if k == ctx.GameId {
			log.Printf("Matched game id")
		}
	}

	gameTransport := MusicServer.transportMap[ctx.GameId]
	if gameTransport == nil {
		return nil, fmt.Errorf("unknown game ID: %v", ctx.GameId)
	}

	// Look up the WAV file path for the requested sound effect
	effectBin := MusicServer.soundFiles[req.Effect]
	if effectBin == nil {
		return nil, fmt.Errorf("unknown sound effect: %v", req.Effect)
	}

	respPb := &musicpb.PlayMusicResp{AudioPayload: effectBin}
	common.WebTransportSendBuf(gameTransport.Writer, respPb)

	// Close stream if this is a DIE message
	if req.Effect == musicpb.SoundEffect_DIE {
		log.Printf("Closing audio stream")
		(*gameTransport.WtStream.(*webtransport.Stream)).Close()
	}

	// Return empty response (opus_payload is a placeholder for future streaming)
	return &emptypb.Empty{}, nil
}
