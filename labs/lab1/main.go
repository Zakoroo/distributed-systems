package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

// type FileLocker struct {
// 	fileMap sync.Map
// }
//
// func (l FileLocker) read(fileName string) (data []byte, err error) {
// 	// TODO
// 	// NOTE: If a file does not exist in the map, then it does not necessarily mean that the file does not exist it only means that no is currently locking the file.
// 	// APPROACH: Maintain a lock for each file created or modified during this session.
// 	return nil, nil
// }
//
// func (l FileLocker) write(fileName string, data []byte) (err error) {
// 	// TDDO
// 	// NOTE: If a file does not exist in the map, then it does not necessarily mean that the file does not exist it only means that no is currently locking the file.
// 	// APPROACH: Maintain a lock for each file created or modified during this session.
// 	return nil
// }

var logger log.Logger = *log.Default()

func main() {
	if err := start(); err != nil {
		log.Fatal(err)
	}
}

func start() (err error) {
	port, err := parsePort()
	if err != nil {
		return err
	}

	// Create listening socket on all network interfaces
	logger.Printf("Starting server on port %v ...\n", port)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}
	defer listener.Close()

	// TODO: Implement limit on the number of Go routines
	sem, err := NewSemaphore(10)
	if err != nil {
		return err
	}

	for {
		// Wait for client to connect
		conn, err := listener.Accept()
		if err != nil {
			logger.Println(err)
			continue
		}
		logger.Printf("Accepted new client connection: %v", conn.RemoteAddr())
		sem.Down()
		logger.Printf("Current counter is: %v", sem.count)

		// Handle client requests
		go func() {
			if err := handleRequest(conn); err != nil {
				logger.Println(err)
			}
			sem.Up()
			logger.Printf("Current counter is: %v", sem.count)
		}()
	}
}

func parsePort() (port int, err error) {
	// Make sure the right number of arguments are supplied
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %v <port>\n", os.Args[0])
		return 0, fmt.Errorf("Missing argument <port>")
	}

	// Make sure the port number is provided properly
	port, err = strconv.Atoi(os.Args[1])
	if err != nil {
		return 0, err
	}

	// Make sure the port is in valid range
	if port < 0 || port > 65535 {
		return 0, fmt.Errorf("Invalid port number: %v\n", port)
	}

	return port, nil
}

func handleRequest(conn net.Conn) (err error) {
	defer conn.Close()

	// Create a new buffer
	reader := bufio.NewReader(conn)

	// Read entire request
	req, err := http.ReadRequest(reader)
	if err != nil {
		return err
	}

	// Extract request detials
	method, path, body, err := extractRequestDetails(req)
	if err != nil {
		return err
	}

	// TODO: Implement Get and POST
	// Log request
	logger.Printf("method: %v, path: %v, body: %v\n", method, path, body)

	// Form the response
	responseText := "Hello world!\n"
	response := fillResponseDetails(200, []byte(responseText), "text/html")

	// Send response
	return response.Write(conn)
}

func extractRequestDetails(req *http.Request) (method, path string, body []byte, err error) {
	method = req.Method
	path = req.URL.Path

	if method == http.MethodPost && req.Body != nil {
		body, err = io.ReadAll(req.Body)
		req.Body.Close()

		if err != nil {
			return "", "", nil, err
		}
	}

	return method, path, body, nil
}

func fillResponseDetails(statusCode int, body []byte, contentType string) http.Response {
	var status = fmt.Sprintf("%v %v", statusCode, http.StatusText(statusCode))
	headers := make(http.Header)
	if contentType != "" {
		headers.Set("Content-Type", contentType)
	}

	return http.Response{
		Status:        status,
		StatusCode:    statusCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        headers,
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
	}
}
