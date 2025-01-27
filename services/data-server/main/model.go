package main

import "time"

type OwnTracksMessage struct {
	Battery   int     `json:"batt"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	CreatedAt int     `json:"created_at"`
}

type PushMessageRequest struct {
	Positions []Position `json:"positions"`
	SendTime  time.Time  `json:"send_time"`
}

type Position struct {
	Boat        string    `json:"boat"`
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
	MeasureTime time.Time `json:"measure_time"`
}

type ReadMessageRequest struct {
	Boat      string    `json:"boat"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type ReadMessageResponse struct {
	PositionsAtTime []PositionAtTime `json:"positions_at_time"`
}

type PositionAtTime struct {
	Longitude   float64   `json:"longitude"`
	Latitude    float64   `json:"latitude"`
	MeasureTime time.Time `json:"measure_time"`
	SendTime    time.Time `json:"send_time"`
	ReceiveTime time.Time `json:"receive_time"`
}

type BatteryMessage struct {
	BatteryLevel []BatteryLevel `json:"battery_level"`
}

type BatteryLevel struct {
	BatteryLevel float64   `json:"battery_level"`
	MeasureTime  time.Time `json:"measure_time"`
}
