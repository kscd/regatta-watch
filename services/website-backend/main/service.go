package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	nauticalMilesPerDegree  = 60
	buoyFarOffPointDistance = 1000
)

type regattaService struct {
	storageClient    storageInterface
	httpClient       *http.Client
	dataServerURL    string
	regattaStartTime time.Time
	regattaEndTime   time.Time
	clock            clockInterface
}

type clockInterface interface {
	Now() time.Time
	RealNow() time.Time
	SetCurrentTimeAs(fakeTime time.Time)
	SetSpeed(speed float64)
	Reset()
}

type storageInterface interface {
	InsertPositions(ctx context.Context, position []StoragePosition) error
	GetLastPosition(ctx context.Context, boat string, lowerBound, upperBound time.Time) (*StoragePosition, error)
	GetPositions(ctx context.Context, boat string, startTime, endTime time.Time) ([]Position, error)
	GetRegattaAtTime(ctx context.Context, time time.Time) (*string, error)
	GetBuoysAtTime(ctx context.Context, time time.Time) ([]buoy, error)
	GetCurrentRound(ctx context.Context, regattaID, boatID string) (int, error)
	GetCurrentSection(ctx context.Context, roundID int, regattaID, boatID string) (int, error)
	GetLastCompletedRound(ctx context.Context, regattaID, boatID string) (int, error)
	GetLastCompletedSection(ctx context.Context, roundID int, regattaID, boatID string) (int, error)
	StartRound(ctx context.Context, roundID int, regattaID, boatID string, startTime time.Time) error
	StartSection(ctx context.Context, sectionID, roundID int, regattaID, boatID string, startTime time.Time, buoyIdStart string, buoyVersionStart int, buoyIdEnd string, buoyVersionEnd int) error
	EndRound(ctx context.Context, roundID int, regattaID, boatID string, endTime time.Time) error
	EndSection(ctx context.Context, sectionID, roundID int, regattaID, boatID string, endTime time.Time) error
	GetRoundsToTime(ctx context.Context, regattaID, boatID string, time time.Time) ([]Round, error)
	GetSectionsToTime(ctx context.Context, regattaID, boatID string, time time.Time) ([]Section, error)
}

func newRegattaService(
	storageClient storageInterface,
	dataServerURL string,
	regattaStartTime time.Time,
	regattaEndTime time.Time,
	httpClient *http.Client) *regattaService {
	return &regattaService{
		storageClient:    storageClient,
		dataServerURL:    dataServerURL,
		httpClient:       httpClient,
		regattaStartTime: regattaStartTime,
		regattaEndTime:   regattaEndTime,
		clock:            newClock(),
	}
}

func (s *regattaService) LogError(err error) {
	log.Println(err.Error())
}

func (s *regattaService) LogDebug(message string) {
	log.Println(message)
}

