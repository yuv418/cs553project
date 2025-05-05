package initiator

import (
	"log"

	"github.com/google/uuid"
	"github.com/yuv418/cs553project/backend/common"
	"github.com/yuv418/cs553project/backend/commondata"
	enginepb "github.com/yuv418/cs553project/backend/protos/game_engine"
	initiatorpb "github.com/yuv418/cs553project/backend/protos/initiator"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
	"google.golang.org/protobuf/types/known/emptypb"
)

func StartGame(ctx *commondata.ReqCtx, req *initiatorpb.StartGameReq) (*initiatorpb.StartGameResp, error) {
	// Generate a random game id

	log.Printf("Got request params %v\n", req)

	gameIdUUID := uuid.New()
	gameId := gameIdUUID.String()

	generatedWorld, err := common.Dispatch[worldgenpb.WorldGenReq, worldgenpb.WorldGenerated](ctx, "Initiator", "GenerateWorld", &worldgenpb.WorldGenReq{
		GameId:         gameId,
		ViewportWidth:  req.ViewportWidth,
		ViewportHeight: req.ViewportHeight,
	})
	log.Printf("(initiator) Generated world for gameId %s...\n", gameId)

	if err != nil {
		return nil, err
	}

	_, err = common.Dispatch[enginepb.GameEngineStartReq, emptypb.Empty](ctx, "Initiator", "EngineStartGame", &enginepb.GameEngineStartReq{
		GameId:         gameId,
		ViewportWidth:  req.ViewportWidth,
		ViewportHeight: req.ViewportHeight,
		BirdWidth:      req.BirdWidth,
		BirdHeight:     req.BirdHeight,
		World:          generatedWorld,
	})
	if err != nil {
		return nil, err
	}

	return &initiatorpb.StartGameResp{
		GameId: gameId,
	}, nil
}
