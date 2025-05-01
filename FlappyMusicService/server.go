package main

import (
    "bytes"           // For creating io.Reader from []byte
    "context"        // For handling gRPC request contexts
    "fmt"            // For formatting error messages
    "log"            // For logging server activity
    "net"            // For TCP network listener
    "os"             // For reading WAV files

    "github.com/hajimehoshi/ebiten/v2/audio"     // Ebiten audio package for playback
    "github.com/hajimehoshi/ebiten/v2/audio/wav" // WAV decoding support
    "google.golang.org/grpc"                     // gRPC server implementation

    "flappy-music-service/proto/musicpb" // Generated protobuf package
)

// Constants for server configuration
const (
    port       = ":50051" // gRPC server port
    sampleRate = 44100    // Audio sample rate (44.1 kHz, standard for WAV files)
)

// musicServer implements the MusicService gRPC interface
type musicServer struct {
    musicpb.UnimplementedMusicServiceServer // Embed for forward compatibility
    audioContext *audio.Context             // Shared audio context for playback
    soundFiles   map[musicpb.SoundEffect]string // Maps SoundEffect to WAV file paths
}

// newMusicServer initializes the server with audio context and sound file mappings
func newMusicServer() (*musicServer, error) {
    // Create a new Ebiten audio context with 44.1 kHz sample rate
    audioContext := audio.NewContext(sampleRate)

    // Define mapping of SoundEffect enums to WAV file paths
    soundFiles := map[musicpb.SoundEffect]string{
        musicpb.SoundEffect_JUMP:           "wing.wav",  // Flap sound
        musicpb.SoundEffect_SCORE_INCREASED: "point.wav", // Score increment sound
        musicpb.SoundEffect_DIE:            "hit.wav",   // Collision sound
    }

    return &musicServer{
        audioContext: audioContext,
        soundFiles:   soundFiles,
    }, nil
}

// loadSound reads a WAV file and creates a new audio.Player
func (s *musicServer) loadSound(filePath string) (*audio.Player, error) {
    // Read WAV file into memory
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read %s: %v", filePath, err)
    }

    // Convert []byte to io.Reader for WAV decoding
    reader := bytes.NewReader(data)

    // Decode WAV data with specified sample rate
    wavStream, err := wav.DecodeWithSampleRate(sampleRate, reader)
    if err != nil {
        return nil, fmt.Errorf("failed to decode WAV: %v", err)
    }

    // Create a new audio player from the decoded stream
    player, err := s.audioContext.NewPlayer(wavStream)
    if err != nil {
        return nil, fmt.Errorf("failed to create player: %v", err)
    }

    return player, nil
}

// PlayMusic implements the PlayMusic RPC to play a sound effect
func (s *musicServer) PlayMusic(ctx context.Context, req *musicpb.PlayMusicReq) (*musicpb.PlayMusicResp, error) {
    // Log incoming request for debugging
    log.Printf("Received PlayMusic request: game_id=%s, effect=%v", req.GameId, req.Effect)

    // Look up the WAV file path for the requested sound effect
    filePath, ok := s.soundFiles[req.Effect]
    if !ok {
        return nil, fmt.Errorf("unknown sound effect: %v", req.Effect)
    }

    // Create a new player for each playback to avoid state issues
    player, err := s.loadSound(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to load sound: %v", err)
    }

    // Play the sound (non-blocking)
    player.Play()

    // Return empty response (opus_payload is a placeholder for future streaming)
    return &musicpb.PlayMusicResp{OpusPayload: []byte{}}, nil
}

// main starts the gRPC server
func main() {
    // Create a TCP listener on port 50051
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    // Initialize the music server
    server, err := newMusicServer()
    if err != nil {
        log.Fatalf("Failed to create music server: %v", err)
    }

    // Create a new gRPC server
    grpcServer := grpc.NewServer()

    // Register the MusicService implementation
    musicpb.RegisterMusicServiceServer(grpcServer, server)

    // Log server start
    log.Printf("gRPC server listening on %s", port)

    // Start serving requests (blocks until error or shutdown)
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}