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

func (fs *fakeStorage) InsertPositions(_ context.Context, _ []StoragePosition) error {
	return nil
}

func (fs *fakeStorage) GetLastPosition(_ context.Context, boat string, _, _ time.Time) (*StoragePosition, error) {
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

	return &StoragePosition{
		Longitude: lastPosition.Longitude,
		Latitude:  lastPosition.Latitude,
		Heading:   newHeading,
		Velocity:  fakeVelocity,
		Distance:  lastPosition.Distance,
	}, nil
}

func (fs *fakeStorage) GetRegattaAtTime(_ context.Context, _ time.Time) (*string, error) {
	testID := "test"
	return &testID, nil
}

func (fs *fakeStorage) GetBuoysAtTime(_ context.Context, _ time.Time) ([]buoy, error) {
	return []buoy{
		{ // buoy Schwanenwik bridge
			Latitude:                 53.565538,
			Longitude:                10.014123 - 0.005,
			PassAngle:                90,
			IsPassDirectionClockwise: true,
			ToleranceInMeters:        100,
		},
		{ // buoy Kennedy bridge
			Latitude:                 53.558766 + 0.0035,
			Longitude:                9.998720 + 0.0055,
			PassAngle:                225,
			IsPassDirectionClockwise: true,
			ToleranceInMeters:        100,
		},
		{ // buoy Langer Zug entry
			Latitude:                 53.576497 - 0.001,
			Longitude:                10.004418 + 0.001,
			PassAngle:                45,
			IsPassDirectionClockwise: true,
			ToleranceInMeters:        100,
		},
		{ // pier (placed north of the Langer Zug pointing down)
			Latitude:                 53.577880,
			Longitude:                10.008151,
			PassAngle:                160,
			IsPassDirectionClockwise: true,
			ToleranceInMeters:        100,
		},
	}, nil
}
