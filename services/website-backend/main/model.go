package main

import "time"

type FetchPositionRequest struct {
	Boat        string    `json:"boat"`
	NoLaterThan time.Time `json:"no_later_than"`
}

type FetchPositionResponse struct {
	MeasureTime time.Time `json:"measure_time"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Heading     float64   `json:"heading"`
	Distance    float64   `json:"distance"`
	Velocity    float64   `json:"velocity"`
	Round       int       `json:"round"`
	Section     int       `json:"section"`
	Crew0       string    `json:"crew0"`
	Crew1       string    `json:"crew1"`
	NextCrew0   string    `json:"next_crew0"`
	NextCrew1   string    `json:"next_crew1"`
}

type FetchPearlChainRequest struct {
	Boat        string    `json:"boat"`
	NoLaterThan time.Time `json:"no_later_than"`
	Duration    int       `json:"duration"` // in seconds
	Step        int       `json:"step"`     // in seconds
}

type position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Heading   float64 `json:"heading"`
}

type FetchPearlChainResponse struct {
	Positions []position `json:"positions"`
}

type buoy struct {
	Latitude                 float64 `json:"latitude"`
	Longitude                float64 `json:"longitude"`
	PassAngle                float64 `json:"pass_angle"`
	IsPassDirectionClockwise bool    `json:"is_pass_direction_clockwise"`
	ToleranceInMeters        float64 `json:"tolerance"`
}

type PositionAtTime struct {
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
	MeasureTime time.Time `json:"measure_time"`
	SendTime    time.Time `json:"send_time"`
	ReceiveTime time.Time `json:"receive_time"`
}

type FetchRoundTimeRequest struct {
	Boat string `json:"boat"`
}

type FetchRoundTimeResponse struct {
	RoundTimes   []float64 `json:"round_times"`
	SectionTimes []float64 `json:"section_times"`
}

type DataServerReadMessageResponse struct {
	PositionsAtTime []PositionAtTime `json:"positions_at_time"`
}

type ReadMessageRequest struct {
	Boat      string    `json:"boat"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

var roundToCrew = map[int][]string{
	0:  {"Heiko", "Gabriel"},
	1:  {"Gabriel", "Jana"},
	2:  {"Jana", "Birgitt"},
	3:  {"Birgitt", "Kevin"},
	4:  {"Kevin", "Michael"},
	5:  {"Michael", "Raymund"},
	6:  {"Raymund", "Liz"},
	7:  {"Dirk", "Liz"},
	8:  {"Heiko", "Dirk"},
	9:  {"Gabriel", "Birgitt"},
	10: {"Jana", "Kevin"},
	11: {"Raymund", "Jana"},
	12: {"Heiko", "Liz"},
	13: {"Birgitt", "Michael"},
	14: {"Dirk", "Gabriel"},
	15: {"Kevin", "Raymund"},
	16: {"Heiko", "Birgitt"},
	17: {"Dirk", "Michael"},
	18: {"Gabriel", "Liz"},
	19: {"Michael", "Jana"},
	20: {"Dirk", "Kevin"},
	21: {"Raymund", "Heiko"},
	22: {"Birgitt", "Liz"},
	23: {"Kevin", "Gabriel"},
}
