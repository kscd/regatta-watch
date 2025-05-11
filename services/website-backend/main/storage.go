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

// GetLastTwoPositions returns the last two positions of a boat.
func (c *databaseClient) GetLastTwoPositions(ctx context.Context, boat string, upperBound time.Time) (*LastTwoPositions, error) {
	query := fmt.Sprintf(`
		       SELECT longitude, latitude, measure_time
			   FROM %s
			   WHERE boat = $1
			   AND measure_time <= $2
			   ORDER BY measure_time DESC
			   LIMIT 2;
		       `, c.table)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	rows, err := c.database.QueryContext(ctx, query, boat, upperBound)
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
		)
		if err != nil {
			return nil, fmt.Errorf("parse row: %w", err)
		}
		positions = append(positions, position)
	}

	var currentPosition *Position
	if len(positions) >= 1 {
		currentPosition = &positions[0]
	}

	var lastPosition *Position
	if len(positions) >= 2 {
		lastPosition = &positions[1]
	}

	return &LastTwoPositions{
		CurrentPosition: currentPosition,
		LastPosition:    lastPosition,
	}, nil
}

// GetPositions returns all positions of a boat in the given time range in
// ascending order.
func (c *databaseClient) GetPositions(ctx context.Context, boat string, startTime, endTime time.Time) ([]Position, error) {
	query := fmt.Sprintf(`
		       SELECT longitude, latitude, measure_time
			   FROM %s
			   WHERE boat = $1
			   AND measure_time > $2
			   AND measure_time <= $3
			   ORDER BY measure_time ASC
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
		)
		if err != nil {
			return nil, fmt.Errorf("parse row: %w", err)
		}
		positions = append(positions, position)
	}

	return positions, nil
}

// InsertPositions inserts a list of positions of a boat into the database.
func (c *databaseClient) InsertPositions(ctx context.Context, boat string, position *DataServerReadMessageResponse) error {
	if position == nil {
		return errors.New("position is set to nil")
	}

	query := fmt.Sprintf(`
       INSERT INTO %s(boat, longitude, latitude, measure_time, send_time) VALUES ($1, $2, $3, $4, $5);
       `, c.table)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	for i := range position.PositionsAtTime {
		_, err := c.database.ExecContext(
			ctx,
			query,
			boat,
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
