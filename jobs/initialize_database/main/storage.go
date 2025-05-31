package main

import (
	"context"
	"database/sql"
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

func (c *DatabaseClient) CreateBoatTable(ctx context.Context) error {
	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS boats (
            id text PRIMARY KEY,
            class text,
            yardstick pg_catalog.float8
        );
        INSERT INTO boats (id, class, yardstick) VALUES ('Bluebird', 'Conger', 118.0) ON CONFLICT DO NOTHING;
        INSERT INTO boats (id, class, yardstick) VALUES ('Vivace', 'Kielzugvogel', 108.0) ON CONFLICT DO NOTHING;
        INSERT INTO boats (id, class, yardstick) VALUES ('Polyflyer', 'H-Jolle Elb', 110.0) ON CONFLICT DO NOTHING;
    `)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	_, err := c.Database.ExecContext(ctx, query)
	return err
}

func (c *DatabaseClient) CreateRegattaTable(ctx context.Context) error {
	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS regattas (
            id text PRIMARY KEY,
            start_time timestamptz NOT NULL,
            end_time timestamptz NOT NULL
        );
        INSERT INTO regattas (id, start_time, end_time) VALUES ('ASV.24h.2024', '2024-08-03 13:00:00+02', '2024-08-04 14:00:00+02') ON CONFLICT DO NOTHING;
        INSERT INTO regattas (id, start_time, end_time) VALUES ('ASV.24h.2025', '2025-08-02 13:00:00+02', '2025-08-03 14:00:00+02') ON CONFLICT DO NOTHING;
        INSERT INTO regattas (id, start_time, end_time) VALUES ('Test',         '2025-05-31 00:00:00+00', '2025-08-01 00:00:00+00') ON CONFLICT DO NOTHING;
    `)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	_, err := c.Database.ExecContext(ctx, query)
	return err
}

func (c *DatabaseClient) CreateBuoyTable(ctx context.Context) error {
	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS buoys (
            id text NOT NULL,
            version int NOT NULL,
            latitude pg_catalog.float8 NOT NULL,
            longitude pg_catalog.float8 NOT NULL,
            pass_angle pg_catalog.float8 NOT NULL,
            is_pass_direction_clockwise boolean NOT NULL,
            start_time timestamptz NOT NULL,
            end_time timestamptz,
            PRIMARY KEY (id, version)
        );
        INSERT INTO buoys (id, version, latitude, longitude, pass_angle, is_pass_direction_clockwise, start_time) VALUES ('Schwanenwik bridge', 1, 53.565538, 10.009123,  90, true, '2024-01-01 00:00:00+02') ON CONFLICT DO NOTHING;
        INSERT INTO buoys (id, version, latitude, longitude, pass_angle, is_pass_direction_clockwise, start_time) VALUES ('Kennedy bridge',     1, 53.562266, 10.00422,  225, true, '2024-01-01 00:00:00+02') ON CONFLICT DO NOTHING;
        INSERT INTO buoys (id, version, latitude, longitude, pass_angle, is_pass_direction_clockwise, start_time) VALUES ('Langer Zug',         1, 53.575497, 10.005418,  45, true, '2024-01-01 00:00:00+02') ON CONFLICT DO NOTHING;
        INSERT INTO buoys (id, version, latitude, longitude, pass_angle, is_pass_direction_clockwise, start_time) VALUES ('Pier',               1, 53.577880, 10.008151, 180, true, '2024-01-01 00:00:00+02') ON CONFLICT DO NOTHING;
        `)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	_, err := c.Database.ExecContext(ctx, query)
	return err
}

func (c *DatabaseClient) CreateRoundTable(ctx context.Context) error {
	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS rounds (
            id int NOT NULL,
            regatta_id text NOT NULL,
            boat_id text NOT NULL,
            start_time timestamptz NOT NULL,
            end_time timestamptz,

            PRIMARY KEY (id, regatta_id, boat_id),

            CONSTRAINT fk_rounds_regatta
                FOREIGN KEY (regatta_id)
                REFERENCES regattas (id)
                ON DELETE RESTRICT
                ON UPDATE CASCADE,
                            
            CONSTRAINT fk_rounds_boat
                FOREIGN KEY (boat_id)
                REFERENCES boats (id)
                ON DELETE RESTRICT
                ON UPDATE CASCADE
        );
        `)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	_, err := c.Database.ExecContext(ctx, query)
	return err
}

func (c *DatabaseClient) CreateSectionTable(ctx context.Context) error {
	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS sections (
            id int NOT NULL,
            round_id int NOT NULL,
            regatta_id text NOT NULL,
            boat_id text NOT NULL,
            buoy_id_start text NOT NULL,
            buoy_version_start int NOT NULL,
            buoy_id_end text NOT NULL,
            buoy_version_end int NOT NULL,
            start_time timestamptz NOT NULL,
            end_time timestamptz,

            PRIMARY KEY (id, round_id, regatta_id, boat_id),

            CONSTRAINT fk_sections_round
                FOREIGN KEY (round_id, regatta_id, boat_id)
                REFERENCES rounds (id, regatta_id, boat_id)
                ON DELETE RESTRICT
                ON UPDATE CASCADE,

            CONSTRAINT fk_sections_regatta
                FOREIGN KEY (regatta_id)
                REFERENCES regattas (id)
                ON DELETE RESTRICT
                ON UPDATE CASCADE,

            CONSTRAINT fk_sections_boat
                FOREIGN KEY (boat_id)
                REFERENCES boats (id)
                ON DELETE RESTRICT
                ON UPDATE CASCADE,

            CONSTRAINT fk_sections_buoy_start
                FOREIGN KEY (buoy_id_start, buoy_version_start)
                REFERENCES buoys (id, version)
                ON DELETE RESTRICT
                ON UPDATE CASCADE,

            CONSTRAINT fk_sections_buoy_end
                FOREIGN KEY (buoy_id_end, buoy_version_end)
                REFERENCES buoys (id, version)
                ON DELETE RESTRICT
                ON UPDATE CASCADE
        );
        `)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	_, err := c.Database.ExecContext(ctx, query)
	return err
}

func (c *DatabaseClient) CreateGPSDataTable(ctx context.Context) error {
	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS gps_data (
            id bigserial PRIMARY KEY,
            regatta_id text,
            boat_id text NOT NULL,
            latitude pg_catalog.float8 NOT NULL,
            longitude pg_catalog.float8 NOT NULL,
            measure_time timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
            send_time timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
            receive_time timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
            distance pg_catalog.float8 NOT NULL DEFAULT 0,
            heading pg_catalog.float8 NOT NULL DEFAULT 0,
            velocity pg_catalog.float8 NOT NULL DEFAULT 0,

            CONSTRAINT fk_gps_data_regatta
                FOREIGN KEY (regatta_id)
                REFERENCES regattas (id)
                ON DELETE RESTRICT
                ON UPDATE CASCADE,

            CONSTRAINT fk_gps_data_boat
                FOREIGN KEY (boat_id)
                REFERENCES boats (id)
                ON DELETE RESTRICT
                ON UPDATE CASCADE
        );
        `)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	_, err := c.Database.ExecContext(ctx, query)
	return err
}
