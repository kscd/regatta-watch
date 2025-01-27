package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type regattaService struct {
	dbClient *databaseClient
}

func newRegattaService(dbClient *databaseClient) *regattaService {
	return &regattaService{
		dbClient: dbClient,
	}
}

func (s *regattaService) LogError(err error) {
	log.Println(err.Error())
}

func (s *regattaService) Ping(w http.ResponseWriter, _ *http.Request) {
	if _, err := w.Write([]byte("pong")); err != nil {
		err = fmt.Errorf("ping: write to http response writer: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

func (s *regattaService) PushPositions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Println("endpoint /pushposition called")

	// read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("push position: read http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// parse data from request
	var m OwnTracksMessage
	if err = json.Unmarshal(body, &m); err != nil {
		err = fmt.Errorf("push position: unmarshal http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// convert data
	pmr := PushMessageRequest{
		Positions: []Position{
			{
				Boat:        "Bluebird",
				Longitude:   m.Longitude,
				Latitude:    m.Latitude,
				MeasureTime: time.Unix(int64(m.CreatedAt), 0),
			},
		},
		SendTime: time.Unix(int64(m.CreatedAt), 0),
	}

	// store new data in DB
	err = s.dbClient.InsertPositions(ctx, &pmr)
	if err != nil {
		err = fmt.Errorf("push position: insert into database: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

func (s *regattaService) ReadPositions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Println("endpoint /readpositions called")

	// parse data from request
	var m ReadMessageRequest
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

	positions, err := s.dbClient.GetPositions(ctx, m.Boat, m.StartTime, m.EndTime)
	if err != nil {
		err = fmt.Errorf("read position: extract from database: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	response := ReadMessageResponse{
		PositionsAtTime: positions,
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

func (s *regattaService) PushBattery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// parse data from request
	var m BatteryMessage
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("push battery: read http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &m); err != nil {
		err = fmt.Errorf("push battery: unmarshal http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// store new data in DB
	err = s.dbClient.InsertBatteryLevels(ctx, &m)
	if err != nil {
		err = fmt.Errorf("push battery: insert into database: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}

func (s *regattaService) ReadBattery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// parse data from request
	var b ReadMessageRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("read battery: read http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &b); err != nil {
		err = fmt.Errorf("read battery: unmarshal http body: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	batteryMessage, err := s.dbClient.ExtractBatteryLevel(ctx, b.StartTime, b.EndTime)
	if err != nil {
		err = fmt.Errorf("read battery: extract from database: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	responseBytes, err := json.Marshal(batteryMessage)
	if err != nil {
		err = fmt.Errorf("read battery: marshal response: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	_, err = w.Write(responseBytes)
	if err != nil {
		err = fmt.Errorf("read battery: write to http writer: %w", err)
		s.LogError(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
}
