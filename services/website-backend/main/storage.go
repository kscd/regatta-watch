package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type databaseClient struct {
	database       *sql.DB
	defaultTimeout time.Duration
	table          string
}

type databaseConfig struct {
	Host         string
	Port         int
	DatabaseName string
	UserName     string
	UserPassword string
}

type Position struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Distance  float64   `json:"distance"`
	Time      time.Time `json:"time"`
}

type StoragePosition struct {
	RegattaID   *string   `json:"regatta_id"`
	BoatID      string    `json:"boat_id"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Distance    float64   `json:"distance"`
	Heading     float64   `json:"heading"`
	Velocity    float64   `json:"velocity"`
	MeasureTime time.Time `json:"measure_time"`
	SendTime    time.Time `json:"send_time"`
}

func newDatabaseClient(config databaseConfig, table string) (*databaseClient, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password='%s' dbname=%s sslmode=disable",
		config.Host, config.Port, config.UserName, config.UserPassword, config.DatabaseName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database 'regatta': %w", err)
	}
	return &databaseClient{
		database:       db,
		defaultTimeout: time.Minute,
		table:          table,
	}, nil
}

// GetPositions returns all positions of a boat in the given time range in
// ascending order.
func (c *databaseClient) GetPositions(ctx context.Context, boat string, startTime, endTime time.Time) ([]Position, error) {
	query := fmt.Sprintf(`
		       SELECT longitude, latitude, measure_time, distance
			   FROM %s
			   WHERE boat_id = $1
			   AND measure_time > $2
			   AND measure_time <= $3
			   ORDER BY measure_time DESC
		       `, c.table)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	rows, err := c.database.QueryContext(ctx, query, boat, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}

	var positions []Position
	for rows.Next() {
		var position Position
		err = rows.Scan(
			&position.Longitude,
			&position.Latitude,
			&position.Time,
			&position.Distance,
		)
		if err != nil {
			return nil, fmt.Errorf("parse row: %w", err)
		}
		positions = append(positions, position)
	}

	return positions, nil
}

// GetLastPosition returns the last position of a boat before or equal
// to the upper bound time and after or equal to the lower bound time.
func (c *databaseClient) GetLastPosition(ctx context.Context, boat string, lowerBound, upperBound time.Time) (*StoragePosition, error) {
	query := fmt.Sprintf(`
		       SELECT longitude, latitude, measure_time, send_time, distance, heading, velocity
			   FROM %s
			   WHERE boat_id = $1
			   AND measure_time >= $2
			   AND measure_time <= $3
			   ORDER BY measure_time DESC
			   LIMIT 1;
		       `, c.table)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	row := c.database.QueryRowContext(ctx, query, boat, lowerBound, upperBound)

	var position StoragePosition
	err := row.Scan(
		&position.Longitude,
		&position.Latitude,
		&position.MeasureTime,
		&position.SendTime,
		&position.Distance,
		&position.Heading,
		&position.Velocity,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No positions found for the boat
		}
		return nil, fmt.Errorf("parse row: %w", err)
	}

	return &position, nil
}

// InsertPositions inserts a list of positions of a boat into the database.
func (c *databaseClient) InsertPositions(ctx context.Context, positions []StoragePosition) error {

	if positions == nil {
		return errors.New("position is set to nil")
	}

	queryWithRegatta := fmt.Sprintf(`
       INSERT INTO %s(regatta_id, boat_id, latitude, longitude, measure_time, send_time, distance, heading, velocity) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
       `, c.table)

	queryWithoutRegatta := fmt.Sprintf(`
       INSERT INTO %s(boat_id, latitude, longitude, measure_time, send_time, distance, heading, velocity) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
       `, c.table)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	var err error
	for _, position := range positions {

		if position.RegattaID == nil {
			_, err = c.database.ExecContext(
				ctx,
				queryWithoutRegatta,
				position.BoatID,
				position.Latitude,
				position.Longitude,
				position.MeasureTime,
				position.SendTime,
				position.Distance,
				position.Heading,
				position.Velocity,
			)
		} else {
			_, err = c.database.ExecContext(
				ctx,
				queryWithRegatta,
				*position.RegattaID,
				position.BoatID,
				position.Latitude,
				position.Longitude,
				position.MeasureTime,
				position.SendTime,
				position.Distance,
				position.Heading,
				position.Velocity,
			)
		}

		if err != nil {
			return fmt.Errorf("insert position: %w", err)
		}
	}

	return nil
}
