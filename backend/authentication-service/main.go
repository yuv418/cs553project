package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	authpb "github.com/yuv418/cs553project/backend/protos"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// UserCredentials holds username and encrypted password.
type UserCredentials struct {
	Username          string `json:"username"`
	EncryptedPassword string `json:"encrypted_password"` // Base64 encoded + encrypted password
}

// Config holds server configuration.
type Config struct {
	ListenAddr  string
	CertFile    string
	KeyFile     string
	JWTSecret   string
	TokenExpiry time.Duration
	UserFile    string // Path to the static user file
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.ListenAddr, "addr", getEnv("AUTH_LISTEN_ADDR", ":50051"), "gRPC server listen address")
	flag.StringVar(&cfg.CertFile, "cert", getEnv("AUTH_CERT_FILE", "../transport-server-demo/cert.pem"), "TLS certificate file path") // Default relative path
	flag.StringVar(&cfg.KeyFile, "key", getEnv("AUTH_KEY_FILE", "../transport-server-demo/key.pem"), "TLS key file path")             // Default relative path
	flag.StringVar(&cfg.JWTSecret, "jwt-secret", getEnv("AUTH_JWT_SECRET", "your-super-secret-key"), "Secret key for signing JWTs and encrypting passwords")
	flag.StringVar(&cfg.UserFile, "user-file", getEnv("AUTH_USER_FILE", "users.json"), "Path to the static user credentials file")
	tokenExpiryStr := flag.String("token-expiry", getEnv("AUTH_TOKEN_EXPIRY", "1h"), "JWT token expiry duration (e.g., 1h, 15m)")
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

	expiry, err := time.ParseDuration(*tokenExpiryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid token expiry duration '%s': %w", *tokenExpiryStr, err)
	}
	cfg.TokenExpiry = expiry

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Derive a 32-byte key for AES-256 from the JWT secret using SHA-256.
func deriveEncryptionKey(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:]
}

// Encrypt encrypts data using AES-GCM. Returns base64 encoded ciphertext.
func encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts data using AES-GCM. Expects base64 encoded ciphertext.
func decrypt(ciphertextBase64 string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedMessage := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		// Common error if the key is wrong or data corrupted
		return nil, fmt.Errorf("failed to decrypt (check key/data): %w", err)
	}

	return plaintext, nil
}

type authServer struct {
	authpb.UnimplementedAuthServiceServer
	jwtSecret      []byte
	encryptionKey  []byte // Derived key for AES
	tokenExpiry    time.Duration
	userStore      map[string]string // username -> encrypted password
	userStoreMutex sync.RWMutex
	userFilePath   string
}

func NewAuthServer(jwtSecret string, tokenExpiry time.Duration, userFilePath string) (*authServer, error) {
	encryptionKey := deriveEncryptionKey(jwtSecret)
	server := &authServer{
		jwtSecret:     []byte(jwtSecret),
		encryptionKey: encryptionKey,
		tokenExpiry:   tokenExpiry,
		userStore:     make(map[string]string),
		userFilePath:  userFilePath,
	}
	err := server.loadUsers()
	if err != nil && !os.IsNotExist(err) { // Ignore "file not found" on initial load
		return nil, fmt.Errorf("failed to load users: %w", err)
	} else if os.IsNotExist(err) {
		log.Printf("User file '%s' not found. Will create if users are added.", userFilePath)
		server.addUser("admin", "password") // Example
	}
	return server, nil
}

func (s *authServer) loadUsers() error {
	s.userStoreMutex.Lock()
	defer s.userStoreMutex.Unlock()

	data, err := os.ReadFile(s.userFilePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		log.Printf("User file '%s' is empty.", s.userFilePath)
		s.userStore = make(map[string]string) // Ensure store is empty
		return nil
	}

	var users []UserCredentials
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("failed to unmarshal user data from %s: %w", s.userFilePath, err)
	}

	s.userStore = make(map[string]string)
	for _, u := range users {
		s.userStore[u.Username] = u.EncryptedPassword
	}
	log.Printf("Loaded %d users from %s", len(s.userStore), s.userFilePath)
	return nil
}

func (s *authServer) addUser(username, password string) error {
	s.userStoreMutex.Lock()
	defer s.userStoreMutex.Unlock()

	if _, exists := s.userStore[username]; exists {
		return fmt.Errorf("user '%s' already exists", username)
	}

	encryptedPassword, err := encrypt([]byte(password), s.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt password for user '%s': %w", username, err)
	}

	s.userStore[username] = encryptedPassword
	log.Printf("Added user '%s' to in-memory store.", username)

	return s.saveUsers()
}

func (s *authServer) saveUsers() error {
	var users []UserCredentials
	for uname, encPass := range s.userStore {
		users = append(users, UserCredentials{Username: uname, EncryptedPassword: encPass})
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	err = os.WriteFile(s.userFilePath, data, 0644) // Write with standard permissions
	if err != nil {
		return fmt.Errorf("failed to write user file %s: %w", s.userFilePath, err)
	}
	log.Printf("Saved %d users to %s", len(users), s.userFilePath)
	return nil
}

// Authenticate handles authentication requests and generates JWTs.
func (s *authServer) Authenticate(ctx context.Context, req *authpb.AuthRequest) (*authpb.AuthResponse, error) {
	log.Printf("Authenticate request received for username: %s", req.Username)

	if req.Username == "" || req.Password == "" {
		return nil, status.Errorf(codes.InvalidArgument, "username and password cannot be empty")
	}

	s.userStoreMutex.RLock()
	encryptedPassword, ok := s.userStore[req.Username]
	s.userStoreMutex.RUnlock()

	if !ok {
		log.Printf("Authentication failed: username '%s' not found", req.Username)
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	decryptedPasswordBytes, err := decrypt(encryptedPassword, s.encryptionKey)
	if err != nil {
		log.Printf("Authentication failed: could not decrypt password for user '%s': %v", req.Username, err)
		return nil, status.Errorf(codes.Internal, "authentication processing error")
	}
	decryptedPassword := string(decryptedPasswordBytes)

	if req.Password != decryptedPassword {
		log.Printf("Authentication failed: incorrect password for username '%s'", req.Username)
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	// Generate JWT
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   req.Username,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenExpiry)),
		Issuer:    "authentication-service",
		// 'aud': []string{"service1", "service2"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		log.Printf("Error signing JWT for user '%s': %v", req.Username, err)
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	log.Printf("Generated JWT for username: %s", req.Username)
	return &authpb.AuthResponse{JwtToken: signedToken}, nil
}

func main() {
	log.Println("Starting Authentication Service...")

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded:")
	log.Printf("  Listen Address: %s", cfg.ListenAddr)
	log.Printf("  TLS Cert File: %s", cfg.CertFile)
	log.Printf("  TLS Key File: %s", cfg.KeyFile)
	log.Printf("  User File: %s", cfg.UserFile)
	log.Printf("  Token Expiry: %s", cfg.TokenExpiry)
	log.Printf("  JWT Secret: <hidden>")

	certificate, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		log.Fatalf("Failed to load TLS key pair: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	}

	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewTLS(tlsConfig)),
	}
	grpcServer := grpc.NewServer(opts...)

	authSrv, err := NewAuthServer(cfg.JWTSecret, cfg.TokenExpiry, cfg.UserFile)
	if err != nil {
		log.Fatalf("Failed to create auth server: %v", err)
	}
	authpb.RegisterAuthServiceServer(grpcServer, authSrv)

	lis, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", cfg.ListenAddr, err)
	}
	log.Printf("gRPC server listening securely on %s", cfg.ListenAddr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
