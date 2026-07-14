package middleware

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "firebase_user"

// FirebaseUser represents the authenticated user info extracted from Firebase ID Token
type FirebaseUser struct {
	UID           string `json:"uid"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GoogleCertificates cache manager
type GoogleCertificates struct {
	sync.RWMutex
	certs      map[string]string
	expiration time.Time
}

var (
	googleCerts = &GoogleCertificates{
		certs: make(map[string]string),
	}
	googleCertURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"
)

// fetchGooglePublicKeys retrieves public certificates from Google API
func (gc *GoogleCertificates) fetchKeys() (map[string]string, error) {
	gc.RLock()
	// Cache is valid for 1 hour
	if len(gc.certs) > 0 && time.Now().Before(gc.expiration) {
		defer gc.RUnlock()
		return gc.certs, nil
	}
	gc.RUnlock()

	gc.Lock()
	defer gc.Unlock()

	// Double check inside lock
	if len(gc.certs) > 0 && time.Now().Before(gc.expiration) {
		return gc.certs, nil
	}

	log.Println("[Auth] Fetching fresh public certificates from Google...")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(googleCertURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch certificates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch certificates: status %d", resp.StatusCode)
	}

	var newCerts map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&newCerts); err != nil {
		return nil, fmt.Errorf("failed to parse certificates JSON: %v", err)
	}

	gc.certs = newCerts
	// Refresh certificates every 30 minutes
	gc.expiration = time.Now().Add(30 * time.Minute)
	return gc.certs, nil
}

// verifyToken verifies the Firebase JWT token and returns claims
func verifyToken(tokenString string) (*FirebaseUser, error) {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		projectID = "boardgame-hub-demo" // Default fallback for local testing
		log.Println("[Auth] WARNING: FIREBASE_PROJECT_ID env var is not set. Using fallback: " + projectID)
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signature method is RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get key ID (kid) from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid header in token")
		}

		// Fetch public keys
		certs, err := googleCerts.fetchKeys()
		if err != nil {
			return nil, err
		}

		certPEM, ok := certs[kid]
		if !ok {
			return nil, fmt.Errorf("public key not found for kid: %s", kid)
		}

		// Parse PEM certificate
		block, _ := pem.Decode([]byte(certPEM))
		if block == nil {
			return nil, errors.New("failed to decode certificate PEM")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse x509 certificate: %v", err)
		}

		return cert.PublicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing/verification failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Validate Firebase-specific claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to parse token claims")
	}

	// Verify Issuer (iss)
	iss, _ := claims["iss"].(string)
	expectedIss := fmt.Sprintf("https://securetoken.google.com/%s", projectID)
	// Skip strict project id checking if using default placeholder project for easier local testing
	if projectID != "boardgame-hub-demo" && iss != expectedIss {
		return nil, fmt.Errorf("invalid issuer: got %s, expected %s", iss, expectedIss)
	}

	// Verify Audience (aud)
	aud, _ := claims["aud"].(string)
	if projectID != "boardgame-hub-demo" && aud != projectID {
		return nil, fmt.Errorf("invalid audience: got %s, expected %s", aud, projectID)
	}

	// Extract user claims
	uid, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	emailVerified, _ := claims["email_verified"].(bool)
	name, _ := claims["name"].(string)
	picture, _ := claims["picture"].(string)

	if uid == "" {
		return nil, errors.New("token subject (uid) is empty")
	}

	return &FirebaseUser{
		UID:           uid,
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
		Picture:       picture,
	}, nil
}

// AuthMiddleware wraps HTTP handlers to authenticate users using Firebase ID Tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle OPTIONS request for CORS preflight
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authorization header missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authorization header format must be Bearer <token>"})
			return
		}

		tokenString := parts[1]
		firebaseUser, err := verifyToken(tokenString)
		if err != nil {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			log.Printf("[Auth] Auth failed: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "Unauthorized",
				"details": err.Error(),
			})
			return
		}

		// Inject user into context
		ctx := context.WithValue(r.Context(), UserContextKey, firebaseUser)
		
		// Add CORS header to normal response
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
