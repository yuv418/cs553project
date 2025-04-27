package main

import (
	"log"
	"os"

	abstraction "github.com/yuv418/cs553project/backend/common"
	engine "github.com/yuv418/cs553project/backend/game_engine"
	worldgen "github.com/yuv418/cs553project/backend/world_gen"

	enginepb "github.com/yuv418/cs553project/backend/protos/game_engine"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func setupMonolithTables() {
	abstraction.InsertServiceData(abstraction.AbsCtx, "gameEngine", os.Getenv("GAME_ENGINE_URL"), "/game_engine.GameEngineService")
	abstraction.InsertServiceData(abstraction.AbsCtx, "worldGen", os.Getenv("WORLD_GEN_URL"), "/world_gen.WorldGenService")

	// Any internal microservice functions don't have to be validated.
	abstraction.InsertDispatchTable[enginepb.GameEngineStartReq, emptypb.Empty](abstraction.AbsCtx, "gameEngine", "StartGame", engine.StartGame, false)
	abstraction.InsertDispatchTable[enginepb.GameEngineInputReq, emptypb.Empty](abstraction.AbsCtx, "gameEngine", "HandleInput", engine.HandleInput, true)

	abstraction.InsertDispatchTable[worldgenpb.WorldGenReq, worldgenpb.WorldGenerated](abstraction.AbsCtx, "worldGen", "GenerateWorld", worldgen.GenerateWorld, false)
}

func main() {
	log.Printf("Microservices is set to %v\n", abstraction.AbsCtx.Microservice)

	setupMonolithTables()

	abstraction.AbsCtx.Run()
}
