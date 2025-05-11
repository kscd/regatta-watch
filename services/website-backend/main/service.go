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
	"time"
)

const (
	nauticalMilesPerDegree  = 60
	buoyFarOffPointDistance = 1000
	pearlChainLength        = 10
)

var buoys = []buoy{
	{ // buoy Schwanenwik bridge
		Latitude:                 53.565538,
		Longitude:                10.014123 - 0.005,
		PassAngle:                90,
		IsPassDirectionClockwise: true,
		ToleranceInMeters:        100,
	},
	{ // buoy Kennedy bridge
		Latitude:                 53.558766 + 0.0035,
		Longitude:                9.998720 + 0.0055,
		PassAngle:                225,
		IsPassDirectionClockwise: true,
		ToleranceInMeters:        100,
	},
	{ // buoy Langer Zug entry
		Latitude:                 53.576497 - 0.001,
		Longitude:                10.004418 + 0.001,
		PassAngle:                45,
		IsPassDirectionClockwise: true,
		ToleranceInMeters:        100,
	},
	{ // pier (placed north of the Langer Zug pointing down)
		Latitude:                 53.577880,
		Longitude:                10.008151,
		PassAngle:                160,
		IsPassDirectionClockwise: true,
		ToleranceInMeters:        100,
	},
}

type boatState struct {
	distance               float64
	velocity               float64
	heading                float64
	currentRound           int
	currentSection         int
	roundTimes             []float64 // in seconds
	sectionTimes           []float64 // in seconds
	lastSectionTimestamp   time.Time
	lastRoundTimestamp     time.Time
	lastDataPointTimestamp time.Time
	lastPosition           *Position
	pearlChainPositions    []position
	pearlChainTimes        []time.Time
}

type regattaService struct {
	storageClient    storageInterface
	httpClient       *http.Client
	dataServerURL    string
	boatStates       map[string]*boatState
	pearlChainLength int
	pearlChainStep   float64
	hackertalkTime   time.Time
}

type storageInterface interface {
	GetLastTwoPositions(_ context.Context, boat string, _ time.Time) (*LastTwoPositions, error)
	GetPositions(ctx context.Context, boat string, startTime, endTime time.Time) ([]Position, error)
	InsertPositions(ctx context.Context, position *DataServerReadMessageResponse) error
	GetMode() string
}

