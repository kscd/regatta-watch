package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"time"
)

// regatta start time 2024: time.Date(2024, 8, 3, 10, 0, 0, 0, time.UTC),

const (
	nauticalMilesPerDegree  = 60
	buoyFarOffPointDistance = 1000
)

type boatState struct {
	currentRound           int
	currentSection         int
	roundTimes             []float64 // in seconds
	sectionTimes           []float64 // in seconds
	lastSectionTimestamp   time.Time
	lastRoundTimestamp     time.Time
	lastDataPointTimestamp time.Time
	lastPosition           *Position
}

type regattaService struct {
	storageClient    storageInterface
	httpClient       *http.Client
	dataServerURL    string
	boatStates       map[string]*boatState
	pearlChainLength int
	pearlChainStep   float64
	regattaStartTime time.Time
	regattaEndTime   time.Time
	clock            clockInterface
}

type clockInterface interface {
	Now() time.Time
}

type storageInterface interface {
	InsertPositions(ctx context.Context, position []StoragePosition) error
	GetLastPosition(ctx context.Context, boat string, lowerBound, upperBound time.Time) (*StoragePosition, error)
	GetPositions(ctx context.Context, boat string, startTime, endTime time.Time) ([]Position, error)
	GetRegattaAtTime(ctx context.Context, time time.Time) (*string, error)
	GetBuoysAtTime(ctx context.Context, time time.Time) ([]buoy, error)
}

func newRegattaService(
	storageClient storageInterface,
	dataServerURL string,
	pearlChainLength int,
	pearlChainStep float64,
	regattaStartTime time.Time,
	regattaEndTime time.Time,
	httpClient *http.Client) *regattaService {
	return &regattaService{
		storageClient:    storageClient,
		dataServerURL:    dataServerURL,
		httpClient:       httpClient,
		boatStates:       make(map[string]*boatState),
		pearlChainLength: pearlChainLength,
		pearlChainStep:   pearlChainStep,
		regattaStartTime: regattaStartTime,
		regattaEndTime:   regattaEndTime,
		clock:            newClock(),
	}
}

