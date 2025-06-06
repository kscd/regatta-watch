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
	gpsTable       string
	regattaTable   string
	buoyTable      string
	roundTable     string
	sectionTable   string
	boatTable      string
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

func newDatabaseClient(config databaseConfig) (*databaseClient, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password='%s' dbname=%s sslmode=disable",
		config.Host, config.Port, config.UserName, config.UserPassword, config.DatabaseName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database 'regatta': %w", err)
	}
	return &databaseClient{
		database:       db,
		defaultTimeout: time.Minute,
		gpsTable:       "gps_data",
		regattaTable:   "regattas",
		buoyTable:      "buoys",
		roundTable:     "rounds",
		sectionTable:   "sections",
		boatTable:      "boats",
	}, nil
}

// GetPositions returns all positions of a boat in the given time range in
// ascending order.
func (c *databaseClient) GetPositions(ctx context.Context, boat string, startTime, endTime time.Time) ([]Position, error) {
	query := fmt.Sprintf(`
		       SELECT latitude, longitude, measure_time, distance
			   FROM %s
			   WHERE boat_id = $1
			   AND measure_time > $2
			   AND measure_time <= $3
			   ORDER BY measure_time DESC
		       `, c.gpsTable)

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
			&position.Latitude,
			&position.Longitude,
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
		       SELECT regatta_id, latitude, longitude, measure_time, send_time, distance, heading, velocity
			   FROM %s
			   WHERE boat_id = $1
			   AND measure_time >= $2
			   AND measure_time <= $3
			   ORDER BY measure_time DESC
			   LIMIT 1;
		       `, c.gpsTable)

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	row := c.database.QueryRowContext(ctx, query, boat, lowerBound, upperBound)

	var position StoragePosition
	err := row.Scan(
		&position.RegattaID,
		&position.Latitude,
		&position.Longitude,
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
       INSERT INTO %s(regatta_id, boat_id, latitude, longitude, measure_time, send_time, distance, heading, velocity)
       VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
       `, c.gpsTable)

	queryWithoutRegatta := fmt.Sprintf(`
       INSERT INTO %s(boat_id, latitude, longitude, measure_time, send_time, distance, heading, velocity)
       VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
       `, c.gpsTable)

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

