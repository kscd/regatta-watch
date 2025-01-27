package main

import (
	"context"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	c, err := loadConfig()
	if err != nil {
		log.Fatal("error loading config: ", err)
	}

	dbClient, err := NewDatabaseClient(c.DBConfig, "positions_data_server_test")
	if err != nil {
		log.Fatal(err)
	}

	cornerPositions := []Position{
		{ // North East, actually the last, just here to offer historical data
			Boat:        "Bluebird",
			Latitude:    53.576946,
			Longitude:   10.001763,
			MeasureTime: time.Date(2023, time.December, 31, 23, 59, 54, 0, time.UTC),
		},
		{ // Bellevue street
			Boat:        "Bluebird",
			Latitude:    53.577113,
			Longitude:   10.004230,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{ // middle of Langer Zug entry
			Boat:        "Bluebird",
			Latitude:    53.576092,
			Longitude:   10.005250,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 4, 0, time.UTC),
		},
		{ // pier north
			Boat:        "Bluebird",
			Latitude:    53.577342,
			Longitude:   10.009071,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 8, 0, time.UTC),
		},
		{ // pier south
			Boat:        "Bluebird",
			Latitude:    53.576996,
			Longitude:   10.009439,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 10, 0, time.UTC),
		},
		{ // middle of Langer Zug entry
			Boat:        "Bluebird",
			Latitude:    53.576092,
			Longitude:   10.005250,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 14, 0, time.UTC),
		},
		{ // Hansa Steg
			Boat:        "Bluebird",
			Latitude:    53.568951,
			Longitude:   10.009498,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 20, 0, time.UTC),
		},
		{ // Schwanenwik bridge
			Boat:        "Bluebird",
			Latitude:    53.565728,
			Longitude:   10.015460,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 24, 0, time.UTC),
		},
		{ // Atlantic hotel
			Boat:        "Bluebird",
			Latitude:    53.558063,
			Longitude:   10.002398,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 30, 0, time.UTC),
		},
		{ // Bottomleft corner
			Boat:        "Bluebird",
			Latitude:    53.558629,
			Longitude:   9.996797,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 36, 0, time.UTC),
		},
		{ // Alsterufer
			Boat:        "Bluebird",
			Latitude:    53.565851,
			Longitude:   10.002505,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 42, 0, time.UTC),
		},
		{ // Alsterpark
			Boat:        "Bluebird",
			Latitude:    53.573737,
			Longitude:   10.003055,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 48, 0, time.UTC),
		},
		{ // North East
			Boat:        "Bluebird",
			Latitude:    53.576946,
			Longitude:   10.001763,
			MeasureTime: time.Date(2024, time.January, 1, 0, 0, 54, 0, time.UTC),
		},
		{ // Bellevue street
			Boat:        "Bluebird",
			Latitude:    53.577113,
			Longitude:   10.004230,
			MeasureTime: time.Date(2024, time.January, 1, 0, 1, 0, 0, time.UTC),
		},
	}

	var positions []Position

	for i := 0; i < len(cornerPositions)-1; i++ {
		timeDiff := int(cornerPositions[i+1].MeasureTime.Unix() - cornerPositions[i].MeasureTime.Unix())
		deltaLon := (cornerPositions[i+1].Longitude - cornerPositions[i].Longitude) / float64(timeDiff)
		deltaLat := (cornerPositions[i+1].Latitude - cornerPositions[i].Latitude) / float64(timeDiff)

		for j := 0; j < timeDiff; j++ {
			position := Position{
				Boat:        cornerPositions[i].Boat,
				Longitude:   cornerPositions[i].Longitude + float64(j)*deltaLon,
				Latitude:    cornerPositions[i].Latitude + float64(j)*deltaLat,
				MeasureTime: cornerPositions[i].MeasureTime.Add(time.Duration(j) * time.Second),
			}
			positions = append(positions, position)
		}
	}

	pushMessageRequest := &PushMessageRequest{
		Positions: positions,
		SendTime:  time.Now(),
	}

	err = dbClient.InsertPositions(context.Background(), pushMessageRequest)
	if err != nil {
		log.Fatal(err)
	}
}
