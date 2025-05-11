package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type fakeStorage struct {
	positionsByBoat  map[string][]Position
	oldHeadingByBoat map[string]float64
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{
		positionsByBoat: map[string][]Position{
			"Bluebird": {
				{
					Latitude:  53.5675975,
					Longitude: 10.004,
					Time:      time.Now(),
				},
			},
			"Vivace": {
				{
					Latitude:  53.5675975,
					Longitude: 10.008,
					Time:      time.Now(),
				},
			},
		},
		oldHeadingByBoat: map[string]float64{
			"Bluebird": 0,
			"Vivace":   180,
		},
	}
}

type LastTwoPositions struct {
	CurrentPosition *Position
	LastPosition    *Position
}

type Position struct {
	Longitude float64   `json:"longitude"`
	Latitude  float64   `json:"latitude"`
	Time      time.Time `json:"time"`
}

// GetLastTwoPositions generates a new position for the boat based on the last
// position and a random heading and returns this position and the former
// position of the boat.
func (fs *fakeStorage) GetLastTwoPositions(_ context.Context, boat string, _ time.Time) (*LastTwoPositions, error) {
	if boat != "Bluebird" && boat != "Vivace" {
		return nil, fmt.Errorf("boat not found: %s", boat)
	}

	if len(fs.positionsByBoat[boat]) == 0 {
		return nil, fmt.Errorf("no positions found for boat: %s", boat)
	}

	lastPosition := fs.positionsByBoat[boat][len(fs.positionsByBoat[boat])-1]
	newHeading := fs.oldHeadingByBoat[boat] + 30*rand.Float64() - 10 // new heading is random but has a bias to the right
	fakeVelocity := 0.0005 * rand.Float64()
	newLatitude := lastPosition.Latitude + fakeVelocity*math.Cos(newHeading*math.Pi/180)
	newLongitude := lastPosition.Longitude + fakeVelocity*math.Sin(newHeading*math.Pi/180)

	currentPosition := Position{
		Latitude:  newLatitude,
		Longitude: newLongitude,
		Time:      time.Now(),
	}

	fs.positionsByBoat[boat] = append(fs.positionsByBoat[boat], currentPosition)
	fs.oldHeadingByBoat[boat] = newHeading

	return &LastTwoPositions{
		CurrentPosition: &currentPosition,
		LastPosition:    &lastPosition,
	}, nil
}

// GetPositions returns all positions of a boat in the given time range in
// ascending order.
func (fs *fakeStorage) GetPositions(_ context.Context, boat string, startTime, endTime time.Time) ([]Position, error) {
	if boat != "Bluebird" && boat != "Vivace" {
		return nil, fmt.Errorf("boat not found: %s", boat)
	}

	var positions []Position
	for _, pos := range fs.positionsByBoat[boat] {
		if pos.Time.Before(startTime) {
			continue
		} else if pos.Time.After(endTime) {
			break
		}
		positions = append(positions, pos)
	}

	return positions, nil
}

// InsertPositions does nothing in the fake storage
func (fs *fakeStorage) InsertPositions(_ context.Context, _ string, _ *DataServerReadMessageResponse) error {
	return nil
}
