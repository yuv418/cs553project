package common

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// Goal: provide common scaffolding for starting a connect gRPC web server.
// https://pkg.go.dev/connectrpc.com/connect#NewUnaryHandler

type SrvCfg struct {
	ListenAddr string
	CertFile   string
	KeyFile    string
	JWTSecret  string
}

type CommonServer struct {
	mux    *http.ServeMux
	server *http.Server
	Cfg    *SrvCfg
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
		w.Header().Set("Access-Control-Allow-Origin", commonSrv.Cfg.ListenAddr)
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

		commonSrv.server = &http.Server{
			Addr:    cfg.ListenAddr,
			Handler: corsHandler,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS12,
			},
		}
	} else {
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
	log.Printf("Starting Connect server on %s", commonSrv.Cfg.ListenAddr)
	if commonSrv.Cfg.CertFile != "" && commonSrv.Cfg.KeyFile != "" {
		log.Fatal(commonSrv.server.ListenAndServeTLS(commonSrv.Cfg.CertFile, commonSrv.Cfg.KeyFile))
	} else {
		log.Fatal(commonSrv.server.ListenAndServe())
	}
}
