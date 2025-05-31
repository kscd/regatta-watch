package main

import (
	"context"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
)

func main() {
	ctx := context.Background()

	c, err := loadConfig()
	if err != nil {
		log.Fatal("error loading config: ", err)
	}

	dbClient, err := NewDatabaseClient(c.DBConfig, "positions_data_server_test")
	if err != nil {
		log.Fatal(err)
	}

	err = dbClient.CreateBoatTable(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = dbClient.CreateRegattaTable(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = dbClient.CreateBuoyTable(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = dbClient.CreateRoundTable(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = dbClient.CreateSectionTable(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = dbClient.CreateGPSDataTable(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
