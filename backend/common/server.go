package common

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
	"github.com/yuv418/cs553project/backend/commondata"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protodelim"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Goal: provide common scaffolding for starting a connect gRPC web server.
// https://pkg.go.dev/connectrpc.com/connect#NewUnaryHandler

type SrvCfg struct {
	ListenAddr    string
	WtpListenAddr string
	CertFile      string
	KeyFile       string
	JWTSecret     string
}

type CommonServer struct {
	mux       *http.ServeMux
	wtpMux    *http.ServeMux
	server    *http.Server
	wtpServer *webtransport.Server
	cert      tls.Certificate
	Cfg       *SrvCfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func LoadSrvCfg() (*SrvCfg, error) {
	cfg := &SrvCfg{}

	flag.StringVar(&cfg.ListenAddr, "addr", getEnv("AUTH_LISTEN_ADDR", ":50051"), "gRPC server listen address")
	flag.StringVar(&cfg.WtpListenAddr, "wtp-addr", getEnv("AUTH_LISTEN_WTP_ADDR", ":4433"), "Webtransport server listen address")
	flag.StringVar(&cfg.CertFile, "cert", getEnv("AUTH_CERT_FILE", "../transport-server-demo/cert.pem"), "TLS certificate file path") // Default relative path
	flag.StringVar(&cfg.KeyFile, "key", getEnv("AUTH_KEY_FILE", "../transport-server-demo/key.pem"), "TLS key file path")             // Default relative path
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", getEnv("AUTH_JWT_SECRET", "your-super-secret-key"), "Secret key for signing JWTs and encrypting passwords")
	flag.Parse()

	if cfg.JWTSecret == "your-super-secret-key" {
		log.Println("Warning: Using default JWT secret. Set AUTH_JWT_SECRET environment variable or --jwt-secret flag for production.")
	}
	if _, err := os.Stat(cfg.CertFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("TLS cert file not found: %s. Set AUTH_CERT_FILE or --cert flag", cfg.CertFile)
	}
	if _, err := os.Stat(cfg.KeyFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("TLS key file not found: %s. Set AUTH_KEY_FILE or --key flag", cfg.KeyFile)
	}
	return cfg, nil
}

func NewCommonServer() *CommonServer {
	commonSrv := &CommonServer{}

	cfg, err := LoadSrvCfg()
	if err != nil {
		log.Fatal(err)
	}
	commonSrv.Cfg = cfg

	commonSrv.mux = http.NewServeMux()

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, Connect-Protocol-Version")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		commonSrv.mux.ServeHTTP(w, r)
	})

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			log.Fatalf("Failed to load TLS certificates: %v", err)
		}
		commonSrv.cert = cert

		commonSrv.server = &http.Server{
			Addr:    cfg.ListenAddr,
			Handler: corsHandler,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{commonSrv.cert},
				MinVersion:   tls.VersionTLS12,
			},
		}
	} else {
		// Webtransport server will fail.

		commonSrv.server = &http.Server{
			Addr:    cfg.ListenAddr,
			Handler: h2c.NewHandler(corsHandler, &http2.Server{}),
		}
	}

	commonSrv.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Hello, %q\n", html.EscapeString(r.URL.Path))
	})

	return commonSrv
}

func SetupWebTransport(commonSrv *CommonServer) {
	// https://gist.github.com/filewalkwithme/0199060b2cb5bbc478c5

	log.Printf("(CALServer) Setting up web transport at %s\n", commonSrv.Cfg.WtpListenAddr)

	commonSrv.wtpMux = http.NewServeMux()
	commonSrv.wtpServer =
		&webtransport.Server{
			H3: http3.Server{
				Handler: commonSrv.wtpMux,
				Addr:    commonSrv.Cfg.WtpListenAddr,
				TLSConfig: &tls.Config{
					Certificates: []tls.Certificate{commonSrv.cert},
					NextProtos:   []string{"h3"},
				},
			},
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for demo!
				return true
			},
		}
}

func WebTransportSendBuf[Res any, PtrRes interface {
	ProtoReflect() protoreflect.Message
	*Res
}](byteWriter *bufio.Writer, resp PtrRes) error {
	ptrResp := PtrRes(resp)

	protodelim.MarshalTo(byteWriter, ptrResp)

	// For latency reasons
	byteWriter.Flush()
	return nil
}

