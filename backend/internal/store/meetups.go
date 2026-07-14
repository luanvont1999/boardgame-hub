package store

import "sync"

// Meetup represents a boardgame session
type Meetup struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Game          string  `json:"game"`
	HostName      string  `json:"host_name"`
	HostUID       string  `json:"host_uid"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	PlayersCount  int     `json:"players_count"`
	PlayersNeeded int     `json:"players_needed"`
	Time          string  `json:"time"`
	Color         string  `json:"color"`
}

// MeetupStore is a thread-safe in-memory database for meetups
type MeetupStore struct {
	sync.RWMutex
	meetups []Meetup
}

// NewMeetupStore initializes store with default values
func NewMeetupStore() *MeetupStore {
	return &MeetupStore{
		meetups: []Meetup{
			{
				ID:            "1",
				Title:         "Hội Ma Sói Đêm Trăng Q1",
				Game:          "Ultimate Werewolf",
				HostName:      "Minh Tuấn",
				HostUID:       "default-host-1",
				Lat:           10.7769,
				Lng:           106.7009,
				PlayersCount:  11,
				PlayersNeeded: 15,
				Time:          "2026-07-10T19:30",
				Color:         "#bca0f5", // purple
			},
			{
				ID:            "2",
				Title:         "Sân Chơi Mèo Nổ Q3",
				Game:          "Exploding Kittens",
				HostName:      "Thanh Trúc",
				HostUID:       "default-host-2",
				Lat:           10.7828,
				Lng:           106.6896,
				PlayersCount:  4,
				PlayersNeeded: 5,
				Time:          "2026-07-11T15:00",
				Color:         "#ffa4b2", // pink
			},
			{
				ID:            "3",
				Title:         "CLB Cờ Tỷ Phú Bình Thạnh",
				Game:          "Monopoly Deal",
				HostName:      "Khánh Huy",
				HostUID:       "default-host-3",
				Lat:           10.7981,
				Lng:           106.7051,
				PlayersCount:  3,
				PlayersNeeded: 6,
				Time:          "2026-07-12T18:00",
				Color:         "#ffe869", // yellow
			},
			{
				ID:            "4",
				Title:         "Chiến Thần Catan Hoàn Kiếm",
				Game:          "Settlers of Catan",
				HostName:      "Hoàng Lâm",
				HostUID:       "default-host-4",
				Lat:           21.0285,
				Lng:           105.8542,
				PlayersCount:  2,
				PlayersNeeded: 4,
				Time:          "2026-07-11T19:00",
				Color:         "#9ee3b2", // green
			},
			{
				ID:            "5",
				Title:         "Hội Avalon Tây Hồ",
				Game:          "Avalon",
				HostName:      "Thu Giang",
				HostUID:       "default-host-5",
				Lat:           21.0588,
				Lng:           105.8285,
				PlayersCount:  5,
				PlayersNeeded: 10,
				Time:          "2026-07-12T14:30",
				Color:         "#a4f0fd", // cyan
			},
		},
	}
}

// GetAll returns a copy of all meetups
func (s *MeetupStore) GetAll() []Meetup {
	s.RLock()
	defer s.RUnlock()
	
	// Create a copy to prevent external mutation
	result := make([]Meetup, len(s.meetups))
	copy(result, s.meetups)
	return result
}

// Add appends a new meetup to the store
func (s *MeetupStore) Add(meetup Meetup) {
	s.Lock()
	defer s.Unlock()
	s.meetups = append(s.meetups, meetup)
}
