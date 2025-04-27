// Deal with monolith vs microservice stuff.

package abstraction

import (
	engine "github.com/yuv418/cs553project/backend/game_engine"
	worldgen "github.com/yuv418/cs553project/backend/world_gen"
	"os"
)

type Action struct {
	verb string
	prefix string
	fn  any
}

var (
	Microservice = GetMicroserviceStatus()
	GameEnginePrefix = os.Getenv("GAME_ENGINE_URL") + "/game_engine.GameEngineService"
	WorldGenPrefix   = os.Getenv("WORLD_GEN_URL") + "/world_gen.WorldGenService"
	DispatchTable    = make(map[string]Action)
)

func GetMicroserviceStatus() bool {
	if os.Getenv("MICROSERVICE") == "1" {
		return true
	} else {
		return false
	}
}

func InsertDispatchTable(prefix string, verb string, any) {
	DispatchTable[verb] = Action{
		verb: verb,
		prefix: prefix,
		fn:  fn,
	}
}

func SetupDispatchTable() {
	InsertDispatchTable(GameEnginePrefix, "StartGame", engine.StartGame)
	InsertDispatchTable(GameEnginePrefix, "HandleInput", engine.HandleInput)

	InsertDispatchTable(WorldGenPrefix, "GenerateWorld", worldgen.GenerateWorld)
}
