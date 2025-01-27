package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type DatabaseInterface interface {
	InsertPosition(ctx context.Context, position *PushMessageRequest) error
}

type databaseClient struct {
	Database       *sql.DB
	defaultTimeout time.Duration

	mode string // "normal", "test"
}

type databaseConfig struct {
	Host         string
	Port         int
	DatabaseName string
	UserName     string
	UserPassword string
}

func newDatabaseClient(config databaseConfig, mode string) (*databaseClient, error) {
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
	}, nil
}

func (c *databaseClient) InsertPositions(ctx context.Context, position *PushMessageRequest) error {
	if position == nil {
		return errors.New("position is set to nil")
	}

	query := `INSERT INTO "positions_data_server"(boat, longitude, latitude, measure_time, send_time) VALUES ($1, $2, $3, $4, $5);`

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	for i := range position.Positions {
		_, err := c.Database.ExecContext(
			ctx,
			query,
			position.Positions[i].Boat,
			position.Positions[i].Longitude,
			position.Positions[i].Latitude,
			position.Positions[i].MeasureTime,
			position.SendTime,
		)

		if err != nil {
			return fmt.Errorf("insert position: %w", err)
		}
	}

	return nil
}

func (c *databaseClient) GetPositions(ctx context.Context, boat string, start time.Time, end time.Time) ([]PositionAtTime, error) {
	table := "positions_data_server"
	startTime := start
	endTime := end

	if c.mode == "test" {
		table = "positions_data_server_test" // should have been positions_website_backend_test
		endTime = time.Date(2024, 01, 01, 00, 00, end.Second(), 0, time.UTC)
		if endTime.Second() < startTime.Second() {
			startTime = time.Date(2023, 12, 31, 23, 59, start.Second(), 0, time.UTC)
		} else {
			startTime = time.Date(2024, 01, 01, 00, 00, start.Second(), 0, time.UTC)
		}
	}

	query := fmt.Sprintf(`
       SELECT longitude, latitude, measure_time, send_time, receive_time
       FROM %s
       WHERE boat = $1
       AND measure_time > $2
       AND measure_time <= $3
       ORDER BY measure_time ASC;
       `, table)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	rows, err := c.Database.QueryContext(
		ctx,
		query,
		boat,
		startTime,
		endTime)

	if err != nil {
		return nil, fmt.Errorf("query position: %w", err)
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
		positions = append(positions, position)
	}

	if c.mode == "test" {
		for i := range positions {
			if positions[i].MeasureTime.Second() > end.Second() {
				positions[i].MeasureTime = time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), end.Minute(), positions[i].MeasureTime.Second(), 0, end.Location()).Add(-time.Minute)
			} else {
				positions[i].MeasureTime = time.Date(end.Year(), end.Month(), end.Day(), end.Hour(), end.Minute(), positions[i].MeasureTime.Second(), 0, end.Location())
			}
		}
	}

	if len(positions) == 0 {
		fmt.Println(start, startTime, end, endTime)
	}

	return positions, nil
}

func (c *databaseClient) InsertBatteryLevels(ctx context.Context, batteryMessage *BatteryMessage) error {
	if batteryMessage == nil {
		return errors.New("position is set to nil")
	}

	query := `INSERT INTO "battery"(batteryLevel, measure_time) VALUES ($1, $2);`

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	for i := range batteryMessage.BatteryLevel {
		_, err := c.Database.ExecContext(
			ctx,
			query,
			batteryMessage.BatteryLevel[i].BatteryLevel,
			batteryMessage.BatteryLevel[i].MeasureTime,
		)

		if err != nil {
			return fmt.Errorf("insert battery level: %w", err)
		}
	}

	return nil
}

func (c *databaseClient) ExtractBatteryLevel(ctx context.Context, start time.Time, end time.Time) (*BatteryMessage, error) {
	query := `SELECT batteryLevel, measure_time FROM battery WHERE measure_time >= $1 AND measure_time < $2;`

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	rows, err := c.Database.QueryContext(
		ctx,
		query,
		start,
		end)

	if err != nil {
		return nil, fmt.Errorf("query battery level: %w", err)
	}

	var batteryLevelSlice []BatteryLevel
	for rows.Next() {
		var batteryLevel BatteryLevel
		err = rows.Scan(
			&batteryLevel.BatteryLevel,
			&batteryLevel.MeasureTime,
		)
		if err != nil {
			return nil, fmt.Errorf("parse row: %w", err)
		}
		batteryLevelSlice = append(batteryLevelSlice, batteryLevel)
	}

	batteryMessage := BatteryMessage{
		BatteryLevel: batteryLevelSlice}

	return &batteryMessage, nil
}
