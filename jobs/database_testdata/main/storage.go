package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type DatabaseClient struct {
	Database       *sql.DB
	Table          string
	defaultTimeout time.Duration
}

type DatabaseConfig struct {
	Host         string
	Port         int
	DatabaseName string
	UserName     string
	UserPassword string
}

func NewDatabaseClient(config DatabaseConfig, table string) (*DatabaseClient, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password='%s' dbname=%s sslmode=disable",
		config.Host, config.Port, config.UserName, config.UserPassword, config.DatabaseName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database 'regatta': %w", err)
	}
	return &DatabaseClient{
		Database:       db,
		Table:          table,
		defaultTimeout: time.Minute,
	}, nil
}

func (c *DatabaseClient) InsertPositions(ctx context.Context, position *PushMessageRequest) error {
	if position == nil {
		return errors.New("position is set to nil")
	}

	query := fmt.Sprintf(`INSERT INTO "%s"(boat, longitude, latitude, measure_time, send_time) VALUES ($1, $2, $3, $4, $5);`, c.Table)

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
