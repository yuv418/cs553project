// Deal with monolith vs microservice stuff.

package common

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	"connectrpc.com/connect"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yuv418/cs553project/backend/commondata"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
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
	dispatchTable map[string]*Action
	serviceData   map[string]AbstractionService
	CommonServer  *CommonServer
}

var AbsCtx = &AbstractionServer{
	Microservice:  GetMicroserviceStatus(),
	serviceData:   make(map[string]AbstractionService),
	dispatchTable: make(map[string]*Action),
	CommonServer:  NewCommonServer(),
}

func InsertServiceData(absCtx *AbstractionServer, key string, url string, prefix string) error {
	// https://stackoverflow.com/questions/57278822/sending-grpc-communications-over-a-specific-port
	// https://gist.github.com/marzocchi/c4d3e2254853c5ff02b420044e796aea
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	client, err := grpc.NewClient(
		url,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		log.Fatalf("(CAL) Couldn't add client for microservice with url %s, got error %s\n", url, err)
		return err
	} else {
		log.Printf("(CAL) Established client for %s at url %s\n", key, url)
	}
	absCtx.serviceData[key] = AbstractionService{
		url:    url,
		client: client,
		prefix: prefix,
	}

	return nil
}

func InsertDispatchTable(
	absCtx *AbstractionServer,
	svcName string,
	verb string,
) {
	absCtx.dispatchTable[verb] = &Action{
		verb:    verb,
		svcName: svcName,
	}
}

// TODO: set up web server as well.
func InsertDispatchTableHandler[ReqT any, RespT any](
	absCtx *AbstractionServer,
	svcName string,
	verb string,
	handlerFn any,
	shouldVerifyJwt bool,
) error {
	svcData := absCtx.serviceData[svcName]
	absCtx.dispatchTable[verb].fn = handlerFn.(any)
	// TODO dry
	route := svcData.prefix + "/" + verb
	log.Println("(CAL) Adding route ", route)

	AddRoute(absCtx.CommonServer, route,
		func(ctx context.Context, req *connect.Request[ReqT]) (*connect.Response[RespT], error) {
			var username string
			var jwtString string
			// Get username from ctx
			claims := ctx.Value("claims")
			if claims != nil {
				username = claims.(jwt.MapClaims)["username"].(string)
			}
			// Get JWT from ctx
			jwt := ctx.Value("jwt")
			if jwt != nil {
				jwtString = jwt.(string)
			}
			resp, err := (handlerFn.(func(*commondata.ReqCtx, *ReqT) (*RespT, error)))(&commondata.ReqCtx{
				HttpCtx:  &ctx,
				Username: username,
				Jwt:      jwtString,
			}, req.Msg)

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
		loc := svcData.prefix + "/" + dispatchTableData.verb
		log.Printf("(CAL Dispatch) Invoking microservice request on %s at %s\n", svcData.url, loc)

		// Create a context with the JWT as an authorization header
		md := map[string]string{
			"authorization": "Bearer " + ctx.Jwt,
		}
		callCtx := metadata.NewOutgoingContext(context.Background(), metadata.New(md))

		err := client.Invoke(callCtx, loc, req, resp)
		if err != nil {
			log.Printf("Request failed with %s\n", err)
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
	absCtx.CommonServer.StartServer()
}
