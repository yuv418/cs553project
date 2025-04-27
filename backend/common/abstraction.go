// Deal with monolith vs microservice stuff.

package common

import (
	"context"
	"log"
	"os"

	"connectrpc.com/connect"

	"github.com/yuv418/cs553project/backend/commondata"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	client *grpc.ClientConn
}

type AbstractionServer struct {
	Microservice  bool
	dispatchTable map[string]Action
	serviceData   map[string]AbstractionService
	commonServer  *CommonServer
}

var AbsCtx = &AbstractionServer{
	Microservice:  GetMicroserviceStatus(),
	serviceData:   make(map[string]AbstractionService),
	dispatchTable: make(map[string]Action),
	commonServer:  NewCommonServer(),
}

func InsertServiceData(absCtx *AbstractionServer, key string, url string, prefix string) error {
	client, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	absCtx.serviceData[key] = AbstractionService{
		url:    url,
		client: client,
		prefix: prefix,
	}

	return nil
}

// TODO: set up web server as well.
func InsertDispatchTable[ReqT any, RespT any](
	absCtx *AbstractionServer,
	svcName string,
	verb string,
	handlerFn any,
	shouldVerifyJwt bool,
) error {
	svcData := absCtx.serviceData[svcName]
	absCtx.dispatchTable[verb] = Action{
		verb:    verb,
		svcName: svcName,
		fn:      handlerFn,
	}

	// TODO dry
	route := svcData.prefix + "/" + verb
	log.Println("(CAL) Adding route ", route)

	AddRoute(absCtx.commonServer, route,
		func(ctx context.Context, req *connect.Request[ReqT]) (*connect.Response[RespT], error) {
			resp, err := (handlerFn.(func(*commondata.ReqCtx, *ReqT) (*RespT, error)))(&commondata.ReqCtx{HttpCtx: &ctx}, req.Msg)

			if err != nil {
				return nil, err
			} else {
				return connect.NewResponse(resp), nil

			}
		}, shouldVerifyJwt)

	return nil
}

func Dispatch[Req any, Resp any](ctx *commondata.ReqCtx, verb string, req *Req) (*Resp, error) {
	// https://sahansera.dev/building-grpc-client-go/
	if AbsCtx.Microservice {
		// https://pkg.go.dev/google.golang.org/grpc#ClientConn.Invoke
		// Adapted from protobuf generated svcs
		dispatchTableData := AbsCtx.dispatchTable[verb]
		svcData := AbsCtx.serviceData[dispatchTableData.svcName]
		client := svcData.client

		// https://www.freecodecamp.org/news/new-vs-make-functions-in-go/
		// https://chatgpt.com/share/680de978-f87c-8012-bd76-a8a6ae618438
		resp := new(Resp)
		err := client.Invoke(context.Background(), svcData.prefix+"/"+dispatchTableData.verb, req, resp)
		if err != nil {
			return nil, err
		} else {
			return resp, nil
		}

	} else {
		// Dispatch some stuff
		returnedResp, err := (AbsCtx.dispatchTable[verb].fn.(func(*commondata.ReqCtx, *Req) (*Resp, error))(ctx, req))
		if err != nil {
			return nil, err
		} else {
			return returnedResp, nil
		}
	}
}

func (absCtx *AbstractionServer) Run() {
	absCtx.commonServer.StartServer()
}
