package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"

	"github.com/yuv418/cs553project/backend/commondata"
	authpb "github.com/yuv418/cs553project/backend/protos/auth"
)

// UserCredentials holds username and encrypted password.
type UserCredentials struct {
	Username          string `json:"username"`
	EncryptedPassword string `json:"encrypted_password"` // Base64 encoded + encrypted password
}

// Config holds server configuration.
type AuthConfig struct {
	TokenExpiry time.Duration
	UserFile    string // Path to the static user file
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

type AuthServer struct {
	jwtSecret      []byte
	encryptionKey  []byte
	tokenExpiry    time.Duration
	userStore      map[string]string
	userStoreMutex sync.RWMutex
	userFilePath   string
}

func NewAuthServer(jwtSecret string, tokenExpiry time.Duration, userFilePath string) (*AuthServer, error) {
	encryptionKey := deriveEncryptionKey(jwtSecret)
	server := &AuthServer{
		jwtSecret:     []byte(jwtSecret),
		encryptionKey: encryptionKey,
		tokenExpiry:   tokenExpiry,
		userStore:     make(map[string]string),
		userFilePath:  userFilePath,
	}
	err := server.loadUsers()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load users: %w", err)
	} else if os.IsNotExist(err) {
		log.Printf("User file '%s' not found. Will create if users are added.", userFilePath)
		server.addUser("admin", "password")
	}
	return server, nil
}

func (s *AuthServer) Authenticate(ctx *commondata.ReqCtx, c *authpb.AuthRequest) (*authpb.AuthResponse, error) {
	if c.Username == "" || c.Password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("username and password cannot be empty"))
	}

	s.userStoreMutex.RLock()
	encryptedPassword, exists := s.userStore[c.Username]
	s.userStoreMutex.RUnlock()

	if !exists {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid username or password"))
	}

	storedPasswordBytes, err := decrypt(encryptedPassword, s.encryptionKey)
	if err != nil {
		log.Printf("Failed to decrypt password for user '%s': %v", c.Username, err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("authentication processing error"))
	}

	if string(storedPasswordBytes) != c.Password {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid username or password"))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": c.Username,
		"exp":      time.Now().Add(s.tokenExpiry).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate token"))
	}

	response := &authpb.AuthResponse{}
	response.JwtToken = tokenString

	return response, nil
}

func (s *AuthServer) loadUsers() error {
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

func (s *AuthServer) addUser(username, password string) error {
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

func (s *AuthServer) saveUsers() error {
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

func LoadAuthConfig() (*AuthConfig, error) {
	cfg := &AuthConfig{}

	cfg.UserFile = commondata.GetEnv("AUTH_USER_FILE", "users.json")
	tokenExpiryStr := commondata.GetEnv("AUTH_TOKEN_EXPIRY", "6360h")

	expiry, err := time.ParseDuration(tokenExpiryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid token expiry duration '%s': %w", tokenExpiryStr, err)
	}
	cfg.TokenExpiry = expiry

	return cfg, nil
}
