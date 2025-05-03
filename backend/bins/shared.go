package main

import (
	"log"
	"os"

	auth "github.com/yuv418/cs553project/backend/auth"
	"github.com/yuv418/cs553project/backend/commondata"
	engine "github.com/yuv418/cs553project/backend/game_engine"
	"github.com/yuv418/cs553project/backend/initiator"
	music "github.com/yuv418/cs553project/backend/music"
	worldgen "github.com/yuv418/cs553project/backend/world_gen"

	abstraction "github.com/yuv418/cs553project/backend/common"
	authpb "github.com/yuv418/cs553project/backend/protos/auth"
	enginepb "github.com/yuv418/cs553project/backend/protos/game_engine"
	initiatorpb "github.com/yuv418/cs553project/backend/protos/initiator"
	musicpb "github.com/yuv418/cs553project/backend/protos/music"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func SetupAuthTables(ctx *abstraction.AbstractionServer) {
	cfg, err := auth.LoadAuthConfig()
	if err != nil {
		log.Fatalf("Auth failed with %s\n", err)
	}

	authServer, err := auth.NewAuthServer(ctx.CommonServer.Cfg.JWTSecret, cfg.TokenExpiry, cfg.UserFile)

	if err != nil {
		log.Fatalf("Failed to create auth server: %v", err)
	}

	abstraction.InsertServiceData(abstraction.AbsCtx, "auth", os.Getenv("AUTH_URL"), "/auth.AuthService")
	abstraction.InsertDispatchTable[authpb.AuthRequest, authpb.AuthResponse](abstraction.AbsCtx, "auth", "Authenticate", authServer.Authenticate, false)

}

func SetupInitiatorTables(ctx *abstraction.AbstractionServer) {
	abstraction.InsertServiceData(abstraction.AbsCtx, "initiator", os.Getenv("INITIATOR_URL"), "/initiator.InitiatorService")
	abstraction.InsertDispatchTable[initiatorpb.StartGameReq, initiatorpb.StartGameResp](abstraction.AbsCtx, "initiator", "StartGame", initiator.StartGame, true)
}

func SetupWorldgenTables(ctx *abstraction.AbstractionServer) {
	abstraction.InsertServiceData(abstraction.AbsCtx, "worldGen", os.Getenv("WORLD_GEN_URL"), "/world_gen.WorldGenService")
	abstraction.InsertDispatchTable[worldgenpb.WorldGenReq, worldgenpb.WorldGenerated](abstraction.AbsCtx, "worldGen", "GenerateWorld", worldgen.GenerateWorld, false)
}

func SetupGameEngineTables(ctx *abstraction.AbstractionServer) {
	abstraction.InsertServiceData(abstraction.AbsCtx, "gameEngine", os.Getenv("GAME_ENGINE_URL"), "/game_engine.GameEngineService")

	// Any internal microservice functions don't have to be validated.
	abstraction.InsertDispatchTable[enginepb.GameEngineStartReq, emptypb.Empty](abstraction.AbsCtx, "gameEngine", "EngineStartGame", engine.StartGame, false)
	abstraction.AddWebTransportRoute[enginepb.GameEngineInputReq, *enginepb.GameEngineInputReq, emptypb.Empty, *emptypb.Empty](abstraction.AbsCtx.CommonServer, "/gameEngine/GameSession", engine.HandleInput, engine.EstablishGameWebTransport)
}

func SetupMusicTables(ctx *abstraction.AbstractionServer) {
	abstraction.InsertServiceData(abstraction.AbsCtx, "music", os.Getenv("MUSIC_URL"), "/music.MusicService/PlayMusic")

	// Any internal microservice functions don't have to be validated.
	abstraction.InsertDispatchTable[musicpb.PlayMusicReq, emptypb.Empty](abstraction.AbsCtx, "music", "PlayMusic", music.PlayMusic, false)
	// Stub out the handler function because it'll never be used.
	abstraction.AddWebTransportRoute[emptypb.Empty, *emptypb.Empty, emptypb.Empty, *emptypb.Empty](
		abstraction.AbsCtx.CommonServer,
		"/music/MusicSession",
		func(ctx *commondata.ReqCtx, inp *emptypb.Empty) (*emptypb.Empty, error) {
			return &emptypb.Empty{}, nil
		},
		music.EstablishMusicWebTransport,
	)
}