func (s *regattaService) LogError(err error) {
	log.Println(err.Error())
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

	position, err := s.storageClient.GetLastPosition(ctx, m.Boat, s.regattaStartTime, s.clock.Now())
	if err != nil {
		s.LogError(fmt.Errorf("get positions: %v", err))
		return
	}

	if position == nil {
		return
	}

	crew, ok := roundToCrew[s.boatStates[m.Boat].currentRound]
	if !ok {
		crew = []string{"?", "?"}
	}
	nextCrew, ok := roundToCrew[s.boatStates[m.Boat].currentRound+1]
	if !ok {
		nextCrew = []string{"?", "?"}
	}

	response := FetchPositionResponse{
		MeasureTime: position.MeasureTime,
		Latitude:    position.Latitude,
		Longitude:   position.Longitude,
		Heading:     position.Heading,
		Distance:    position.Distance,
		Velocity:    position.Velocity,
		Round:       s.boatStates[m.Boat].currentRound + 1,   // so it doesn't start at 0 in the front end
		Section:     s.boatStates[m.Boat].currentSection + 1, // so it doesn't start at 0 in the front end
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

	// Example, last 20 seconds and 10 positions
	pearlChainLength := 5
	pearlChainTime := 20 * time.Second

	endTime := s.clock.Now()
	startTime := endTime.Add(-pearlChainTime)

	// positions sorted in descending order
	positions, err := s.storageClient.GetPositions(ctx, m.Boat, startTime, endTime)
	if err != nil {
		fmt.Printf("/fetchposition get positions: %v", err)
		return
	}

	// TODO: If insufficient data points are available, return an empty response, need at least 2.

	// TODO: Also add a mode for static chain links. Can do this with a bit of book keeping by manipulating the
	// initial value of nextStop.

	// Calculate the time step for the pearl chain from database data
	dbEndTime := positions[0].Time
	dbStartTime := positions[len(positions)-2].Time                                // need an offset of 1 for heading calculation
	pearlChainStep := dbEndTime.Sub(dbStartTime) / time.Duration(pearlChainLength) // time.Duration is needed for type matching
	nextStop := endTime.Add(-pearlChainStep)

	var pearlChain []position
	for index := range positions {
		if positions[index].Time.Sub(nextStop) < 0 && index+1 < len(positions)-1 {
			pearlChain = append(pearlChain, position{
				Latitude:  positions[index].Latitude,
				Longitude: positions[index].Longitude,
				Heading: calculateHeading(
					positions[index].Latitude,
					positions[index].Longitude,
					positions[index+1].Latitude,
					positions[index+1].Longitude),
			})
			nextStop = nextStop.Add(-pearlChainStep)
		}
	}

	response := FetchPearlChainResponse{Positions: pearlChain}

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

	_ = r.Context()

	// parse data from request
	var m FetchRoundTimeRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("fetch round time: read http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &m); err != nil {
		err = fmt.Errorf("fetch round time: unmarshal http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	now := s.clock.Now()
	roundTimeCurrent := now.Sub(s.boatStates[m.Boat].lastRoundTimestamp).Seconds()
	sectionTimeCurrent := now.Sub(s.boatStates[m.Boat].lastSectionTimestamp).Seconds()

	response := FetchRoundTimeResponse{
		RoundTimes:   append(s.boatStates[m.Boat].roundTimes, roundTimeCurrent),
		SectionTimes: append(s.boatStates[m.Boat].sectionTimes, sectionTimeCurrent),
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

	httpBody := &ReadMessageRequest{
		Boat:      boat,
		StartTime: s.boatStates[boat].lastDataPointTimestamp,
		EndTime:   s.clock.Now(),
	}

	// Encode the data to JSON
	httpBodyBytes, err := json.Marshal(httpBody)
	if err != nil {
		err = fmt.Errorf("marhsal http request: %w", err)
		s.LogError(err)
		return
	}

	// Make the HTTP GET request
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

	lastPosition, err := s.storageClient.GetLastPosition(ctx, boat, s.regattaStartTime, s.clock.Now())
	if err != nil {
		return
	}

	if lastPosition != nil && lastPosition.MeasureTime.After(positions.PositionsAtTime[0].MeasureTime) {
		// TODO: Handle this case
		// We have to recalculate the distances from positions.PositionsAtTime[0] again.
		// Currently, we assume that the positions are always in ascending order and don't handle this case.
		fmt.Println("Positions at time is too old")
		return
	}

	var storagePositions []StoragePosition

	regattaID, err := s.storageClient.GetRegattaAtTime(ctx, positions.PositionsAtTime[0].MeasureTime)
	if err != nil {
		return
	}

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

		regattaID, err := s.storageClient.GetRegattaAtTime(ctx, position.MeasureTime)
		if err != nil {
			return
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
		s.LogError(err)
		return
	}

	s.updateState(boat, PositionAtTimeToPosition(positions.PositionsAtTime), true)
}

func (s *regattaService) updateState(boat string, positions []Position, printBuoyUpdate bool) {
	if len(positions) > 0 {

		// Set last time stamp for future DB querying.
		s.boatStates[boat].lastDataPointTimestamp = positions[len(positions)-1].Time

		// Loop through all the received data and update state
		for i := range positions {
			// If the last position is not known, set it to the first location and continue
			if s.boatStates[boat].lastPosition == nil {
				s.boatStates[boat].lastPosition = &positions[i]
				continue
			}
			currentPosition := positions[i]

			// Was one of the buoys passed? Updates rounds and sections + times
			s.calculateIfBuoysPassed(boat, s.boatStates[boat].lastPosition, &currentPosition, printBuoyUpdate)

			s.boatStates[boat].lastPosition = &currentPosition

			if i%1000 == 0 && i > 0 {
				fmt.Printf("positions analysed: %d\n", i)
			}
		}
	}
}

func (s *regattaService) calculateIfBuoysPassed(boat string, positionOld, positionNew *Position, printUpdate bool) {
	buoys, err := s.storageClient.GetBuoysAtTime(context.Background(), positionNew.Time)
	if err != nil {
		return
	}

	timeDeltaInSeconds := positionNew.Time.Sub(positionOld.Time).Seconds()
	if timeDeltaInSeconds > 0 {
		passed, err := calculateIfBuoysPassed(buoys, positionOld, positionNew)
		if err != nil {
			return
		}
		if printUpdate {
			for j := range passed {
				if passed[j] {
					fmt.Println("+++ Buoy passed: ", j, "+++")
				}
			}
		}
		if passed[s.boatStates[boat].currentSection] {
			sectionTime := positionNew.Time.Sub(s.boatStates[boat].lastSectionTimestamp).Seconds()
			s.boatStates[boat].sectionTimes = append(s.boatStates[boat].sectionTimes, sectionTime)
			s.boatStates[boat].lastSectionTimestamp = positionNew.Time

			if s.boatStates[boat].currentSection == 4-1 { // len(buoys) - 1
				roundTime := positionNew.Time.Sub(s.boatStates[boat].lastRoundTimestamp).Seconds()
				s.boatStates[boat].roundTimes = append(s.boatStates[boat].roundTimes, roundTime)
				s.boatStates[boat].lastRoundTimestamp = positionNew.Time

				s.boatStates[boat].currentRound++
				s.boatStates[boat].currentSection = 0
			} else {
				s.boatStates[boat].currentSection++
			}
		}
	}
}

func (s *regattaService) ReinitialiseState(boat string) error {
	now := s.clock.Now()

	positions, err := s.storageClient.GetPositions(context.Background(), boat, s.regattaStartTime, now)
	if err != nil {
		err = fmt.Errorf("load all data: %w", err)
		s.LogError(err)
		return err
	}

	slices.Reverse(positions)

	var lastSectionTimestamp time.Time
	var lastRoundTimestamp time.Time
	if len(positions) == 0 {
		lastSectionTimestamp = now
		lastRoundTimestamp = now
	} else {
		lastSectionTimestamp = positions[0].Time
		lastRoundTimestamp = positions[0].Time
	}

	s.boatStates[boat] = &boatState{
		currentRound:           0,
		currentSection:         0,
		roundTimes:             make([]float64, 0),
		sectionTimes:           make([]float64, 0),
		lastSectionTimestamp:   lastSectionTimestamp,
		lastRoundTimestamp:     lastRoundTimestamp,
		lastDataPointTimestamp: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		lastPosition:           nil,
	}

	s.updateState(boat, positions, false)
	fmt.Printf("Initialised boat %q with %d data points\n", boat, len(positions))
	return nil
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func PositionAtTimeToPosition(p []PositionAtTime) []Position {
	var positions []Position
	for i := range p {
		positions = append(positions, Position{
			Latitude:  p[i].Latitude,
			Longitude: p[i].Longitude,
			Time:      p[i].MeasureTime,
		})
	}
	return positions
}
