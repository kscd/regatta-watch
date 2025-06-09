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

	ctx := context.Background()

	dbClient, err := NewDatabaseClient(c.DBConfig, "positions_data_server_test")
	if err != nil {
		log.Fatal(err)
	}

	var positions []Position

	for i := 0; i < len(cornerPositionsBluebird)-1; i++ {
		timeDiff := int(cornerPositionsBluebird[i+1].MeasureTime.Unix() - cornerPositionsBluebird[i].MeasureTime.Unix())
		deltaLon := (cornerPositionsBluebird[i+1].Longitude - cornerPositionsBluebird[i].Longitude) / float64(timeDiff)
		deltaLat := (cornerPositionsBluebird[i+1].Latitude - cornerPositionsBluebird[i].Latitude) / float64(timeDiff)

		for j := 0; j < timeDiff; j++ {
			position := Position{
				Boat:        cornerPositionsBluebird[i].Boat,
				Longitude:   cornerPositionsBluebird[i].Longitude + float64(j)*deltaLon,
				Latitude:    cornerPositionsBluebird[i].Latitude + float64(j)*deltaLat,
				MeasureTime: cornerPositionsBluebird[i].MeasureTime.Add(time.Duration(j) * time.Second),
			}
			positions = append(positions, position)
		}
	}

	for i := 0; i < len(cornerPositionsVivace)-1; i++ {
		timeDiff := int(cornerPositionsVivace[i+1].MeasureTime.Unix() - cornerPositionsVivace[i].MeasureTime.Unix())
		deltaLon := (cornerPositionsVivace[i+1].Longitude - cornerPositionsVivace[i].Longitude) / float64(timeDiff)
		deltaLat := (cornerPositionsVivace[i+1].Latitude - cornerPositionsVivace[i].Latitude) / float64(timeDiff)

		for j := 0; j < timeDiff; j++ {
			position := Position{
				Boat:        cornerPositionsVivace[i].Boat,
				Longitude:   cornerPositionsVivace[i].Longitude + float64(j)*deltaLon,
				Latitude:    cornerPositionsVivace[i].Latitude + float64(j)*deltaLat,
				MeasureTime: cornerPositionsVivace[i].MeasureTime.Add(time.Duration(j) * time.Second),
			}
			positions = append(positions, position)
		}
	}

	pushMessageRequest := &PushMessageRequest{
		Positions: positions,
		SendTime:  time.Now(),
	}

	err = dbClient.TruncateTable(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = dbClient.InsertPositions(ctx, pushMessageRequest)
	if err != nil {
		log.Fatal(err)
	}
}
