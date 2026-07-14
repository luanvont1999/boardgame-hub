package handler

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"boardgame-hub/backend/internal/middleware"
	"boardgame-hub/backend/internal/store"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	meetupStore = store.NewMeetupStore()
)

func Handler(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{
			Status:  "OK",
			Message: "Boardgame Hub API is running smoothly on Vercel",
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

		colors := []string{"#bca0f5", "#ffa4b2", "#ffe869", "#ffb875", "#9ee3b2", "#a4f0fd"}
		color := colors[rand.Intn(len(colors))]

		hostName := user.Name
		if hostName == "" {
			hostName = user.Email
		}
		if hostName == "" {
			hostName = "Ẩn danh"
		}

		newMeetup := store.Meetup{
			ID:            strconv.Itoa(rand.Intn(1000000)),
			Title:         req.Title,
			Game:          req.Game,
			HostName:      hostName,
			HostUID:       user.UID,
			Lat:           req.Lat,
			Lng:           req.Lng,
			PlayersCount:  1,
			PlayersNeeded: 4,
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

	mux.ServeHTTP(w, r)
}
