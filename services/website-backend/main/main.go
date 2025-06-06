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

	//mode := "normal" // "test", "normal"
	var storageClient storageInterface
	//if mode == "test" {
	//storageClient = newFakeStorage()
	//} else if mode == "normal" {
	storageClient, _ = newDatabaseClient(c.DBConfig)
	//}

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

	// TODO: Set proper pearl chain parameters
	pearlChainLength_ := 10
	pearlChainStep := 1.5
	regattaService := newRegattaService(
		storageClient,
		c.DataServerURL,
		pearlChainLength_,
		pearlChainStep,
		c.RegattaStartTime,
		c.RegattaEndTime,
		client)

	/*
		// TODO: Remove if Vivace is not measured.
		err = regattaService.ReinitialiseState("Vivace")
		if err != nil {
			log.Fatal(`cannot initialise state for boat "Vivace": `, err)
		}
	*/

	http.HandleFunc("/ping", regattaService.Ping)
	http.HandleFunc("/fetchposition", regattaService.FetchPosition)
	http.HandleFunc("/fetchpearlchain", regattaService.FetchPearlChain)
	http.HandleFunc("/fetchroundtime", regattaService.FetchRoundTimes)
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

	boatList := []string{"Bluebird"}

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