func (s *regattaService) Ping(w http.ResponseWriter, _ *http.Request) {
	enableCors(&w)

	// fmt.Println("/ping called")
	if _, err := w.Write([]byte("pong")); err != nil {
		err = fmt.Errorf("ping: write to http response writer: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

func (s *regattaService) FetchPosition(w http.ResponseWriter, r *http.Request) {
	fmt.Println("FetchPositions called")

	enableCors(&w)

	ctx := r.Context()

	// reset clock if we move into the future
	now := s.clock.Now()
	realNow := s.clock.RealNow()
	if now.After(realNow) {
		s.clock.Reset()
	}

	// parse data from request
	var m FetchPositionRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("read position: read http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &m); err != nil {
		err = fmt.Errorf("read position: unmarshal http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	position, err := s.storageClient.GetLastPosition(ctx, m.Boat, s.regattaStartTime, now)
	if err != nil {
		s.LogError(fmt.Errorf("get positions: %v", err))
		return
	}

	if position == nil {
		return
	}

	var round int
	var section int
	if position.RegattaID != nil {
		round, err = s.storageClient.GetCurrentRound(ctx, *position.RegattaID, m.Boat)
		if err != nil {
			s.LogError(fmt.Errorf("get current round: %v", err))
			return
		}
		section, err = s.storageClient.GetCurrentSection(ctx, round, *position.RegattaID, m.Boat)
		if err != nil {
			s.LogError(fmt.Errorf("get current section: %v", err))
			return
		}
	}

	var crew []string
	var nextCrew []string
	if round == 0 {
		crew = []string{"?", "?"}
		nextCrew = []string{"?", "?"}
	} else {
		var ok bool
		crew, ok = roundToCrew[round-1]
		if !ok {
			crew = []string{"?", "?"}
		}
		nextCrew, ok = roundToCrew[round]
		if !ok {
			nextCrew = []string{"?", "?"}
		}
	}

	response := FetchPositionResponse{
		MeasureTime: position.MeasureTime,
		Latitude:    position.Latitude,
		Longitude:   position.Longitude,
		Heading:     position.Heading,
		Distance:    position.Distance,
		Velocity:    position.Velocity,
		Round:       round,
		Section:     section,
		Crew0:       crew[0],
		Crew1:       crew[1],
		NextCrew0:   nextCrew[0],
		NextCrew1:   nextCrew[1],
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		err = fmt.Errorf("read position: marshal response: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	_, err = w.Write(responseBytes)
	if err != nil {
		err = fmt.Errorf("read position: write to http writer: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

func (s *regattaService) FetchPearlChain(w http.ResponseWriter, r *http.Request) {
	fmt.Println("FetchPearlChain called")

	enableCors(&w)

	ctx := r.Context()

	// parse data from request
	var m FetchPearlChainRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("read pearl chain: read http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &m); err != nil {
		err = fmt.Errorf("read position: unmarshal http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if m.Length <= 0 || m.Interval <= 0 {
		return
	}

	pearlChainTime := time.Duration(m.Length) * time.Second // time.Duration is needed for type matching

	endTime := s.clock.Now()
	startTime := endTime.Add(-pearlChainTime)

	// positions sorted in descending order
	positions, err := s.storageClient.GetPositions(ctx, m.Boat, startTime, endTime)
	if err != nil {
		fmt.Printf("/fetchposition get positions: %v", err)
		return
	}

	var response FetchPearlChainResponse
	if len(positions) >= 2 {

		// TODO: Also add a mode for static chain links. Can do this with a bit of book keeping by manipulating the
		// initial value of nextStop.

		// Calculate the time step for the pearl chain from database data
		pearlChainStep := time.Duration(m.Interval) * time.Second
		nextStop := endTime.Add(-pearlChainStep)

		var pearlChain []position
		for index := range positions {
			if positions[index].Time.Sub(nextStop) < 0 && index+1 < len(positions)-1 {
				pearlChain = append(pearlChain, position{
					Latitude:  positions[index].Latitude,
					Longitude: positions[index].Longitude,
					Heading: calculateHeading(
						positions[index+1].Latitude,
						positions[index+1].Longitude,
						positions[index].Latitude,
						positions[index].Longitude),
				})
				nextStop = nextStop.Add(-pearlChainStep)
			}
		}

		response = FetchPearlChainResponse{Positions: pearlChain}
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		err = fmt.Errorf("read position: marshal response: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	_, err = w.Write(responseBytes)
	if err != nil {
		err = fmt.Errorf("read position: write to http writer: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

func (s *regattaService) FetchRoundTimes(w http.ResponseWriter, r *http.Request) {
	fmt.Println("FetchRoundTimes called")

	enableCors(&w)

	ctx := r.Context()

	var m FetchRoundTimeRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("fetch round now: read http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &m); err != nil {
		err = fmt.Errorf("fetch round now: unmarshal http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	now := s.clock.Now()

	regattaID, err := s.storageClient.GetRegattaAtTime(ctx, now)
	if err != nil {
		return
	}

	var roundTimeCurrent []float64
	var sectionTimeCurrent []float64
	if regattaID != nil {

		rounds, err := s.storageClient.GetRoundsToTime(ctx, *regattaID, m.Boat, now)
		if err != nil {
			err = fmt.Errorf("fetch round times: get rounds to now: %w", err)
			s.LogError(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		sections, err := s.storageClient.GetSectionsToTime(ctx, *regattaID, m.Boat, now)
		if err != nil {
			err = fmt.Errorf("fetch round times: get rounds to now: %w", err)
			s.LogError(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for _, round := range rounds {
			if round.EndTime == nil || now.Sub(*round.EndTime) < 0 {
				roundTimeCurrent = append(roundTimeCurrent, now.Sub(round.StartTime).Seconds())
			} else {
				roundTimeCurrent = append(roundTimeCurrent, round.EndTime.Sub(round.StartTime).Seconds())
			}
		}

		for _, section := range sections {
			if section.EndTime == nil || now.Sub(*section.EndTime) < 0 {
				sectionTimeCurrent = append(sectionTimeCurrent, now.Sub(section.StartTime).Seconds())
			} else {
				sectionTimeCurrent = append(sectionTimeCurrent, section.EndTime.Sub(section.StartTime).Seconds())
			}
		}
	}

	response := FetchRoundTimeResponse{
		RoundTimes:   roundTimeCurrent,
		SectionTimes: sectionTimeCurrent,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		err = fmt.Errorf("read round times: marshal response: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	_, err = w.Write(responseBytes)
	if err != nil {
		err = fmt.Errorf("read round times: write to http writer: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

}

func (s *regattaService) SetClockConfiguration(w http.ResponseWriter, r *http.Request) {
	fmt.Println("SetClockConfiguration called")

	enableCors(&w)

	// parse data from request
	var c SetClockConfigurationRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("set clock configuration: read http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &c); err != nil {
		err = fmt.Errorf("set clock configuration: unmarshal http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	s.clock.SetCurrentTimeAs(c.ClockTime)
	s.clock.SetSpeed(c.ClockSpeed)

	return
}

func (s *regattaService) ResetClockConfiguration(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("ResetClockConfiguration called")

	enableCors(&w)

	s.clock.Reset()

	return
}

func (s *regattaService) GetClockTime(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("GetTime called")

	enableCors(&w)

	currentTime := s.clock.Now()

	response := GetClockTimeResponse{
		Time: currentTime,
	}

	// Encode response to JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		err = fmt.Errorf("get time: marshal response: %w", err)
		s.LogError(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Write response
	if _, err = w.Write(responseBytes); err != nil {
		err = fmt.Errorf("get time: write to http writer: %w", err)
		s.LogError(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (s *regattaService) ReceiveDataTicker(boatList []string, done chan struct{}) {
	fmt.Println("Starting ticker")

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	tickInterval := time.Second

	ticker := time.NewTicker(tickInterval)
	go func() {
		for {
			select {
			case <-interruptChannel:
				fmt.Println("Stopping ticker")
				ticker.Stop()
				close(done)
				return
			case <-ticker.C:
				for _, boat := range boatList {
					s.ReceiveData(boat)
				}
			}
		}
	}()
}

func (s *regattaService) ReceiveData(boat string) {
	fmt.Printf("ReceiveData called for boat %q\n", boat)

	ctx := context.Background()

	lastPosition, err := s.storageClient.GetLastPosition(ctx, boat, s.regattaStartTime, s.clock.RealNow())
	if err != nil {
		err = fmt.Errorf("get last position: %w", err)
		s.LogError(err)
		return
	}

	var startTime time.Time
	if lastPosition != nil {
		startTime = lastPosition.MeasureTime
	} else {
		startTime = s.regattaStartTime
	}

	httpBody := &ReadMessageRequest{
		Boat:      boat,
		StartTime: startTime,
		EndTime:   s.clock.RealNow(),
	}

	// Encode data to JSON
	httpBodyBytes, err := json.Marshal(httpBody)
	if err != nil {
		err = fmt.Errorf("marhsal http request: %w", err)
		s.LogError(err)
		return
	}

	// Make HTTP GET request
	req, err := http.NewRequest(http.MethodPost, s.dataServerURL, bytes.NewBuffer(httpBodyBytes))
	if err != nil {
		err = fmt.Errorf("create new HTTP request: %w", err)
		s.LogError(err)
		return
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("receive data from data server: %w", err)
		s.LogError(err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("receive status code %d from data server", resp.StatusCode)
		s.LogError(err)
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("read body from data server: %w", err)
		s.LogError(err)
		return
	}

	var positions *DataServerReadMessageResponse
	err = json.Unmarshal(bodyBytes, &positions)
	if err != nil {
		err = fmt.Errorf("decode HTTP response: %w", err)
		s.LogError(err)
		return
	}

	if len(positions.PositionsAtTime) == 0 {
		return
	}

	// Start analyzing incoming data

	if lastPosition != nil && lastPosition.MeasureTime.After(positions.PositionsAtTime[0].MeasureTime) {
		// TODO: Handle this case
		// We have to recalculate the distances from positions.PositionsAtTime[0] again.
		// Currently, we assume that the positions are always in ascending order and don't handle this case.
		s.LogError(errors.New("positions at time is too old"))
		s.LogDebug(fmt.Sprintf("%s, %s, %s", s.clock.RealNow(), lastPosition.MeasureTime.String(), positions.PositionsAtTime[0].MeasureTime.String()))
		return
	}

	err = s.insertPositions(ctx, lastPosition, boat, positions)
	if err != nil {
		err = fmt.Errorf("insert positions: %w", err)
		s.LogError(err)
		return
	}

	err = s.updateRoundsAndSections(ctx, lastPosition, boat, positions)
	if err != nil {
		err = fmt.Errorf("update rounds and sections: %w", err)
		s.LogError(err)
		return
	}
}

func (s *regattaService) insertPositions(ctx context.Context, lastPosition *StoragePosition, boat string, positions *DataServerReadMessageResponse) error {
	var storagePositions []StoragePosition

	regattaID, err := s.storageClient.GetRegattaAtTime(ctx, positions.PositionsAtTime[0].MeasureTime)
	if err != nil {
		err = fmt.Errorf("get regatta time: %w", err)
		return err
	}

	// Add first position

	storagePosition := StoragePosition{
		RegattaID:   regattaID,
		BoatID:      boat,
		Latitude:    positions.PositionsAtTime[0].Latitude,
		Longitude:   positions.PositionsAtTime[0].Longitude,
		Distance:    0,
		Heading:     0,
		Velocity:    0,
		MeasureTime: positions.PositionsAtTime[0].MeasureTime,
		SendTime:    positions.PositionsAtTime[0].SendTime,
	}

	if lastPosition != nil {
		additionalDistance := calculateDistanceInNM(lastPosition.Latitude, lastPosition.Longitude, positions.PositionsAtTime[0].Latitude, positions.PositionsAtTime[0].Longitude)
		timeDeltaInSeconds := positions.PositionsAtTime[0].MeasureTime.Sub(lastPosition.MeasureTime).Seconds()

		heading := lastPosition.Heading
		if additionalDistance > 0 {
			heading = calculateHeading(lastPosition.Latitude, lastPosition.Longitude, positions.PositionsAtTime[0].Latitude, positions.PositionsAtTime[0].Longitude)
		}

		velocity := lastPosition.Velocity
		if timeDeltaInSeconds > 0 {
			velocity = additionalDistance * 3600 / timeDeltaInSeconds // knots
		}

		storagePosition.Distance = lastPosition.Distance + additionalDistance
		storagePosition.Heading = heading
		storagePosition.Velocity = velocity
	}

	storagePositions = append(storagePositions, storagePosition)

	// Add all other positions

	firstEntry := true
	for i, position := range positions.PositionsAtTime {
		if firstEntry {
			firstEntry = false
			continue
		}

		lastPosition := storagePositions[i-1]

		additionalDistance := calculateDistanceInNM(lastPosition.Latitude, lastPosition.Longitude, position.Latitude, position.Longitude)
		timeDeltaInSeconds := position.MeasureTime.Sub(lastPosition.MeasureTime).Seconds()

		heading := lastPosition.Heading
		if additionalDistance > 0 {
			heading = calculateHeading(lastPosition.Latitude, lastPosition.Longitude, position.Latitude, position.Longitude)
		}

		velocity := lastPosition.Velocity
		if timeDeltaInSeconds > 0 {
			velocity = additionalDistance * 3600 / timeDeltaInSeconds // knots
		}

		regattaID, err = s.storageClient.GetRegattaAtTime(ctx, position.MeasureTime)
		if err != nil {
			err = fmt.Errorf("get regatta time: %w", err)
			return err
		}

		storagePosition := StoragePosition{
			RegattaID:   regattaID,
			BoatID:      boat,
			Latitude:    position.Latitude,
			Longitude:   position.Longitude,
			Distance:    lastPosition.Distance + additionalDistance,
			Heading:     heading,
			Velocity:    velocity,
			MeasureTime: position.MeasureTime,
			SendTime:    position.SendTime,
		}
		storagePositions = append(storagePositions, storagePosition)
	}

	err = s.storageClient.InsertPositions(ctx, storagePositions)
	if err != nil {
		err = fmt.Errorf("inserting positions: %w", err)
		return err
	}

	return nil
}

func (s *regattaService) updateRoundsAndSections(ctx context.Context, lastPosition *StoragePosition, boat string, positions *DataServerReadMessageResponse) error {
	var oldPosition PositionAtTime
	var oldRegattaID *string
	var skipEntry bool

	if lastPosition != nil {
		oldPosition = PositionAtTime{
			Latitude:    lastPosition.Latitude,
			Longitude:   lastPosition.Longitude,
			MeasureTime: lastPosition.MeasureTime,
		}
		var err error
		oldRegattaID, err = s.storageClient.GetRegattaAtTime(ctx, lastPosition.MeasureTime)
		if err != nil {
			err = fmt.Errorf("get regatta time: %w", err)
			return err
		}
	} else {
		// Start first round and first section
		regattaID, err := s.storageClient.GetRegattaAtTime(ctx, positions.PositionsAtTime[0].MeasureTime)
		if err != nil {
			err = fmt.Errorf("get regatta time: %w", err)
			return err
		}

		if regattaID != nil {
			// buoy order: "Schwanenwik bridge", "Kennedy bridge", "Langer Zug", "Pier"
			buoys, err := s.storageClient.GetBuoysAtTime(ctx, positions.PositionsAtTime[0].MeasureTime)
			if err != nil {
				return err
			}

			err = s.storageClient.StartRound(ctx, 1, *regattaID, boat, positions.PositionsAtTime[0].MeasureTime)
			if err != nil {
				err = fmt.Errorf("start first round: %w", err)
				s.LogError(err)
				return err
			}
			err = s.storageClient.StartSection(
				ctx,
				1,
				1,
				*regattaID,
				boat,
				positions.PositionsAtTime[0].MeasureTime,
				buoys[3].ID,
				buoys[3].Version,
				buoys[0].ID,
				buoys[0].Version)
			if err != nil {
				err = fmt.Errorf("start first section: %w", err)
				s.LogError(err)
				return err
			}
		}
		oldPosition = positions.PositionsAtTime[0]
		oldRegattaID = regattaID
		skipEntry = true
	}

	for _, position := range positions.PositionsAtTime {
		if skipEntry {
			skipEntry = false
			continue
		}

		// buoy order: "Schwanenwik bridge", "Kennedy bridge", "Langer Zug", "Pier"
		buoys, err := s.storageClient.GetBuoysAtTime(ctx, position.MeasureTime)
		if err != nil {
			return err
		}

		regattaID, err := s.storageClient.GetRegattaAtTime(ctx, position.MeasureTime)
		if err != nil {
			err = fmt.Errorf("get regatta time: %w", err)
			s.LogError(err)
			return err
		}
		if oldRegattaID == nil && regattaID == nil {
			// No regatta ID available
			// -> skip entry
			continue
		} else if oldRegattaID == nil {
			// We just entered a regatta
			// -> start first round and section
			err := s.storageClient.StartRound(ctx, 1, *regattaID, boat, position.MeasureTime)
			if err != nil {
				err = fmt.Errorf("start first round: %w", err)
				return err
			}

			err = s.storageClient.StartSection(
				ctx,
				1,
				1,
				*regattaID,
				boat,
				position.MeasureTime, buoys[3].ID,
				buoys[3].Version,
				buoys[0].ID,
				buoys[0].Version)
			if err != nil {
				err = fmt.Errorf("start first section: %w", err)
				s.LogError(err)
				return err
			}

		} else if regattaID == nil {
			// We just left a regatta
			// -> end last round and section
			round, err := s.storageClient.GetCurrentRound(ctx, *oldRegattaID, boat)
			if err != nil {
				err = fmt.Errorf("get current round: %w", err)
				s.LogError(err)
				return err
			}

			section, err := s.storageClient.GetCurrentSection(ctx, round, *oldRegattaID, boat)
			if err != nil {
				err = fmt.Errorf("get current section: %w", err)
				s.LogError(err)
				return err
			}

			if round == 0 || section == 0 {
				// No round or section available, so we cannot end it
				err = fmt.Errorf("no round or section available to end for boat %q in regatta %q", boat, *oldRegattaID)
				s.LogError(err)
				return err
			}

			err = s.storageClient.EndSection(ctx, section, round, *oldRegattaID, boat, position.MeasureTime)
			if err != nil {
				err = fmt.Errorf("end section: %w", err)
				s.LogError(err)
				return err
			}

			err = s.storageClient.EndRound(ctx, round, *oldRegattaID, boat, position.MeasureTime)
			if err != nil {
				err = fmt.Errorf("end round: %w", err)
				s.LogError(err)
				return err
			}

		} else if *oldRegattaID != *regattaID {
			// We left a regatta and entered a new one
			// -> end last round and section for old regatta
			// -> start new round and section for new regatta
			round, err := s.storageClient.GetCurrentRound(ctx, *oldRegattaID, boat)
			if err != nil {
				err = fmt.Errorf("get current round: %w", err)
				s.LogError(err)
				return err
			}

			section, err := s.storageClient.GetCurrentSection(ctx, round, *oldRegattaID, boat)
			if err != nil {
				err = fmt.Errorf("get current section: %w", err)
				s.LogError(err)
				return err
			}

			if round == 0 || section == 0 {
				// No round or section available, so we cannot end it
				err = fmt.Errorf("no round or section available to end for boat %q in regatta %q", boat, *oldRegattaID)
				s.LogError(err)
				return err
			}

			err = s.storageClient.EndSection(ctx, section, round, *oldRegattaID, boat, position.MeasureTime)
			if err != nil {
				err = fmt.Errorf("end section: %w", err)
				s.LogError(err)
				return err
			}

			err = s.storageClient.EndRound(ctx, round, *oldRegattaID, boat, position.MeasureTime)
			if err != nil {
				err = fmt.Errorf("end round: %w", err)
				s.LogError(err)
				return err
			}

			err = s.storageClient.StartRound(ctx, 1, *regattaID, boat, position.MeasureTime)
			if err != nil {
				err = fmt.Errorf("start first round: %w", err)
				return err
			}

			err = s.storageClient.StartSection(
				ctx,
				1,
				1,
				*regattaID,
				boat,
				position.MeasureTime,
				buoys[3].ID,
				buoys[3].Version,
				buoys[0].ID,
				buoys[0].Version)
			if err != nil {
				err = fmt.Errorf("start first section: %w", err)
				s.LogError(err)
				return err
			}
		} else {
			// We are still in the same regatta
			// -> check if we need to update rounds and sections

			// check if we have an open round, start new one if not
			round, err := s.storageClient.GetCurrentRound(ctx, *regattaID, boat)
			if err != nil {
				err = fmt.Errorf("get current round: %w", err)
				s.LogError(err)
				s.LogDebug(fmt.Sprintf("hit! %s, %s", *regattaID, boat))
				return err
			}

			if round == 0 {
				round, err = s.storageClient.GetLastCompletedRound(ctx, *regattaID, boat)
				if err != nil {
					err = fmt.Errorf("get last completed round: %w", err)
					return err
				}

				round += 1

				err = s.storageClient.StartRound(ctx, round, *regattaID, boat, position.MeasureTime)
				if err != nil {
					err = fmt.Errorf("start round: %w", err)
					s.LogError(err)
					return err
				}

				err = s.storageClient.StartSection(
					ctx,
					1,
					round,
					*regattaID,
					boat,
					position.MeasureTime,
					buoys[3].ID,
					buoys[3].Version,
					buoys[0].ID,
					buoys[0].Version,
				)
			}

			// check if we have an open section, start new one if not
			section, err := s.storageClient.GetCurrentSection(ctx, round, *regattaID, boat)
			if err != nil {
				err = fmt.Errorf("get current section: %w", err)
				s.LogError(err)
				return err
			}

			if section == 0 {
				section, err = s.storageClient.GetLastCompletedSection(ctx, round, *regattaID, boat)
				if err != nil {
					err = fmt.Errorf("get last completed section: %w", err)
					return err
				}

				section %= 4
				section += 1

				buoyStart := (section + 2) % 4
				buoyEnd := (section + 3) % 4

				err = s.storageClient.StartSection(
					ctx,
					section,
					round,
					*regattaID,
					boat,
					position.MeasureTime,
					buoys[buoyStart].ID,
					buoys[buoyStart].Version,
					buoys[buoyEnd].ID,
					buoys[buoyEnd].Version)
				if err != nil {
					err = fmt.Errorf("start section: %w", err)
					s.LogError(err)
					return err
				}
			}

			if position.MeasureTime.Sub(oldPosition.MeasureTime).Seconds() == 0 {
				continue
			}

			oldPosition2 := &Position{
				Latitude:  oldPosition.Latitude,
				Longitude: oldPosition.Longitude,
				Distance:  0,
				Time:      oldPosition.MeasureTime,
			}

			position2 := &Position{
				Latitude:  position.Latitude,
				Longitude: position.Longitude,
				Distance:  0,
				Time:      position.MeasureTime,
			}

			passed, err := calculateIfBuoysPassed(buoys, oldPosition2, position2)
			if err != nil {
				return err
			}

			if !passed[section-1] {
				// skip if relevant buoy was not passed
				continue
			}

			err = s.storageClient.EndSection(ctx, section, round, *oldRegattaID, boat, position.MeasureTime)
			if err != nil {
				err = fmt.Errorf("end section: %w", err)
				s.LogError(err)
				return err
			}

			if section < 4 {
				nextSection := section + 1
				buoyStart := (nextSection + 2) % 4
				buoyEnd := (nextSection + 3) % 4

				err = s.storageClient.StartSection(
					ctx,
					nextSection,
					round,
					*oldRegattaID,
					boat,
					position.MeasureTime,
					buoys[buoyStart].ID,
					buoys[buoyStart].Version,
					buoys[buoyEnd].ID,
					buoys[buoyEnd].Version)
				if err != nil {
					err = fmt.Errorf("start section: %w", err)
					s.LogError(err)
					return err
				}
			} else {
				err = s.storageClient.EndRound(ctx, round, *oldRegattaID, boat, position.MeasureTime)
				if err != nil {
					err = fmt.Errorf("end round: %w", err)
					s.LogError(err)
					return err
				}

				nextRound := round + 1
				err = s.storageClient.StartRound(ctx, nextRound, *oldRegattaID, boat, position.MeasureTime)
				if err != nil {
					err = fmt.Errorf("start round: %w", err)
					s.LogError(err)
					return err
				}

				err = s.storageClient.StartSection(
					ctx,
					1,
					nextRound,
					*oldRegattaID,
					boat,
					position.MeasureTime,
					buoys[3].ID,
					buoys[3].Version,
					buoys[0].ID,
					buoys[0].Version)
				if err != nil {
					err = fmt.Errorf("start section: %w", err)
					s.LogError(err)
					return err
				}
			}
		}

		oldPosition = position
	}

	return nil
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
