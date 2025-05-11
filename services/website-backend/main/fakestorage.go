package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type fakeStorage struct {
	positionsByBoat  map[string][]PositionAtTime
	oldHeadingByBoat map[string]float64
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{
		positionsByBoat: map[string][]PositionAtTime{
			"Bluebird": {
				{
					Latitude:    53.5675975,
					Longitude:   10.004,
					MeasureTime: time.Now(),
					SendTime:    time.Now(),
					ReceiveTime: time.Now(),
				},
			},
			"Vivace": {
				{
					Latitude:    53.5675975,
					Longitude:   10.008,
					MeasureTime: time.Now(),
					SendTime:    time.Now(),
					ReceiveTime: time.Now(),
				},
			},
		},
		oldHeadingByBoat: map[string]float64{
			"Bluebird": 0,
			"Vivace":   180,
		},
	}
}

func (fs *fakeStorage) GetPositions(_ context.Context, boat string, _ time.Time, limit int) ([]PositionAtTime, error) {
	fmt.Println("Get fake position")

	if boat != "Bluebird" && boat != "Vivace" {
		return nil, fmt.Errorf("boat not found: %s", boat)
	}

	if limit > 2 {
		// return pearl chain

		numPositions := len(fs.positionsByBoat[boat])

		pearlLength := limit
		if numPositions <= limit {
			pearlLength = numPositions
		}

		return fs.positionsByBoat[boat][numPositions-pearlLength:], nil
	}

	// create and return new position
	lastPosition := fs.positionsByBoat[boat][len(fs.positionsByBoat[boat])-1]

	newHeading := fs.oldHeadingByBoat[boat] + 30*rand.Float64() - 10
	fakeVelocity := 0.0005 * rand.Float64()
	newLatitude := lastPosition.Latitude + fakeVelocity*math.Cos(newHeading*math.Pi/180)
	newLongitude := lastPosition.Longitude + fakeVelocity*math.Sin(newHeading*math.Pi/180)

	newPosition := &PositionAtTime{
		Latitude:    newLatitude,
		Longitude:   newLongitude,
		MeasureTime: time.Now(),
	}

	return []PositionAtTime{*newPosition}, nil
}

func (fs *fakeStorage) InsertPositions(_ context.Context, position *DataServerReadMessageResponse) error {
	return nil
}

func (fs *fakeStorage) GetMode() string {
	return "fake"
}
