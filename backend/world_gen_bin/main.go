package main

import (
	"log"
	"os"

	abstraction "github.com/yuv418/cs553project/backend/common"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"
	worldgen "github.com/yuv418/cs553project/backend/world_gen"
)

func setupMicroserviceTables() {
	abstraction.InsertServiceData(abstraction.AbsCtx, "worldGen", os.Getenv("WORLD_GEN_URL"), "/world_gen.WorldGenService")
	abstraction.InsertDispatchTable[worldgenpb.WorldGenReq, worldgenpb.WorldGenerated](abstraction.AbsCtx, "worldGen", "GenerateWorld", worldgen.GenerateWorld, false)
}

func main() {
	log.Printf("Microservices is set to %v\n", abstraction.AbsCtx.Microservice)

	setupMicroserviceTables()

	abstraction.AbsCtx.Run()
}
