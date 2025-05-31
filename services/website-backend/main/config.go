package main

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type config struct {
	DBConfig          databaseConfig
	DataServerURL     string
	RegattaStartTime  time.Time
	RegattaEndTime    time.Time
	GetDataFromServer bool
}

func loadConfig() (*config, error) {
	err := godotenv.Load("../.env")
	if err != nil {
		return nil, errors.New("error loading .env file")
	}

	host, ok := os.LookupEnv("HOST")
	if !ok {
		return nil, errors.New("HOST was not defined")
	}

	portStr, ok := os.LookupEnv("PORT")
	if !ok {
		return nil, errors.New("PORT was not defined")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	dbName, ok := os.LookupEnv("DB_NAME")
	if !ok {
		return nil, errors.New("DB_NAME was not defined")
	}

	dbUserName, ok := os.LookupEnv("DB_USER_NAME")
	if !ok {
		return nil, errors.New("DB_USER_NAME was not defined")
	}

	dbUserPassword, ok := os.LookupEnv("DB_USER_PASSWORD")
	if !ok {
		return nil, errors.New("DB_USER_PASSWORD was not defined")
	}

	dataServerURL, ok := os.LookupEnv("DATA_SERVER_URL")
	if !ok {
		return nil, errors.New("DATA_SERVER_URL was not defined")
	}

	regattaStartTimeRaw, ok := os.LookupEnv("REGATTA_START_TIME")
	if !ok {
		return nil, errors.New("REGATTA_START_TIME was not defined")
	}
	regattaStartTime, err := time.Parse(time.RFC3339, regattaStartTimeRaw)
	if err != nil {
		return nil, errors.New("error parsing REGATTA_START_TIME")
	}

	regattaEndTimeRaw, ok := os.LookupEnv("REGATTA_END_TIME")
	if !ok {
		return nil, errors.New("REGATTA_END_TIME was not defined")
	}
	regattaEndTime, err := time.Parse(time.RFC3339, regattaEndTimeRaw)
	if err != nil {
		return nil, errors.New("error parsing REGATTA_END_TIME")
	}

	getDataFromServer, ok := os.LookupEnv("GET_DATA_FROM_SERVER")
	if !ok {
		return nil, errors.New("GET_DATA_FROM_SERVER was not defined")
	}
	if getDataFromServer != "true" && getDataFromServer != "false" {
		return nil, errors.New("GET_DATA_FROM_SERVER must be 'true' or 'false'")
	}
	getDataFromServerBool := getDataFromServer == "true"

	dbConfig := databaseConfig{
		Host:         host,
		Port:         port,
		DatabaseName: dbName,
		UserName:     dbUserName,
		UserPassword: dbUserPassword,
	}

	return &config{
		DBConfig:          dbConfig,
		DataServerURL:     dataServerURL,
		RegattaStartTime:  regattaStartTime,
		RegattaEndTime:    regattaEndTime,
		GetDataFromServer: getDataFromServerBool,
	}, nil
}