func newRegattaService(storageClient storageInterface, dataServerURL string, pearlChainLength int, pearlChainStep float64, httpClient *http.Client) *regattaService {
	return &regattaService{
		storageClient:    storageClient,
		dataServerURL:    dataServerURL,
		httpClient:       httpClient,
		boatStates:       make(map[string]*boatState),
		pearlChainLength: pearlChainLength,
		pearlChainStep:   pearlChainStep,
		hackertalkTime:   time.Date(2024, 8, 3, 10, 0, 0, 0, time.UTC),
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

	_ = r.Context()

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

	positions, err := s.storageClient.GetLastTwoPositions(context.Background(), m.Boat, time.Now())
	if err != nil {
		s.LogError(fmt.Errorf("get positions: %v", err))
		return
	}

	if positions == nil {
		return
	}

	heading := calculateHeading(
		positions.LastPosition.Latitude,
		positions.LastPosition.Longitude,
		positions.CurrentPosition.Latitude,
		positions.CurrentPosition.Longitude,
	)
	additionalDistance := calculateDistanceInNM(
		positions.LastPosition.Latitude,
		positions.LastPosition.Longitude,
		positions.CurrentPosition.Latitude,
		positions.CurrentPosition.Longitude,
	)

	s.boatStates[m.Boat].distance += additionalDistance

	var timeDeltaInSeconds float64
	timeDeltaInSeconds = positions.CurrentPosition.Time.Sub(positions.LastPosition.Time).Seconds()
	velocity := additionalDistance * 3600 / timeDeltaInSeconds // knots
	if timeDeltaInSeconds == 0 {
		velocity = 0
	}
	s.boatStates[m.Boat].velocity = velocity
	s.boatStates[m.Boat].heading = heading

	crew, ok := roundToCrew[s.boatStates[m.Boat].currentRound]
	if !ok {
		crew = []string{"?", "?"}
	}
	nextCrew, ok := roundToCrew[s.boatStates[m.Boat].currentRound+1]
	if !ok {
		nextCrew = []string{"?", "?"}
	}

	lastPosition := s.boatStates[m.Boat].lastPosition
	if lastPosition == nil {
		lastPosition = &Position{}
	}

	response := FetchPositionResponse{
		MeasureTime: lastPosition.Time,
		Latitude:    positions.CurrentPosition.Latitude,  // lastPosition.Latitude
		Longitude:   positions.CurrentPosition.Longitude, // lastPosition.Longitude
		Heading:     s.boatStates[m.Boat].heading,
		Distance:    s.boatStates[m.Boat].distance,
		Velocity:    s.boatStates[m.Boat].velocity,
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

	_ = r.Context()

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

	/*positions, err := s.storageClient.GetPositions(context.Background(), m.Boat, time.Now(), pearlChainLength+1) // m.NoLaterThan
	if err != nil {
		fmt.Printf("/fetchposition get positions: %v", err)
		return
	}

	// positions are the wrong way, have to fix this when re-doing the pearl chain
	reverseSlice(positions)

	pearlLength := pearlChainLength
	if len(positions) < pearlChainLength+1 {
		pearlLength = len(positions) - 1 // 1 offset for heading calculations
	}

	var pearlChainPositions []position
	for i := 0; i < pearlLength; i++ {
		heading := calculateHeading(
			positions[i+1].Latitude,
			positions[i+1].Longitude,
			positions[i].Latitude,
			positions[i].Longitude)

		pearlChainPositions = append(pearlChainPositions,
			position{
				Latitude:  positions[i].Latitude,
				Longitude: positions[i].Longitude,
				Heading:   heading,
			},
		)
	}
	*/

	response := FetchPearlChainResponse{Positions: s.boatStates[m.Boat].pearlChainPositions}

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

	roundTimeCurrent := time.Now().Sub(s.boatStates[m.Boat].lastRoundTimestamp).Seconds()
	sectionTimeCurrent := time.Now().Sub(s.boatStates[m.Boat].lastSectionTimestamp).Seconds()
	if s.storageClient.GetMode() == "hackertalk" {
		roundTimeCurrent = s.hackertalkTime.Sub(s.boatStates[m.Boat].lastRoundTimestamp).Seconds()
		sectionTimeCurrent = s.hackertalkTime.Sub(s.boatStates[m.Boat].lastSectionTimestamp).Seconds()
	}

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

func (s *regattaService) ReceiveDataTicker(done chan struct{}) {
	fmt.Println("Starting ticker")

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)

	tickInterval := time.Second
	if s.storageClient.GetMode() != "hackertalk" {
		tickInterval = time.Second
	}

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
				for boat := range s.boatStates {
					s.ReceiveData(boat)
				}
			}
		}
	}()
}

