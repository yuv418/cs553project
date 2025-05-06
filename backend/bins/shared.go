package main

import (
	"log"
	"os"

	auth "github.com/yuv418/cs553project/backend/auth"
	"github.com/yuv418/cs553project/backend/commondata"
	engine "github.com/yuv418/cs553project/backend/game_engine"
	"github.com/yuv418/cs553project/backend/initiator"
	music "github.com/yuv418/cs553project/backend/music"
	"github.com/yuv418/cs553project/backend/score"
	worldgen "github.com/yuv418/cs553project/backend/world_gen"

	abstraction "github.com/yuv418/cs553project/backend/common"
	authpb "github.com/yuv418/cs553project/backend/protos/auth"
	enginepb "github.com/yuv418/cs553project/backend/protos/game_engine"
	initiatorpb "github.com/yuv418/cs553project/backend/protos/initiator"
	musicpb "github.com/yuv418/cs553project/backend/protos/music"
	scorepb "github.com/yuv418/cs553project/backend/protos/score"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func SetupServiceData(ctx *abstraction.AbstractionServer) {
	abstraction.InsertServiceData(abstraction.AbsCtx, "score", os.Getenv("SCORE_URL"), "/score.ScoreService")
	abstraction.InsertServiceData(abstraction.AbsCtx, "auth", os.Getenv("AUTH_URL"), "/auth.AuthService")
	abstraction.InsertServiceData(abstraction.AbsCtx, "initiator", os.Getenv("INITIATOR_URL"), "/initiator.InitiatorService")
	abstraction.InsertServiceData(abstraction.AbsCtx, "music", os.Getenv("MUSIC_URL"), "/music.MusicService/PlayMusic")
	abstraction.InsertServiceData(abstraction.AbsCtx, "worldGen", os.Getenv("WORLD_GEN_URL"), "/world_gen.WorldGenService")
	abstraction.InsertServiceData(abstraction.AbsCtx, "gameEngine", os.Getenv("GAME_ENGINE_URL"), "/game_engine.GameEngineService")
}

func SetupDispatchTable(ctx *abstraction.AbstractionServer) {
	abstraction.InsertDispatchTable(abstraction.AbsCtx, "auth", "Authenticate")
	abstraction.InsertDispatchTable(abstraction.AbsCtx, "initiator", "StartGame")
	abstraction.InsertDispatchTable(abstraction.AbsCtx, "gameEngine", "EngineStartGame")
	abstraction.InsertDispatchTable(abstraction.AbsCtx, "initiator", "StartGame")
	abstraction.InsertDispatchTable(abstraction.AbsCtx, "worldGen", "GenerateWorld")
	abstraction.InsertDispatchTable(abstraction.AbsCtx, "gameEngine", "EngineStartGame")
	abstraction.InsertDispatchTable(abstraction.AbsCtx, "music", "PlayMusic")
	abstraction.InsertDispatchTable(abstraction.AbsCtx, "score", "UpdateScore")
	abstraction.InsertDispatchTable(abstraction.AbsCtx, "score", "GetScores")
}

func SetupScoreHandler(ctx *abstraction.AbstractionServer) {
	scoreCtx, err := score.LoadScoreCtx()
	if err != nil {
		log.Fatalf("Failed to load auth score context with %s\n", err)
	}

	abstraction.InsertDispatchTableHandler[scorepb.ScoreEntry, emptypb.Empty](abstraction.AbsCtx, "score", "UpdateScore", scoreCtx.UpdateScore, true)
	abstraction.InsertDispatchTableHandler[emptypb.Empty, scorepb.GetScoresResp](abstraction.AbsCtx, "score", "GetScores", scoreCtx.GetScores, true)

}

func SetupAuthHandler(ctx *abstraction.AbstractionServer) {
	cfg, err := auth.LoadAuthConfig()
	if err != nil {
		log.Fatalf("Auth config load failed with %s\n", err)
	}

	authServer, err := auth.NewAuthServer(ctx.CommonServer.Cfg.JWTSecret, cfg.TokenExpiry, cfg.UserFile)

	if err != nil {
		log.Fatalf("Failed to create auth server: %v", err)
	}

	abstraction.InsertDispatchTableHandler[authpb.AuthRequest, authpb.AuthResponse](abstraction.AbsCtx, "auth", "Authenticate", authServer.Authenticate, false)

}

func SetupInitiatorHandler(ctx *abstraction.AbstractionServer) {
	abstraction.InsertDispatchTableHandler[initiatorpb.StartGameReq, initiatorpb.StartGameResp](abstraction.AbsCtx, "initiator", "StartGame", initiator.StartGame, true)
}

func SetupWorldgenHandler(ctx *abstraction.AbstractionServer) {
	abstraction.InsertDispatchTableHandler[worldgenpb.WorldGenReq, worldgenpb.WorldGenerated](abstraction.AbsCtx, "worldGen", "GenerateWorld", worldgen.GenerateWorld, false)
}

func SetupGameEngineHandler(ctx *abstraction.AbstractionServer) {

	// Any internal microservice functions don't have to be validated.
	abstraction.InsertDispatchTableHandler[enginepb.GameEngineStartReq, emptypb.Empty](abstraction.AbsCtx, "gameEngine", "EngineStartGame", engine.StartGame, false)
	abstraction.AddWebTransportRoute[enginepb.GameEngineInputReq, *enginepb.GameEngineInputReq, emptypb.Empty, *emptypb.Empty](
		abstraction.AbsCtx.CommonServer,
		"GameEngine",
		"/gameEngine/GameSession",
		engine.HandleInput,
		engine.EstablishGameWebTransport,
	)
}

func SetupMusicHandler(ctx *abstraction.AbstractionServer) {

	// Any internal microservice functions don't have to be validated.
	abstraction.InsertDispatchTableHandler[musicpb.PlayMusicReq, emptypb.Empty](abstraction.AbsCtx, "music", "PlayMusic", music.PlayMusic, false)
	// Stub out the handler function because it'll never be used.
	abstraction.AddWebTransportRoute[emptypb.Empty, *emptypb.Empty, emptypb.Empty, *emptypb.Empty](
		abstraction.AbsCtx.CommonServer,
		"Music",
		"/music/MusicSession",
		func(ctx *commondata.ReqCtx, inp *emptypb.Empty) (*emptypb.Empty, error) {
			return &emptypb.Empty{}, nil
		},
		music.EstablishMusicWebTransport,
	)
}
