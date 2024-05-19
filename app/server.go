package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type Request struct {
	RequestLine string
	Headers     []string
	Body        string
}

func ParseRequest(req []byte) (Request, bool) {
	request_line, rest, found := strings.Cut(string(req), "/r/n")

	fmt.Println(rest)

	if !found {
		return Request{}, found
	}

	headers, body, found := strings.Cut(rest, "\r\n\r\n")

	if !found {
		return Request{}, found
	}

	return Request{RequestLine: request_line, Headers: strings.Split(headers, "/r/n"), Body: body}, true
}

func (req *Request) GetUserAgent() (string, bool) {

	for _, v := range req.Headers {
		key, val, exists := strings.Cut(v, ":")

		if !exists {
			continue
		}

		if strings.ToLower(key) == "user-agent" {
			return strings.Trim(val, " "), true
		}
	}

	return "", false
}

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

	request_, found := ParseRequest(request)

	if !found {
		writeStatusNotFound(conn)
	}

	fmt.Println(request_)

	if strings.HasPrefix(request_.RequestLine, "GET /user-agent") {
		user_agen, found := request_.GetUserAgent()

		fmt.Println(found)

		if !found {
			writeStatusNotFound(conn)
		}

		writeStatusOk(conn, user_agen)
	} else if strings.HasPrefix(request_.RequestLine, "GET /echo/") {
		str := strings.TrimSuffix(strings.TrimPrefix(request_.RequestLine, "GET /echo/"), " HTTP/1.1")

		writeStatusOk(conn, str)
	} else if strings.HasPrefix(request_.RequestLine, "GET / HTTP/1.1") {
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
