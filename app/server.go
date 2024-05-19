package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	request := make([]byte, 1024)

	_, err = conn.Read(request)

	if err != nil {
		log.Fatal(err)
	}

	if strings.HasPrefix(string(request), "GET /echo/") {
		req, _, found := strings.Cut(string(request), "\r\n")

		if !found {
			writeStatusNotFound(conn)
		}

		fmt.Println(strings.TrimSuffix(strings.TrimPrefix(string(req), "GET /echo/"), " HTTP/1.1"))
	} else if strings.HasPrefix(string(request), "GET / HTTP/1.1") {
		writeStatusOk(conn, "")
	} else {
		writeStatusNotFound(conn)
	}

}

func writeStatusOk(conn net.Conn, body string) {
	if len(body) == 0 {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	}

	content_type := "text/plain"
	content_length := len(body)

	response_text := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", content_type, content_length, body)

	conn.Write([]byte(response_text))
}

func writeStatusNotFound(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}