func (s *regattaService) ReceiveData(boat string) {
	fmt.Printf("ReceiveData called for boat %q\n", boat)

	if s.storageClient.GetMode() != "normal" && s.storageClient.GetMode() != "hackertalk" {
		return
	}

	endTime := time.Now()
	if s.storageClient.GetMode() == "hackertalk" {
		endTime = s.hackertalkTime
	}

	httpBody := &ReadMessageRequest{
		Boat:      boat,
		StartTime: s.boatStates[boat].lastDataPointTimestamp,
		EndTime:   endTime,
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

	err = s.storageClient.InsertPositions(context.Background(), positions)
	if err != nil {
		err = fmt.Errorf("inserting positions: %w", err)
		s.LogError(err)
		return
	}

	if s.storageClient.GetMode() == "hackertalk" {
		s.hackertalkTime = s.hackertalkTime.Add(120 * time.Second)
	}
	s.updateState(boat, PostitionAtTimeToPosition(positions.PositionsAtTime), true)
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

			s.boatStates[boat].heading = calculateHeading(s.boatStates[boat].lastPosition.Latitude, s.boatStates[boat].lastPosition.Longitude, currentPosition.Latitude, currentPosition.Longitude)

			additionalDistance := calculateDistanceInNM(s.boatStates[boat].lastPosition.Latitude, s.boatStates[boat].lastPosition.Longitude, currentPosition.Latitude, currentPosition.Longitude)
			s.boatStates[boat].distance += additionalDistance

			timeDeltaInSeconds := currentPosition.Time.Sub(s.boatStates[boat].lastPosition.Time).Seconds()
			velocity := additionalDistance * 3600 / timeDeltaInSeconds // knots
			if timeDeltaInSeconds == 0 {
				velocity = 0
			}
			s.boatStates[boat].velocity = velocity

			// Was one of the buoys passed? Updates rounds and sections + times
			s.calculateIfBuoysPassed(boat, s.boatStates[boat].lastPosition, &currentPosition, printBuoyUpdate)

			s.boatStates[boat].lastPosition = &currentPosition

			// Update Pearl Chain
			lenPC := len(s.boatStates[boat].pearlChainPositions)
			var lastPCTime time.Time
			if lenPC > 0 {
				lastPCTime = s.boatStates[boat].pearlChainTimes[lenPC-1]
			} else {
				lastPCTime = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
			}
			if currentPosition.Time.Sub(lastPCTime).Seconds() > s.pearlChainStep {
				s.boatStates[boat].pearlChainPositions = append(s.boatStates[boat].pearlChainPositions, position{
					Latitude:  currentPosition.Latitude,
					Longitude: currentPosition.Longitude,
					Heading:   s.boatStates[boat].heading,
				})
				s.boatStates[boat].pearlChainTimes = append(s.boatStates[boat].pearlChainTimes, currentPosition.Time)
				if lenPC+1 > s.pearlChainLength {
					s.boatStates[boat].pearlChainPositions = s.boatStates[boat].pearlChainPositions[1:]
					s.boatStates[boat].pearlChainTimes = s.boatStates[boat].pearlChainTimes[1:]
				}
			}

			if i%1000 == 0 && i > 0 {
				fmt.Printf("positions analysed: %d\n", i)
			}
		}
	}
}

func (s *regattaService) calculateIfBuoysPassed(boat string, positionOld, positionNew *Position, printUpdate bool) {
	timeDeltaInSeconds := positionNew.Time.Sub(positionOld.Time).Seconds()
	if timeDeltaInSeconds > 0 {
		passed, err := calculateIfBuoysPassed(positionOld, positionNew)
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

			if s.boatStates[boat].currentSection == len(buoys)-1 {
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
	positions, err := s.storageClient.GetPositions(context.Background(), boat, time.Now().AddDate(0, 1, 0), time.Now())
	if err != nil {
		err = fmt.Errorf("load all data: %w", err)
		s.LogError(err)
		return err
	}

	// TODO: change before regatta starts, this is the regatta start time
	var lastSectionTimestamp time.Time
	var lastRoundTimestamp time.Time
	if len(positions) == 0 {
		if s.storageClient.GetMode() == "hackertalk" {
			lastSectionTimestamp = time.Date(2024, time.August, 3, 11, 0, 0, 0, time.UTC)
			lastRoundTimestamp = time.Date(2024, time.August, 3, 11, 0, 0, 0, time.UTC)
		} else {
			lastSectionTimestamp = time.Now()
			lastRoundTimestamp = time.Now()
		}
	} else {
		lastSectionTimestamp = positions[0].Time
		lastRoundTimestamp = positions[0].Time
	}

	s.boatStates[boat] = &boatState{
		distance:               0,
		velocity:               0,
		heading:                0,
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

func PostitionAtTimeToPosition(p []PositionAtTime) []Position {
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