// GetRegattaAtTime returns the ID of the regatta that is active at the given time.
func (c *databaseClient) GetRegattaAtTime(ctx context.Context, time time.Time) (*string, error) {
	query := fmt.Sprintf(`
		SELECT id
		FROM %s
		WHERE start_time <= $1 AND end_time >= $1
		LIMIT 1;
	`, c.regattaTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	row := c.database.QueryRowContext(ctx, query, time)

	var regattaID string
	err := row.Scan(&regattaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan regatta ID: %w", err)
	}

	return &regattaID, nil
}

func (c *databaseClient) GetBuoysAtTime(ctx context.Context, time time.Time) ([]buoy, error) {
	query := fmt.Sprintf(`
		SELECT id, version, latitude, longitude, pass_angle, is_pass_direction_clockwise
		FROM %s
        WHERE id = ANY($1)
		AND start_time <= $2
        AND (end_time >= $2 OR end_time IS NULL)
		ORDER BY array_position($1, id);
	`, c.buoyTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	buoyIDList := []string{"Schwanenwik bridge", "Kennedy bridge", "Langer Zug", "Pier"}
	rows, err := c.database.QueryContext(ctx, query, buoyIDList, time)
	if err != nil {
		return nil, fmt.Errorf("query buoys: %w", err)
	}

	var buoys []buoy
	for rows.Next() {
		var buoy buoy
		err = rows.Scan(
			&buoy.ID,
			&buoy.Version,
			&buoy.Latitude,
			&buoy.Longitude,
			&buoy.PassAngle,
			&buoy.IsPassDirectionClockwise,
		)
		if err != nil {
			return nil, fmt.Errorf("parse row: %w", err)
		}
		buoy.ToleranceInMeters = 100
		buoys = append(buoys, buoy)
	}

	return buoys, nil
}

func (c *databaseClient) GetCurrentRound(ctx context.Context, regattaID, boatID string) (int, error) {
	query := fmt.Sprintf(`
		SELECT id
		FROM %s
        WHERE regatta_id = $1
		AND boat_id = $2
        AND end_time IS NULL
		LIMIT 1;
	`, c.roundTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	row := c.database.QueryRowContext(ctx, query, regattaID, boatID)
	var roundID int
	err := row.Scan(&roundID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("scan current round ID: %w", err)
	}

	return roundID, nil
}

func (c *databaseClient) GetCurrentSection(ctx context.Context, roundID int, regattaID, boatID string) (int, error) {
	query := fmt.Sprintf(`
		SELECT id
		FROM %s
        WHERE round_id = $1
		AND regatta_id = $2
        AND boat_id = $3
        AND end_time IS NULL
		LIMIT 1;
	`, c.sectionTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	row := c.database.QueryRowContext(ctx, query, roundID, regattaID, boatID)
	var sectionID int
	err := row.Scan(&sectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("scan current round ID: %w", err)
	}

	return sectionID, nil
}

func (c *databaseClient) GetLastCompletedRound(ctx context.Context, regattaID, boatID string) (int, error) {
	query := fmt.Sprintf(`
		SELECT id
		FROM %s
		WHERE regatta_id = $1
		AND boat_id = $2
		AND end_time IS NOT NULL
		ORDER BY end_time DESC
		LIMIT 1;
	`, c.roundTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	row := c.database.QueryRowContext(ctx, query, regattaID, boatID)
	var roundID int
	err := row.Scan(&roundID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("scan last completed round ID: %w", err)
	}

	return roundID, nil
}

func (c *databaseClient) GetLastCompletedSection(ctx context.Context, roundID int, regattaID, boatID string) (int, error) {
	query := fmt.Sprintf(`
		SELECT id
		FROM %s
		WHERE round_id = $1
		AND regatta_id = $2
		AND boat_id = $3
		AND end_time IS NOT NULL
		ORDER BY end_time DESC
		LIMIT 1;
	`, c.sectionTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	row := c.database.QueryRowContext(ctx, query, roundID, regattaID, boatID)
	var sectionID int
	err := row.Scan(&sectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("scan last completed round ID: %w", err)
	}

	return sectionID, nil
}

func (c *databaseClient) StartRound(ctx context.Context, roundID int, regattaID, boatID string, startTime time.Time) error {
	query := fmt.Sprintf(`
		INSERT INTO %s(id, regatta_id, boat_id, start_time)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id, regatta_id, boat_id) DO NOTHING;
	`, c.roundTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	_, err := c.database.ExecContext(ctx, query, roundID, regattaID, boatID, startTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("start round: %w", err)
		}
		return fmt.Errorf("start round: %w", err)
	}

	return nil
}

func (c *databaseClient) StartSection(ctx context.Context, sectionID, roundID int, regattaID, boatID string, startTime time.Time, buoyIdStart string, buoyVersionStart int, buoyIdEnd string, buoyVersionEnd int) error {
	query := fmt.Sprintf(`
		INSERT INTO %s(id, round_id, regatta_id, boat_id, start_time, buoy_id_start, buoy_version_start, buoy_id_end, buoy_version_end)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id, round_id, regatta_id, boat_id) DO NOTHING;
	`, c.sectionTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	_, err := c.database.ExecContext(ctx, query, sectionID, roundID, regattaID, boatID, startTime, buoyIdStart, buoyVersionStart, buoyIdEnd, buoyVersionEnd)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("start section: %w", err)
		}
		return fmt.Errorf("start section: %w", err)
	}

	return nil
}

func (c *databaseClient) EndRound(ctx context.Context, roundID int, regattaID, boatID string, endTime time.Time) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET end_time = $1
		WHERE id = $2
		AND regatta_id = $3
		AND boat_id = $4;
	`, c.roundTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	result, err := c.database.ExecContext(ctx, query, endTime, roundID, regattaID, boatID)
	if err != nil {
		return fmt.Errorf("end round: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("no rows updated")
	}

	return nil
}

func (c *databaseClient) EndSection(ctx context.Context, sectionID, roundID int, regattaID, boatID string, endTime time.Time) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET end_time = $1
		WHERE id = $2
		AND round_id = $3
		AND regatta_id = $4
		AND boat_id = $5;
	`, c.sectionTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	result, err := c.database.ExecContext(ctx, query, endTime, sectionID, roundID, regattaID, boatID)
	if err != nil {
		return fmt.Errorf("end section: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("no rows updated")
	}

	return nil
}

func (c *databaseClient) GetRoundsToTime(ctx context.Context, regattaID, boatID string, time time.Time) ([]Round, error) {
	query := fmt.Sprintf(`
		SELECT id, start_time, end_time
		FROM %s
		WHERE regatta_id = $1
		AND boat_id = $2
		AND start_time <= $3
		ORDER BY start_time ASC;
	`, c.roundTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	rows, err := c.database.QueryContext(ctx, query, regattaID, boatID, time)
	if err != nil {
		return nil, fmt.Errorf("query rounds: %w", err)
	}

	var rounds []Round
	for rows.Next() {
		var round Round
		err = rows.Scan(&round.ID, &round.StartTime, &round.EndTime)
		if err != nil {
			return nil, fmt.Errorf("parse row: %w", err)
		}
		rounds = append(rounds, round)
	}

	return rounds, nil
}

func (c *databaseClient) GetSectionsToTime(ctx context.Context, regattaID, boatID string, time time.Time) ([]Section, error) {
	query := fmt.Sprintf(`
		SELECT id, round_id, start_time, end_time
		FROM %s
		WHERE regatta_id = $1
		AND boat_id = $2
		AND start_time <= $3
		ORDER BY start_time ASC;
	`, c.sectionTable)

	ctx, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	rows, err := c.database.QueryContext(ctx, query, regattaID, boatID, time)
	if err != nil {
		return nil, fmt.Errorf("query sections: %w", err)
	}

	var sections []Section
	for rows.Next() {
		var section Section
		err = rows.Scan(&section.ID, &section.RoundID, &section.StartTime, &section.EndTime)
		if err != nil {
			return nil, fmt.Errorf("parse row: %w", err)
		}
		sections = append(sections, section)
	}

	return sections, nil
}