// TODO: add auth
// https://stackoverflow.com/questions/69573113/how-can-i-instantiate-a-non-nil-pointer-of-type-argument-with-generic-go/69575720#69575720
// https://pkg.go.dev/bufio#Writer.Flush
func AddWebTransportRoute[Req any, PtrReq interface {
	ProtoReflect() protoreflect.Message
	*Req
}, Res any, PtrRes interface {
	ProtoReflect() protoreflect.Message
	*Res
}](
	commonSrv *CommonServer,
	route string,
	handlerFn func(*commondata.ReqCtx, *Req) (*Res, error),
	insertWebTransport func(*commondata.ReqCtx, *bufio.Writer) error,
) {

	if commonSrv.wtpServer == nil {
		SetupWebTransport(commonSrv)
	}

	log.Printf("(CALServer) Adding WebTransport route %s\n", route)

	commonSrv.wtpMux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		session, err := commonSrv.wtpServer.Upgrade(w, r)
		if err != nil {
			log.Printf("failed to upgrade: %v", err)
			http.Error(w, "failed to upgrade", http.StatusInternalServerError)
			return
		}

		log.Printf("Received a WebTransport Connection at %s\n", route)
		go (func(session *webtransport.Session) {
			defer session.CloseWithError(0, "session closed")

			for {
				stream, err := session.AcceptStream(context.Background())
				if err != nil {
					log.Printf("failed to accept stream: %s", err)
					return
				}
				go (func(stream webtransport.Stream) {

					log.Printf("Beginning stream for connection\n")
					buf := PtrReq(new(Req))
					reqCtx := &commondata.ReqCtx{}

					// https://pkg.go.dev/io#ByteScanner
					byteReader := bufio.NewReader(stream)
					byteWriter := bufio.NewWriter(stream)

					insertWebTransport(reqCtx, byteWriter)

					for {
						err := protodelim.UnmarshalFrom(byteReader, buf)

						if err != nil {
							log.Printf("Failed to unmarshal protobuf %s\n", err)
							// We closed the stream
							if err == io.EOF {
								return
							}
							continue
						}

						// TODO add the header for JWT
						resp, err := handlerFn(reqCtx, buf)
						ptrResp := PtrRes(resp)
						if err != nil {
							log.Printf("Handler for WebTransport failed! %s\n", err)
						} else {
							WebTransportSendBuf(byteWriter, ptrResp)
						}

					}

				})(stream)
			}

		})(session)
	})
}

func AddRoute[Req any, Res any](commonSrv *CommonServer, route string,
	handlerFn func(context.Context, *connect.Request[Req]) (*connect.Response[Res], error),
	shouldVerifyJwt bool) {
	commonSrv.mux.Handle(route, connect.NewUnaryHandler(
		route,
		handlerFn,
		connect.WithInterceptors(
			connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
				return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
					log.Printf("Request: %s", req.Spec().Procedure)

					if shouldVerifyJwt {
						// https://pkg.go.dev/net/http#Header
						// https://www.reddit.com/r/golang/comments/cgbkel/why_are_headers_mapstringstring/
						// https://pkg.go.dev/strings
						// https://go.dev/doc/tutorial/handle-errors

						auth, ok := req.Header()["Authorization"]
						if ok && strings.HasPrefix(auth[0], "Bearer") {
							jwt := strings.Split(auth[0], " ")[1]
							if commonSrv.Cfg.ValidateJwt(jwt) {
								return next(ctx, req)
							}
						}

						// Return unauthorized
						// https://connectrpc.com/docs/go/errors/
						return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("Missing or invalid JWT token"))
					} else {
						return next(ctx, req)
					}

				}
			}),
		),
	))

}

func (commonSrv *CommonServer) StartServer() {
	if commonSrv.Cfg.CertFile != "" && commonSrv.Cfg.KeyFile != "" {
		if commonSrv.wtpServer != nil {
			log.Printf("Starting WebTransport server at %s\n", commonSrv.Cfg.WtpListenAddr)
			go (func() {
				commonSrv.wtpServer.ListenAndServeTLS(commonSrv.Cfg.CertFile, commonSrv.Cfg.KeyFile)
			})()
		}
		log.Printf("Starting Connect server on %s", commonSrv.Cfg.ListenAddr)
		log.Fatal(commonSrv.server.ListenAndServeTLS(commonSrv.Cfg.CertFile, commonSrv.Cfg.KeyFile))
	} else {
		log.Fatal(commonSrv.server.ListenAndServe())
	}
}
