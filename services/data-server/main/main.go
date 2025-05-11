package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)

	c, err := loadConfig()
	if err != nil {
		log.Fatal("error loading config: ", err)
	}

	dbClient, err := newDatabaseClient(c.DBConfig, "test") // normal
	if err != nil {
		log.Fatal(err)
	}

	regattaService := newRegattaService(dbClient)

	//certFile := "../../../https_certificate/cert.pem"
	//keyFile := "../../../https_certificate/key.pem"

	http.HandleFunc("/ping", regattaService.Ping)
	http.HandleFunc("/pushposition", regattaService.PushPositions)
	http.HandleFunc("/readposition", regattaService.ReadPositions)
	http.HandleFunc("/pushbattery", regattaService.PushBattery)
	http.HandleFunc("/readbattery", regattaService.ReadBattery)

	fmt.Println("Service started and listening")
	//err = http.ListenAndServeTLS(":8090", certFile, keyFile, nil)
	err = http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("error in http handler: ", err)
	}
}
