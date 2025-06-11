package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

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

	storageClient, err := newDatabaseClient(c.DBConfig)
	if err != nil {
		log.Fatal("error creating database client: ", err)
	}

	/*
		certPool := x509.NewCertPool()
		serverCert, err := os.ReadFile("../../../https_certificate/cert.pem")
		if err != nil {
			log.Fatalf("Failed to read server certificate: %v", err)
		}
		certPool.AppendCertsFromPEM(serverCert)

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		}
	*/

	client := &http.Client{
		//Transport: transport,
		Timeout: 5 * time.Second,
	}

	regattaService := newRegattaService(
		storageClient,
		c.DataServerURL,
		c.RegattaStartTime,
		c.RegattaEndTime,
		client)

	http.HandleFunc("/ping", regattaService.Ping)
	http.HandleFunc("/fetchposition", regattaService.FetchPosition)
	http.HandleFunc("/fetchpearlchain", regattaService.FetchPearlChain)
	http.HandleFunc("/fetchroundtime", regattaService.FetchRoundTimes)
	http.HandleFunc("/setclockconfiguration", regattaService.SetClockConfiguration)
	http.HandleFunc("/resetclockconfiguration", regattaService.ResetClockConfiguration)
	http.HandleFunc("/getclocktime", regattaService.GetClockTime)
	http.HandleFunc("/fetchbuoys", regattaService.Fetchbuoys)
	server := &http.Server{Addr: ":8091"}

	idleConnectionsClosed := make(chan struct{})
	dataReceiverClosed := make(chan struct{})
	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, os.Interrupt)
		<-interruptChannel
		fmt.Println("Server shutting down")
		if err = server.Shutdown(context.Background()); err != nil {
			fmt.Printf("error shutting down: %v", err)
		}
		close(idleConnectionsClosed)
	}()

	boatList := []string{"Bluebird", "Vivace"}

	fmt.Println("Service started and listening")

	if c.GetDataFromServer {
		regattaService.ReceiveDataTicker(boatList, dataReceiverClosed)
	}
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("error in http handler: ", err)
	}

	<-idleConnectionsClosed
	<-dataReceiverClosed
}
