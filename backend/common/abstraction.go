// Deal with monolith vs microservice stuff.

package common

import (
	"context"
	"log"
	"os"

	"connectrpc.com/connect"

	engine "github.com/yuv418/cs553project/backend/game_engine"
	worldgen "github.com/yuv418/cs553project/backend/world_gen"

	enginepb "github.com/yuv418/cs553project/backend/protos/game_engine"
	worldgenpb "github.com/yuv418/cs553project/backend/protos/world_gen"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type Action struct {
	verb    string
	svcName string
	fn      any
}

func GetMicroserviceStatus() bool {
	if os.Getenv("MICROSERVICE") == "1" {
		return true
	} else {
		return false
	}
}

type AbstractionService struct {
	url    string
	prefix string
}

type AbstractionServer struct {
	microservice  bool
	dispatchTable map[string]Action
	serviceData   map[string]AbstractionService
	commonServer  *CommonServer
}

func NewAbstractionServer() *AbstractionServer {
	return &AbstractionServer{
		microservice: GetMicroserviceStatus(),
		serviceData: map[string]AbstractionService{
			"gameEngine": AbstractionService{
				url:    os.Getenv("GAME_ENGINE_URL"),
				prefix: "/game_engine.GameEngineService",
			},
			"worldGen": AbstractionService{
				url:    os.Getenv("WORLD_GEN_URL"),
				prefix: "/world_gen.WorldGenService",
			},
		},
		dispatchTable: make(map[string]Action),
		commonServer:  NewCommonServer(),
	}
}

// TODO: set up web server as well.
func InsertDispatchTable[ReqT any, RespT any](
	absCtx *AbstractionServer,
	svcName string,
	verb string,
	handlerFn any,
) {
	absCtx.dispatchTable[verb] = Action{
		verb:    verb,
		svcName: svcName,
		fn:      handlerFn,
	}

	route := absCtx.serviceData[svcName].prefix + "/" + verb
	log.Println("(CAL) Adding route ", route)

	AddRoute(absCtx.commonServer, route,
		func(ctx context.Context, req *connect.Request[ReqT]) (*connect.Response[RespT], error) {
			return connect.NewResponse((handlerFn.(func(*ReqT) *RespT))(req.Msg)), nil
		})
}

func Dispatch[Req any, Resp any](absCtx *AbstractionServer, verb string, req *Req) (*Resp, error) {
	var empty Resp
	if absCtx.microservice {
		// STUB
		return &empty, nil

	} else {
		return (absCtx.dispatchTable[verb].fn.(func(*Req) *Resp)(req)), nil
	}
}

func (absCtx *AbstractionServer) SetupMonolithDispatchTable() {
	InsertDispatchTable[enginepb.GameEngineStartReq, emptypb.Empty](absCtx, "gameEngine", "StartGame", engine.StartGame)
	InsertDispatchTable[enginepb.GameEngineInputReq, emptypb.Empty](absCtx, "gameEngine", "HandleInput", engine.HandleInput)

	InsertDispatchTable[worldgenpb.WorldGenReq, worldgenpb.WorldGenerated](absCtx, "worldGen", "GenerateWorld", worldgen.GenerateWorld)
}

func (absCtx *AbstractionServer) Run() {
	absCtx.commonServer.StartServer()
}
