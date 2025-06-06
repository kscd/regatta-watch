package main

import (
	"errors"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type config struct {
	DBConfig DatabaseConfig
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

	dbConfig := DatabaseConfig{
		Host:         host,
		Port:         port,
		DatabaseName: dbName,
		UserName:     dbUserName,
		UserPassword: dbUserPassword,
	}

	return &config{
		DBConfig: dbConfig,
	}, nil
}
