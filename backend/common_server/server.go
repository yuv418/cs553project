package server

import "connectrpc.com/connect"

// Goal: provide common scaffolding for starting a connect gRPC web server.
// https://pkg.go.dev/connectrpc.com/connect#NewUnaryHandler

type SrvCfg struct {
	ListenAddr string
	CertFile   string
	KeyFile    string
	JWTSecret  []byte
}

type CommonServer struct {
	mux    *http.ServeMux
	server *http.Server
	cfg    *SrvCfg
}

func LoadSrvCfg() (*Config, error) {
	cfg := &SrvCfg{}

	flag.StringVar(&cfg.ListAddr, "addr", getEnv("AUTH_LISTEN_ADDR", ":50051"), "gRPC server listen address")
	flag.StringVar(&cfg.CertFile, "cert", getEnv("AUTH_CERT_FILE", "../../transport-server-demo/cert.pem"), "TLS certificate file path") // Default relative path
	flag.StringVar(&cfg.KeyFile, "key", getEnv("AUTH_KEY_FILE", "../../transport-server-demo/key.pem"), "TLS key file path")             // Default relative path
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

	commonSrv.cfg, err := LoadSrvCfg()
	if err != nil {
		log.Fatal(err)
	}

	commonSrv.mux := http.NewServeMux()

	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", commonSrv.cfg.ListenAddr)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, Connect-Protocol-Version")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		mux.ServeHTTP(w, r)
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

	return commonSrv
}

func (commonSrv *CommonSrv) AddRoute(route string, unaryRoute string,
	handlerFn func(context.Context, *Request[Req]) (*Response[Res], error)) {
	commonSrv.mux.Handle(route, connect.NewUnaryHandler(
		route+unaryRoute,
		handlerFn,
		connect.WithInterceptors(
			connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
				return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
					log.Printf("Request: %s", req.Spec().Procedure)

					// https://pkg.go.dev/net/http#Header
					// https://www.reddit.com/r/golang/comments/cgbkel/why_are_headers_mapstringstring/

					auth, ok := req.Header()["Authorization"]
					if ok && auth[0].startswith("Bearer") {
						jwt := auth[0].split(" ")[1]
						if commonSrv.cfg.ValidateJwt(jwt) {
							return next(ctx, req)
						}
					}

					// Return unauthorized
					// https://connectrpc.com/docs/go/errors/
					return nil, connect.NewError(connect.CodeUnauthenticated, "Missing or invalid JWT token")
				}
			}),
		),
	))

}

func (commonSrv *CommonSrv) StartServer() {
	log.Printf("Starting Connect server on %s", commonSrv.cfg.ListenAddr)
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		log.Fatal(commonSrv.server.ListenAndServeTLS(commonSrv.cfg.CertFile, commonSrv.cfg.KeyFile))
	} else {
		log.Fatal(commonSrv.server.ListenAndServe())
	}
}
