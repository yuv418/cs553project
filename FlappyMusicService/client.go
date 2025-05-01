package main

import (
    "context" // For creating gRPC request contexts
    "fmt"     // For printing test completion message
    "log"     // For logging client activity
    "time"    // For pausing between sound requests

    "google.golang.org/grpc" // gRPC client implementation
    "flappy-music-service/proto/musicpb" // Generated protobuf package
)

// main tests the MusicService by sending PlayMusic requests
func main() {
    // Connect to the gRPC server on localhost:50051
    // Use WithInsecure for testing (no TLS)
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    // Ensure connection is closed when main exits
    defer conn.Close()

    // Create a MusicService client from the connection
    client := musicpb.NewMusicServiceClient(conn)

    // Define the sound effects to test
    soundEffects := []musicpb.SoundEffect{
        musicpb.SoundEffect_JUMP,           // wing.wav
        musicpb.SoundEffect_SCORE_INCREASED, // point.wav
        musicpb.SoundEffect_DIE,            // hit.wav
    }

    // Iterate through each sound effect
    for _, effect := range soundEffects {
        // Create a PlayMusic request with a dummy game_id
        req := &musicpb.PlayMusicReq{
            GameId: "test-game-123",
            Effect: effect,
        }

        // Send the request to the server
        resp, err := client.PlayMusic(context.Background(), req)
        if err != nil {
            log.Printf("Failed to play %v: %v", effect, err)
            continue
        }

        // Log the successful response
        log.Printf("Played %v, response: %v", effect, resp.OpusPayload)

        // Pause for 1 second to avoid overlapping sounds
        time.Sleep(1 * time.Second)
    }

    // Print completion message
    fmt.Println("Test completed")
}