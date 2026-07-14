package main

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"boardgame-hub/backend/pkg/middleware"
	"boardgame-hub/backend/pkg/store"

	"github.com/golang-jwt/jwt/v5"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ServiceAccount struct {
	Type        string `json:"type"`
	ProjectID   string `json:"project_id"`
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
	TokenURI    string `json:"token_uri"`
}

func getAccessToken() (string, error) {
	saPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON")
	if saPath == "" {
		saPath = "firebase-service-account.json"
	}

	data, err := ioutil.ReadFile(saPath)
	if err != nil {
		return "", fmt.Errorf("không thể đọc file service account %s: %v", saPath, err)
	}

	var sa ServiceAccount
	if err := json.Unmarshal(data, &sa); err != nil {
		return "", fmt.Errorf("lỗi unmarshal service account: %v", err)
	}

	// Parse private key
	block, _ := pem.Decode([]byte(sa.PrivateKey))
	if block == nil {
		return "", errors.New("lỗi decode PEM private key")
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("lỗi parse private key: %v", err)
	}

	// Tạo claims
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   sa.ClientEmail,
		"sub":   sa.ClientEmail,
		"aud":   sa.TokenURI,
		"scope": "https://www.googleapis.com/auth/firebase.messaging",
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	jwtToken, err := token.SignedString(privKey)
	if err != nil {
		return "", fmt.Errorf("lỗi ký token: %v", err)
	}

	// Request access token từ Google
	resp, err := http.PostForm(sa.TokenURI, url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {jwtToken},
	})
	if err != nil {
		return "", fmt.Errorf("lỗi gửi request lấy access token: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("lỗi decode access token response: %v", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("lỗi Google OAuth: %s", result.Error)
	}

	return result.AccessToken, nil
}

func sendFCMNotification(token, title, body string) error {
	accessToken, err := getAccessToken()
	if err != nil {
		return fmt.Errorf("lỗi lấy access token: %v", err)
	}

	projectID := "boardgame-hub-7f7a2"
	urlStr := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID)

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": token,
			"notification": map[string]string{
				"title": title,
				"body":  body,
			},
		},
	}

	payloadData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("lỗi marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(payloadData))
	if err != nil {
		return fmt.Errorf("lỗi tạo request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("lỗi gửi request tới FCM: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("FCM API trả về lỗi status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func main() {
	// Initialize in-memory store
	meetupStore := store.NewMeetupStore()

	// Seed random generator
	rand.Seed(time.Now().UnixNano())

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow CORS for local dev
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{
			Status:  "OK",
			Message: "Boardgame Luna API is running smoothly",
		})
	})

	// Protected Profile Endpoint
	profileHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(middleware.UserContextKey).(*middleware.FirebaseUser)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "User context not found"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Kết nối API xác thực thành công!",
			"user":    user,
		})
	})
	mux.Handle("/api/profile", middleware.AuthMiddleware(profileHandler))

	// Public GET Meetups endpoint
	mux.HandleFunc("/api/meetups", func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS preflight
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(meetupStore.GetAll())
	})

	// Protected POST Create Meetup endpoint
	createMeetupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		user, ok := r.Context().Value(middleware.UserContextKey).(*middleware.FirebaseUser)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "User context not found"})
			return
		}

		// Parse request body
		var req struct {
			Title string  `json:"title"`
			Game  string  `json:"game"`
			Lat   float64 `json:"lat"`
			Lng   float64 `json:"lng"`
			Time  string  `json:"time"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body: " + err.Error()})
			return
		}

		if req.Title == "" || req.Game == "" || req.Lat == 0 || req.Lng == 0 || req.Time == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields (title, game, lat, lng, time)"})
			return
		}

		// Select random pastel color
		colors := []string{"#bca0f5", "#ffa4b2", "#ffe869", "#ffb875", "#9ee3b2", "#a4f0fd"}
		color := colors[rand.Intn(len(colors))]

		// Get host display name or email fallback
		hostName := user.Name
		if hostName == "" {
			hostName = user.Email
		}
		if hostName == "" {
			hostName = "Ẩn danh"
		}

		// Create meetup object
		newMeetup := store.Meetup{
			ID:            strconv.Itoa(rand.Intn(1000000)),
			Title:         req.Title,
			Game:          req.Game,
			HostName:      hostName,
			HostUID:       user.UID,
			Lat:           req.Lat,
			Lng:           req.Lng,
			PlayersCount:  1, // Host is the first player
			PlayersNeeded: 4, // Default needed
			Time:          req.Time,
			Color:         color,
		}

		meetupStore.Add(newMeetup)
		log.Printf("[Meetups] Created new meetup: %s by %s", newMeetup.Title, newMeetup.HostName)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newMeetup)
	})
	mux.Handle("/api/meetups/create", middleware.AuthMiddleware(createMeetupHandler))

	// Public endpoint for sending push notifications via Go Backend Proxy
	mux.HandleFunc("/api/send-notification", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		var req struct {
			FCMToken string `json:"fcmToken"`
			Title    string `json:"title"`
			Body     string `json:"body"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		if req.FCMToken == "" || req.Title == "" || req.Body == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields (fcmToken, title, body)"})
			return
		}

		tokenHint := req.FCMToken
		if len(tokenHint) > 15 {
			tokenHint = tokenHint[:15] + "..."
		}
		log.Printf("[FCM] Đang gửi thông báo đẩy tới token: %s...", tokenHint)
		err := sendFCMNotification(req.FCMToken, req.Title, req.Body)
		if err != nil {
			log.Printf("[FCM ERROR] Gửi push thất bại: %v", err)
			// Trả về OK với warning để tránh làm crash app client khi người dùng chưa cấu hình service-account file
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"warning": "Không thể gửi push thật: " + err.Error() + ". Vui lòng đặt file firebase-service-account.json ở server.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Đã gửi thông báo đẩy thành công!",
		})
	})

	log.Println("Server is starting on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
