package server

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
	"strings"
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

var (
	meetupStore = store.NewMeetupStore()
)

func getAccessToken() (string, error) {
	var sa ServiceAccount

	// Thử đọc trực tiếp nội dung JSON từ biến môi trường
	saJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON")
	if saJSON != "" {
		if err := json.Unmarshal([]byte(saJSON), &sa); err != nil {
			return "", fmt.Errorf("lỗi unmarshal service account từ biến môi trường: %v", err)
		}
	} else {
		// Fallback đọc file cục bộ (local dev)
		saPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON_PATH")
		if saPath == "" {
			saPath = "firebase-service-account.json"
		}

		data, err := ioutil.ReadFile(saPath)
		if err != nil {
			return "", fmt.Errorf("không thể đọc file service account %s: %v", saPath, err)
		}

		if err := json.Unmarshal(data, &sa); err != nil {
			return "", fmt.Errorf("lỗi unmarshal service account từ file: %v", err)
		}
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
		"scope": "https://www.googleapis.com/auth/cloud-platform",
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

func sendFCMNotification(token, title, body, clickAction string) error {
	accessToken, err := getAccessToken()
	if err != nil {
		return fmt.Errorf("lỗi lấy access token: %v", err)
	}

	projectID := "boardgame-hub-7f7a2"
	urlStr := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID)

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": token,
			"data": map[string]string{
				"title":       title,
				"body":        body,
				"clickAction": clickAction,
			},
			"webpush": map[string]interface{}{
				"headers": map[string]string{
					"Urgency": "high",
				},
				"fcm_options": map[string]string{
					"link": clickAction,
				},
			},
			"apns": map[string]interface{}{
				"headers": map[string]string{
					"apns-priority": "10",
				},
				"payload": map[string]interface{}{
					"aps": map[string]interface{}{
						"alert": map[string]string{
							"title": title,
							"body":  body,
						},
						"sound":             "default",
						"mutable-content":   1,
						"content-available": 1,
					},
				},
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

// ── FIRESTORE REST API HELPERS ─────────────────────────────────────────────

type FirestoreField struct {
	StringValue  string        `json:"stringValue,omitempty"`
	IntegerValue string        `json:"integerValue,omitempty"`
	ArrayValue   *FirestoreArray `json:"arrayValue,omitempty"`
	MapValue     *FirestoreMap   `json:"mapValue,omitempty"`
}

type FirestoreArray struct {
	Values []FirestoreField `json:"values"`
}

type FirestoreMap struct {
	Fields map[string]FirestoreField `json:"fields"`
}

type FirestoreDocument struct {
	Name   string                    `json:"name,omitempty"`
	Fields map[string]FirestoreField `json:"fields"`
}

type MeetupData struct {
	ID            string
	Title         string
	Game          string
	HostName      string
	HostUID       string
	HostFcmToken  string
	PlayersCount  int
	PlayersNeeded int
	Time               string
	Color              string
	PendingUids        []string
	ApprovedPendingUids []string
	ApprovedUids       []string
	UserFcmTokens      map[string]string
}

func parseFirestoreDocument(doc *FirestoreDocument) *MeetupData {
	if doc == nil || doc.Fields == nil {
		return nil
	}

	m := &MeetupData{
		UserFcmTokens:       make(map[string]string),
		PendingUids:         []string{},
		ApprovedPendingUids: []string{},
		ApprovedUids:        []string{},
	}

	parts := strings.Split(doc.Name, "/")
	if len(parts) > 0 {
		m.ID = parts[len(parts)-1]
	}

	if f, ok := doc.Fields["title"]; ok {
		m.Title = f.StringValue
	}
	if f, ok := doc.Fields["game"]; ok {
		m.Game = f.StringValue
	}
	if f, ok := doc.Fields["hostName"]; ok {
		m.HostName = f.StringValue
	} else if f, ok := doc.Fields["host_name"]; ok {
		m.HostName = f.StringValue
	}
	if f, ok := doc.Fields["hostUid"]; ok {
		m.HostUID = f.StringValue
	} else if f, ok := doc.Fields["host_uid"]; ok {
		m.HostUID = f.StringValue
	}
	if f, ok := doc.Fields["hostFcmToken"]; ok {
		m.HostFcmToken = f.StringValue
	}
	if f, ok := doc.Fields["playersCount"]; ok {
		m.PlayersCount, _ = strconv.Atoi(f.IntegerValue)
	} else if f, ok := doc.Fields["players_count"]; ok {
		m.PlayersCount, _ = strconv.Atoi(f.IntegerValue)
	}
	if f, ok := doc.Fields["playersNeeded"]; ok {
		m.PlayersNeeded, _ = strconv.Atoi(f.IntegerValue)
	} else if f, ok := doc.Fields["players_needed"]; ok {
		m.PlayersNeeded, _ = strconv.Atoi(f.IntegerValue)
	}
	if f, ok := doc.Fields["time"]; ok {
		m.Time = f.StringValue
	}
	if f, ok := doc.Fields["color"]; ok {
		m.Color = f.StringValue
	}

	if f, ok := doc.Fields["pendingUids"]; ok && f.ArrayValue != nil {
		for _, val := range f.ArrayValue.Values {
			if val.StringValue != "" {
				m.PendingUids = append(m.PendingUids, val.StringValue)
			}
		}
	}
	if f, ok := doc.Fields["approvedPendingUids"]; ok && f.ArrayValue != nil {
		for _, val := range f.ArrayValue.Values {
			if val.StringValue != "" {
				m.ApprovedPendingUids = append(m.ApprovedPendingUids, val.StringValue)
			}
		}
	}
	if f, ok := doc.Fields["approvedUids"]; ok && f.ArrayValue != nil {
		for _, val := range f.ArrayValue.Values {
			if val.StringValue != "" {
				m.ApprovedUids = append(m.ApprovedUids, val.StringValue)
			}
		}
	}

	if f, ok := doc.Fields["userFcmTokens"]; ok && f.MapValue != nil {
		for key, val := range f.MapValue.Fields {
			m.UserFcmTokens[key] = val.StringValue
		}
	}

	return m
}

func buildFirestoreDocument(m *MeetupData) *FirestoreDocument {
	fields := make(map[string]FirestoreField)

	fields["title"] = FirestoreField{StringValue: m.Title}
	fields["game"] = FirestoreField{StringValue: m.Game}
	fields["hostName"] = FirestoreField{StringValue: m.HostName}
	fields["hostUid"] = FirestoreField{StringValue: m.HostUID}
	fields["hostFcmToken"] = FirestoreField{StringValue: m.HostFcmToken}
	fields["playersCount"] = FirestoreField{IntegerValue: strconv.Itoa(m.PlayersCount)}
	fields["playersNeeded"] = FirestoreField{IntegerValue: strconv.Itoa(m.PlayersNeeded)}
	fields["time"] = FirestoreField{StringValue: m.Time}
	fields["color"] = FirestoreField{StringValue: m.Color}

	// pendingUids
	pendingVals := []FirestoreField{}
	for _, uid := range m.PendingUids {
		pendingVals = append(pendingVals, FirestoreField{StringValue: uid})
	}
	fields["pendingUids"] = FirestoreField{ArrayValue: &FirestoreArray{Values: pendingVals}}

	// approvedPendingUids
	appPendingVals := []FirestoreField{}
	for _, uid := range m.ApprovedPendingUids {
		appPendingVals = append(appPendingVals, FirestoreField{StringValue: uid})
	}
	fields["approvedPendingUids"] = FirestoreField{ArrayValue: &FirestoreArray{Values: appPendingVals}}

	// approvedUids
	approvedVals := []FirestoreField{}
	for _, uid := range m.ApprovedUids {
		approvedVals = append(approvedVals, FirestoreField{StringValue: uid})
	}
	fields["approvedUids"] = FirestoreField{ArrayValue: &FirestoreArray{Values: approvedVals}}

	// userFcmTokens
	tokenFields := make(map[string]FirestoreField)
	for key, val := range m.UserFcmTokens {
		tokenFields[key] = FirestoreField{StringValue: val}
	}
	fields["userFcmTokens"] = FirestoreField{MapValue: &FirestoreMap{Fields: tokenFields}}

	return &FirestoreDocument{Fields: fields}
}

func getFirestoreMeetup(meetupId string) (*MeetupData, error) {
	accessToken, err := getAccessToken()
	if err != nil {
		return nil, err
	}

	urlStr := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/boardgame-hub-7f7a2/databases/(default)/documents/meetups/%s", meetupId)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("lỗi đọc meetup từ Firestore status %d: %s", resp.StatusCode, string(body))
	}

	var doc FirestoreDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	return parseFirestoreDocument(&doc), nil
}

func updateFirestoreMeetup(m *MeetupData, updateFields []string) error {
	accessToken, err := getAccessToken()
	if err != nil {
		return err
	}

	doc := buildFirestoreDocument(m)
	bodyData, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	u, _ := url.Parse(fmt.Sprintf("https://firestore.googleapis.com/v1/projects/boardgame-hub-7f7a2/databases/(default)/documents/meetups/%s", m.ID))
	q := u.Query()
	for _, field := range updateFields {
		q.Add("updateMask.fieldPaths", field)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("PATCH", u.String(), bytes.NewBuffer(bodyData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("lỗi cập nhật meetup trên Firestore status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func setFirestoreRequest(meetupId, userUid, userName, status string) error {
	accessToken, err := getAccessToken()
	if err != nil {
		return err
	}

	urlStr := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/boardgame-hub-7f7a2/databases/(default)/documents/meetups/%s/requests/%s", meetupId, userUid)

	fields := map[string]FirestoreField{
		"uid":    {StringValue: userUid},
		"name":   {StringValue: userName},
		"status": {StringValue: status},
	}
	doc := FirestoreDocument{Fields: fields}
	bodyData, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", urlStr, bytes.NewBuffer(bodyData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("lỗi ghi request trên Firestore status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func updateFirestoreRequestStatus(meetupId, userUid, status string) error {
	accessToken, err := getAccessToken()
	if err != nil {
		return err
	}

	urlStr := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/boardgame-hub-7f7a2/databases/(default)/documents/meetups/%s/requests/%s?updateMask.fieldPaths=status", meetupId, userUid)

	doc := FirestoreDocument{
		Fields: map[string]FirestoreField{
			"status": {StringValue: status},
		},
	}
	bodyData, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", urlStr, bytes.NewBuffer(bodyData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("lỗi cập nhật status request: %s", string(body))
	}
	return nil
}

func deleteFirestoreRequest(meetupId, userUid string) error {
	accessToken, err := getAccessToken()
	if err != nil {
		return err
	}

	urlStr := fmt.Sprintf("https://firestore.googleapis.com/v1/projects/boardgame-hub-7f7a2/databases/(default)/documents/meetups/%s/requests/%s", meetupId, userUid)
	req, err := http.NewRequest("DELETE", urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("lỗi xóa request trên Firestore status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ── ROUTER INITIALIZATION ──────────────────────────────────────────────────

// NewRouter khởi tạo ServeMux chứa toàn bộ API endpoints của Boardgame Hub dùng chung
func NewRouter() *http.ServeMux {
	rand.Seed(time.Now().UnixNano())

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
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
			FCMToken    string   `json:"fcmToken"`
			FCMTokens   []string `json:"fcmTokens"`
			Title       string   `json:"title"`
			Body        string   `json:"body"`
			ClickAction string   `json:"clickAction"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		if req.Title == "" || req.Body == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields (title, body)"})
			return
		}

		if req.ClickAction == "" {
			req.ClickAction = "/"
		}

		// Thu thập tất cả các token cần gửi
		var tokens []string
		if req.FCMToken != "" {
			tokens = append(tokens, req.FCMToken)
		}
		if len(req.FCMTokens) > 0 {
			tokens = append(tokens, req.FCMTokens...)
		}

		if len(tokens) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "At least one fcmToken or fcmTokens must be provided"})
			return
		}

		log.Printf("[FCM] Đang gửi thông báo đẩy tới %d thiết bị...", len(tokens))

		var sendErrors []string
		for _, token := range tokens {
			tokenHint := token
			if len(tokenHint) > 15 {
				tokenHint = tokenHint[:15] + "..."
			}
			log.Printf("[FCM] Đang gửi tới token: %s...", tokenHint)
			err := sendFCMNotification(token, req.Title, req.Body, req.ClickAction)
			if err != nil {
				log.Printf("[FCM ERROR] Gửi push thất bại tới token %s: %v", tokenHint, err)
				sendErrors = append(sendErrors, fmt.Sprintf("%s: %v", tokenHint, err))
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if len(sendErrors) > 0 && len(sendErrors) == len(tokens) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"warning": "Tất cả các lượt gửi đều thất bại. Vui lòng kiểm tra lại cấu hình hoặc token.",
				"errors":  sendErrors,
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Đã gửi thông báo đẩy thành công tới %d/%d thiết bị!", len(tokens)-len(sendErrors), len(tokens)),
			"errors":  sendErrors,
		})
	})

	// POST /api/meetups/join
	mux.HandleFunc("/api/meetups/join", func(w http.ResponseWriter, r *http.Request) {
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
			MeetupID string `json:"meetupId"`
			UserUID  string `json:"userUid"`
			UserName string `json:"userName"`
			FCMToken string `json:"fcmToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		if req.MeetupID == "" || req.UserUID == "" || req.UserName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields"})
			return
		}

		// 1. Tạo request tài liệu trong subcollection requests
		err := setFirestoreRequest(req.MeetupID, req.UserUID, req.UserName, "pending")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi lưu request lên Firestore: " + err.Error()})
			return
		}

		// 2. Lấy dữ liệu meetup để cập nhật pendingUids và userFcmTokens
		meetup, err := getFirestoreMeetup(req.MeetupID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi lấy dữ liệu meetup: " + err.Error()})
			return
		}

		// Thêm vào pendingUids nếu chưa có
		found := false
		for _, uid := range meetup.PendingUids {
			if uid == req.UserUID {
				found = true
				break
			}
		}
		if !found {
			meetup.PendingUids = append(meetup.PendingUids, req.UserUID)
		}

		// Lưu token vào map
		if req.FCMToken != "" {
			meetup.UserFcmTokens[req.UserUID] = req.FCMToken
		}

		// Cập nhật lại meetup document
		err = updateFirestoreMeetup(meetup, []string{"pendingUids", "userFcmTokens"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi cập nhật meetup trên Firestore: " + err.Error()})
			return
		}

		// 3. Gửi push notification đến Host
		if meetup.HostFcmToken != "" {
			bodyText := fmt.Sprintf("%s muốn xin vào kèo \"%s\" chơi game %s của bạn.", req.UserName, meetup.Title, meetup.Game)
			_ = sendFCMNotification(meetup.HostFcmToken, "🎯 Yêu cầu tham gia kèo mới!", bodyText, "/?route=manage&meetupId="+req.MeetupID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Đã gửi yêu cầu tham gia kèo và thông báo tới Host!"})
	})

	// POST /api/meetups/approve
	mux.HandleFunc("/api/meetups/approve", func(w http.ResponseWriter, r *http.Request) {
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
			MeetupID  string `json:"meetupId"`
			PlayerUID string `json:"playerUid"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		if req.MeetupID == "" || req.PlayerUID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields"})
			return
		}

		// 1. Cập nhật status của request thành approved trên Firestore (người chơi vẫn chưa vào approvedUids)
		err := updateFirestoreRequestStatus(req.MeetupID, req.PlayerUID, "approved")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi cập nhật request status: " + err.Error()})
			return
		}

		// 2. Đọc meetup để chuyển sang approvedPendingUids
		meetup, err := getFirestoreMeetup(req.MeetupID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi đọc meetup: " + err.Error()})
			return
		}

		// Xóa khỏi pendingUids
		newPending := []string{}
		for _, uid := range meetup.PendingUids {
			if uid != req.PlayerUID {
				newPending = append(newPending, uid)
			}
		}
		meetup.PendingUids = newPending

		// Thêm vào approvedPendingUids
		found := false
		for _, uid := range meetup.ApprovedPendingUids {
			if uid == req.PlayerUID {
				found = true
				break
			}
		}
		if !found {
			meetup.ApprovedPendingUids = append(meetup.ApprovedPendingUids, req.PlayerUID)
		}

		err = updateFirestoreMeetup(meetup, []string{"approvedPendingUids", "pendingUids"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi lưu meetup: " + err.Error()})
			return
		}

		// 3. Gửi push notification đến Player báo được duyệt
		playerToken := meetup.UserFcmTokens[req.PlayerUID]
		if playerToken != "" {
			host := meetup.HostName
			if host == "" {
				host = "Host"
			}
			bodyText := fmt.Sprintf("Bạn đã được duyệt tham gia kèo \"%s\" chơi game %s của %s! Hãy bấm vào đây để xác nhận tham gia kèo chính thức.", meetup.Title, meetup.Game, host)
			_ = sendFCMNotification(playerToken, "🎉 Yêu cầu đã được duyệt!", bodyText, "/?route=manage&meetupId="+req.MeetupID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Đã duyệt thành viên và gửi thông báo!"})
	})

	// POST /api/meetups/confirm
	mux.HandleFunc("/api/meetups/confirm", func(w http.ResponseWriter, r *http.Request) {
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
			MeetupID string `json:"meetupId"`
			UserUID  string `json:"userUid"`
			UserName string `json:"userName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		if req.MeetupID == "" || req.UserUID == "" || req.UserName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields"})
			return
		}

		// 1. Lấy dữ liệu meetup từ Firestore
		meetup, err := getFirestoreMeetup(req.MeetupID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi đọc meetup: " + err.Error()})
			return
		}

		// 2. Chuyển người chơi từ approvedPendingUids sang approvedUids, tăng sĩ số
		// Xóa khỏi approvedPendingUids
		newAppPending := []string{}
		for _, uid := range meetup.ApprovedPendingUids {
			if uid != req.UserUID {
				newAppPending = append(newAppPending, uid)
			}
		}
		meetup.ApprovedPendingUids = newAppPending

		// Thêm vào approvedUids
		found := false
		for _, uid := range meetup.ApprovedUids {
			if uid == req.UserUID {
				found = true
				break
			}
		}
		if !found {
			meetup.ApprovedUids = append(meetup.ApprovedUids, req.UserUID)
			meetup.PlayersCount++
		}

		// Lưu đè meetup document lên Firestore
		err = updateFirestoreMeetup(meetup, []string{"approvedUids", "approvedPendingUids", "playersCount"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi lưu meetup: " + err.Error()})
			return
		}

		// 3. Gửi push notification chào mừng tới tất cả mọi người trong kèo (Approved members + Host)
		// Ngoại trừ chính người vừa confirm
		targets := make(map[string]bool)
		for _, uid := range meetup.ApprovedUids {
			if uid != req.UserUID {
				targets[uid] = true
			}
		}
		if meetup.HostUID != "" && meetup.HostUID != req.UserUID {
			targets[meetup.HostUID] = true
		}

		bodyText := fmt.Sprintf("%s đã xác nhận tham gia kèo \"%s\" chơi game %s.", req.UserName, meetup.Title, meetup.Game)
		for uid := range targets {
			token := meetup.UserFcmTokens[uid]
			if token == "" && uid == meetup.HostUID {
				token = meetup.HostFcmToken
			}
			if token != "" {
				_ = sendFCMNotification(token, "➕ Kèo có thêm người chơi mới!", bodyText, "/?route=manage&meetupId="+req.MeetupID)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Đã xác nhận vào kèo chính thức và thông báo tới mọi người!"})
	})

	// POST /api/meetups/leave
	mux.HandleFunc("/api/meetups/leave", func(w http.ResponseWriter, r *http.Request) {
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
			MeetupID   string `json:"meetupId"`
			PlayerUID  string `json:"playerUid"`
			PlayerName string `json:"playerName"`
			IsKick     bool   `json:"isKick"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		if req.MeetupID == "" || req.PlayerUID == "" || req.PlayerName == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields"})
			return
		}

		// 1. Xóa tài liệu request tương ứng
		_ = deleteFirestoreRequest(req.MeetupID, req.PlayerUID)

		// 2. Đọc meetup để cập nhật approvedUids, approvedPendingUids, pendingUids, và playersCount
		meetup, err := getFirestoreMeetup(req.MeetupID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi đọc meetup: " + err.Error()})
			return
		}

		wasApproved := false
		newApproved := []string{}
		for _, uid := range meetup.ApprovedUids {
			if uid != req.PlayerUID {
				newApproved = append(newApproved, uid)
			} else {
				wasApproved = true
			}
		}
		meetup.ApprovedUids = newApproved

		newPending := []string{}
		for _, uid := range meetup.PendingUids {
			if uid != req.PlayerUID {
				newPending = append(newPending, uid)
			}
		}
		meetup.PendingUids = newPending

		newAppPending := []string{}
		for _, uid := range meetup.ApprovedPendingUids {
			if uid != req.PlayerUID {
				newAppPending = append(newAppPending, uid)
			}
		}
		meetup.ApprovedPendingUids = newAppPending

		if wasApproved && meetup.PlayersCount > 1 {
			meetup.PlayersCount--
		}

		err = updateFirestoreMeetup(meetup, []string{"approvedUids", "approvedPendingUids", "pendingUids", "playersCount"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi cập nhật meetup trên Firestore: " + err.Error()})
			return
		}

		// 3. Gửi push notification
		if req.IsKick {
			// Host kick người chơi ➔ Gửi push riêng cho người bị kick
			kickToken := meetup.UserFcmTokens[req.PlayerUID]
			if kickToken != "" {
				bodyText := fmt.Sprintf("Host đã xóa bạn khỏi danh sách tham gia kèo \"%s\".", meetup.Title)
				_ = sendFCMNotification(kickToken, "✕ Bạn đã bị xóa khỏi kèo", bodyText, "/")
			}

			// Đồng thời thông báo cho những người còn lại trong kèo
			targets := make(map[string]bool)
			for _, uid := range meetup.ApprovedUids {
				if uid != req.PlayerUID {
					targets[uid] = true
				}
			}
			if meetup.HostUID != "" && meetup.HostUID != req.PlayerUID {
				targets[meetup.HostUID] = true
			}

			bodyText := fmt.Sprintf("%s đã không còn tham gia kèo \"%s\".", req.PlayerName, meetup.Title)
			for uid := range targets {
				token := meetup.UserFcmTokens[uid]
				if token == "" && uid == meetup.HostUID {
					token = meetup.HostFcmToken
				}
				if token != "" {
					_ = sendFCMNotification(token, "👋 Thành viên đã rời kèo", bodyText, "/?route=manage&meetupId="+req.MeetupID)
				}
			}
		} else {
			// Người chơi tự rời kèo ➔ Gửi push thông báo cho Host + những người còn lại
			targets := make(map[string]bool)
			for _, uid := range meetup.ApprovedUids {
				if uid != req.PlayerUID {
					targets[uid] = true
				}
			}
			if meetup.HostUID != "" && meetup.HostUID != req.PlayerUID {
				targets[meetup.HostUID] = true
			}

			bodyText := fmt.Sprintf("%s đã rời khỏi kèo \"%s\".", req.PlayerName, meetup.Title)
			for uid := range targets {
				token := meetup.UserFcmTokens[uid]
				if token == "" && uid == meetup.HostUID {
					token = meetup.HostFcmToken
				}
				if token != "" {
					_ = sendFCMNotification(token, "🚪 Thành viên rời kèo", bodyText, "/?route=manage&meetupId="+req.MeetupID)
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Đã xử lý rời kèo/kick và gửi thông báo thành công!"})
	})

	return mux
}
