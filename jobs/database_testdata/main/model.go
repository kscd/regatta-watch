package main

import "time"

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
