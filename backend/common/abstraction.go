// Deal with monolith vs microservice stuff.

package common

import (
	"context"
	"log"
	"os"

	"connectrpc.com/connect"

	"github.com/yuv418/cs553project/backend/commondata"
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

var AbsCtx = &AbstractionServer{
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

// TODO: set up web server as well.
func InsertDispatchTable[ReqT any, RespT any](
	absCtx *AbstractionServer,
	svcName string,
	verb string,
	handlerFn any,
	shouldVerifyJwt bool,
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
			return connect.NewResponse((handlerFn.(func(*commondata.ReqCtx, *ReqT) *RespT))(&commondata.ReqCtx{HttpCtx: &ctx}, req.Msg)), nil
		}, shouldVerifyJwt)
}

func Dispatch[Req any, Resp any](ctx *commondata.ReqCtx, verb string, req *Req) (*Resp, error) {
	var empty Resp
	if AbsCtx.microservice {
		// STUB
		return &empty, nil

	} else {
		return (AbsCtx.dispatchTable[verb].fn.(func(*commondata.ReqCtx, *Req) *Resp)(ctx, req)), nil
	}
}

func (absCtx *AbstractionServer) SetupMonolithDispatchTable() {
	// Any internal microservice functions don't have to be validated.
	InsertDispatchTable[enginepb.GameEngineStartReq, emptypb.Empty](absCtx, "gameEngine", "StartGame", engine.StartGame, false)
	InsertDispatchTable[enginepb.GameEngineInputReq, emptypb.Empty](absCtx, "gameEngine", "HandleInput", engine.HandleInput, true)

	InsertDispatchTable[worldgenpb.WorldGenReq, worldgenpb.WorldGenerated](absCtx, "worldGen", "GenerateWorld", worldgen.GenerateWorld, false)
}

func (absCtx *AbstractionServer) Run() {
	absCtx.commonServer.StartServer()
}
