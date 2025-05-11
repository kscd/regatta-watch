package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

type databaseClient struct {
	Database       *sql.DB
	defaultTimeout time.Duration

	mode string // "normal", "test", "random", "hackertalk"

	// for non-DB test data
	positions  []PositionAtTime
	oldHeading float64
}

type databaseConfig struct {
	Host         string
	Port         int
	DatabaseName string
	UserName     string
	UserPassword string
}

func newDatabaseClient(config databaseConfig, initialPosition PositionAtTime, mode string) (*databaseClient, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password='%s' dbname=%s sslmode=disable",
		config.Host, config.Port, config.UserName, config.UserPassword, config.DatabaseName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database 'regatta': %w", err)
	}
	return &databaseClient{
		Database:       db,
		defaultTimeout: time.Minute,
		mode:           mode,
		// for non-DB test data
		positions: []PositionAtTime{initialPosition},
	}, nil
}

// GetPositions positions are sorted descending
func (c *databaseClient) GetPositions(ctx context.Context, boat string, upperBound time.Time, limit int) ([]PositionAtTime, error) {
	upperBoundFake := upperBound
	table := "positions_website_backend"
	if c.mode == "test" {
		table = "positions_data_server_test" // should have been positions_website_backend_test
		upperBoundFake = time.Date(2024, 01, 01, 00, 00, upperBound.Second(), 0, time.UTC)
	}

	query := fmt.Sprintf(`
       SELECT longitude, latitude, measure_time, send_time, receive_time
	   FROM %s
	   WHERE boat = $1
	   AND measure_time <= $2
	   ORDER BY measure_time ASC
	   LIMIT $3;
       `, table)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	rows, err := c.Database.QueryContext(
		ctx,
		query,
		boat,
		upperBoundFake,
		limit)

	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}

	var positions []PositionAtTime
	for rows.Next() {
		var position PositionAtTime
		err = rows.Scan(
			&position.Longitude,
			&position.Latitude,
			&position.MeasureTime,
			&position.SendTime,
			&position.ReceiveTime,
		)
		if err != nil {
			return nil, fmt.Errorf("parse row: %w", err)
		}

		if c.mode == "test" {
			position.MeasureTime = time.Date(
				upperBound.Year(),
				upperBound.Month(),
				upperBound.Day(),
				upperBound.Hour(),
				upperBound.Minute(),
				position.MeasureTime.Second(),
				position.MeasureTime.Nanosecond(),
				position.MeasureTime.Location())
			position.MeasureTime.Second()
		}
		positions = append(positions, position)
	}
	return positions, nil
}

func (c *databaseClient) InsertPositions(ctx context.Context, position *DataServerReadMessageResponse) error {
	if position == nil {
		return errors.New("position is set to nil")
	}

	query := fmt.Sprintf(`
       INSERT INTO %s(boat, longitude, latitude, measure_time, send_time) VALUES ($1, $2, $3, $4, $5);
       `, "positions_website_backend")

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	for i := range position.PositionsAtTime {
		_, err := c.Database.ExecContext(
			ctx,
			query,
			"Bluebird",
			position.PositionsAtTime[i].Longitude,
			position.PositionsAtTime[i].Latitude,
			position.PositionsAtTime[i].MeasureTime,
			position.PositionsAtTime[i].SendTime,
		)

		if err != nil {
			return fmt.Errorf("insert position: %w", err)
		}
	}

	return nil
}

func (c *databaseClient) GetMode() string {
	return c.mode
}

// 0 degrees are north and the heading rotates clockwise
func calculateHeading(oldLatitude, oldLongitude, newLatitude, newLongitude float64) float64 {
	oldLatitudeRadians := oldLatitude * math.Pi / 180
	oldLongitudeRadians := oldLongitude * math.Pi / 180
	newLatitudeRadians := newLatitude * math.Pi / 180
	newLongitudeRadians := newLongitude * math.Pi / 180

	X := math.Cos(newLatitudeRadians) * math.Sin(newLongitudeRadians-oldLongitudeRadians)
	Y := math.Cos(oldLatitudeRadians)*math.Sin(newLatitudeRadians) - math.Sin(oldLatitudeRadians)*math.Cos(newLatitudeRadians)*math.Cos(newLongitudeRadians-oldLongitudeRadians)
	return math.Atan2(X, Y) * 180 / math.Pi
}

