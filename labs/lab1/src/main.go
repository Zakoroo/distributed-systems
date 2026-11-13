package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Zakoroo/distributed-systems/httpserver"
	"github.com/Zakoroo/distributed-systems/labs/lab1/src/interfaces/fileapi"
	"github.com/Zakoroo/distributed-systems/labs/lab1/src/interfaces/pubsubapi"
	"github.com/Zakoroo/distributed-systems/labs/lab1/src/middleware"
	"github.com/Zakoroo/distributed-systems/labs/lab1/src/pubsub"
)

const (
	publicPath               = "./public/"
	maxConcurrentConnections = 10
)

func main() {
	if err := start(); err != nil {
		log.Fatal(err)
	}
}

func start() error {
	port, err := parsePort()
	if err != nil {
		return err
	}
	s, err := newServer(publicPath)
	if err != nil {
		return err
	}
	log.Printf("Starting server on port %d ...", port)
	return s.Start(fmt.Sprintf(":%d", port))
}

func newServer(root string) (*httpserver.Server, error) {
	s, err := httpserver.New(maxConcurrentConnections)
	if err != nil {
		return nil, err
	}
	if err := s.RegisterMiddleware(".*", middleware.ValidateMethod); err != nil {
		return nil, err
	}
	files := fileapi.New(root)

	broker, err := pubsub.New(64)
	if err != nil {
		return nil, err
	}
	api, err := pubsubapi.New(broker, 15*time.Second)
	if err != nil {
		return nil, err
	}
	for _, endpoint := range []struct {
		path    string
		handler httpserver.Handler
	}{
		{"/subscribe", api.HandleSubscribe},
		{"/unsubscribe", api.HandleUnsubscribe},
		{"/publish", api.HandlePublish},
		{"/events", api.HandleEvents},
	} {
		if err := s.RegisterEndpoint(endpoint.path, endpoint.handler); err != nil {
			return nil, err
		}
	}

	// File validation belongs only to the fallback, leaving custom endpoints free
	// to define their own path policy. Chain runs validation in the listed order.
	// This captures the fucntionality for file reading/creation/update specified in lab 1 basic part.
	err = s.SetFallback(httpserver.Chain(files.Handle, middleware.ValidateFilePath, middleware.ValidateFileExtension))
	return s, err
}

func parsePort() (int, error) {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v <port>\n", os.Args[0])
		return 0, fmt.Errorf("missing argument <port>")
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		return 0, err
	}
	if port < 0 || port > 65535 {
		return 0, fmt.Errorf("invalid port number: %d", port)
	}
	return port, nil
}