// uses euclidean geometry
func calculateDistanceInNM(oldLatitude, oldLongitude, newLatitude, newLongitude float64) float64 {
	cosLatitude := (math.Cos(newLatitude) + math.Cos(oldLatitude)) / 2
	deltaN := (newLatitude - oldLatitude) * nauticalMilesPerDegree
	deltaW := (newLongitude - oldLongitude) * nauticalMilesPerDegree * cosLatitude
	return math.Sqrt(deltaN*deltaN + deltaW*deltaW) // nautical miles
}

func calculateIfBuoysPassed(oldPosition, newPosition *PositionAtTime) ([]bool, error) {
	var isPassed []bool
	for i := range buoys {
		lat1, lon1 := calculateNewPosition(buoys[i].Latitude, buoys[i].Longitude, buoys[i].PassAngle+180, buoys[i].ToleranceInMeters)
		lat2, lon2 := calculateNewPosition(buoys[i].Latitude, buoys[i].Longitude, buoys[i].PassAngle, buoyFarOffPointDistance)

		buoyLS := newLineSegment(lat1, lon1, lat2, lon2)
		boatLS := newLineSegment(oldPosition.Latitude, oldPosition.Longitude, newPosition.Latitude, newPosition.Longitude)

		isBuoyPassed, err := isIntersecting(buoyLS, boatLS)
		if err != nil {
			return nil, fmt.Errorf("error checking if buoys passed: %v", err)
		}
		if !isBuoyPassed {
			isPassed = append(isPassed, false)
			continue
		}
		isBuoyPassed = isPassDirectionCorrect(
			lat1, lon1,
			lat2, lon2,
			oldPosition.Latitude, oldPosition.Longitude,
			newPosition.Latitude, newPosition.Longitude,
			buoys[i].IsPassDirectionClockwise,
		)
		isPassed = append(isPassed, isBuoyPassed)
	}
	return isPassed, nil
}

func isPassDirectionCorrect(buoyLat1, buoyLon1, buoyLat2, buoyLon2, boatLat1, boatLon1, boatLat2, boatLon2 float64, isPassDirectionClockwise bool) bool {
	// For the three vectors buoy to buoy far off (vec a), buoy to boat old
	// (vec b), buoy to boat new (vec c) calculate the polar angle and do
	// some comparisons to figure out which way the boat rotated around the
	// buoy.

	vecAY := boatLat1 - buoyLat1
	vecAX := (boatLon1 - buoyLon1) * math.Cos((boatLat1+buoyLat1)/2/180*math.Pi) // compensate for non-square coordinate system
	vecBY := buoyLat2 - buoyLat1
	vecBX := (buoyLon2 - buoyLon1) * math.Cos((buoyLat2+buoyLat1)/2/180*math.Pi)
	vecCY := boatLat2 - buoyLat1
	vecCX := (boatLon2 - buoyLon1) * math.Cos((boatLat2+buoyLat1)/2/180*math.Pi)

	alpha := math.Atan2(vecAY, vecAX) * 180 / math.Pi
	beta := math.Atan2(vecBY, vecBX) * 180 / math.Pi
	gamma := math.Atan2(vecCY, vecCX) * 180 / math.Pi
	alpha = math.Mod(alpha-beta, 360)
	gamma = math.Mod(gamma-beta, 360)

	if alpha >= 0 && gamma <= 0 {
		// clockwise rotation
		if isPassDirectionClockwise {
			return true
		}
		return false
	}

	if alpha <= 0 && gamma >= 0 {
		// anti-clockwise rotation
		if isPassDirectionClockwise {
			return false
		}
		return true
	}

	// something went wrong, there was never an intersection, return false
	return false
}
