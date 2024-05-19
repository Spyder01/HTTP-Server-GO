package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const SEP string = "\r\n"

type Request struct {
	RequestLine string
	Headers     []string
	Body        string
}

func ParseRequest(req []byte) (Request, bool) {
	request_line, rest, found := strings.Cut(string(req), SEP)

	if !found {
		return Request{}, found
	}

	headers, body, found := strings.Cut(rest, fmt.Sprintf("%s%s", SEP, SEP))

	if !found {
		return Request{}, found
	}

	return Request{RequestLine: request_line, Headers: strings.Split(headers, SEP), Body: body}, true
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

var directory string = "."

func main() {
	args := os.Args[1:]

	if len(args) >= 2 && args[0] == "--directory" {
		directory = args[1]
	}
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		request := make([]byte, 1024)

		n, err := conn.Read(request)

		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(conn, request[:n])
	}
}

func handleConnection(conn net.Conn, request []byte) {

	defer conn.Close()

	request_, found := ParseRequest(request)

	fmt.Println(request_)

	if !found {
		writeStatusNotFound(conn)
	}

	if strings.HasPrefix(request_.RequestLine, "GET /files") {
		file_path := strings.Trim(strings.TrimPrefix(strings.TrimSuffix(request_.RequestLine, "HTTP/1.1"), "GET /files"), " ")

		full_path := filepath.Join(directory, file_path)

		if !Exists(full_path) {
			writeStatusNotFound(conn)
			return
		}

		data := make([]byte, GetFileSize(full_path))

		err := read_file(full_path, data)

		if err != nil {
			writeStatusNotFound(conn)
			return
		}

		writeStatusOk(conn, string(data), "application/octet-stream")

	} else if strings.HasPrefix(request_.RequestLine, "GET /user-agent") {
		user_agen, found := request_.GetUserAgent()

		fmt.Println(found)

		if !found {
			writeStatusNotFound(conn)
		}

		writeStatusOk(conn, user_agen, "text/plain")
	} else if strings.HasPrefix(request_.RequestLine, "GET /echo/") {
		str := strings.TrimSuffix(strings.TrimPrefix(request_.RequestLine, "GET /echo/"), " HTTP/1.1")

		writeStatusOk(conn, str, "text/plain")
	} else if strings.HasPrefix(request_.RequestLine, "GET / HTTP/1.1") {
		writeStatusOk(conn, "", "text/plain")
	} else {
		writeStatusNotFound(conn)
	}
}

func writeStatusOk(conn net.Conn, body string, content_type string) {
	if len(body) == 0 {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	}

	content_length := len(body)

	response_text := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", content_type, content_length, body)

	conn.Write([]byte(response_text))
}

func writeStatusNotFound(conn net.Conn) {
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func GetFileSize(path string) int {
	stat, err := os.Stat(path)

	if err != nil {
		return 0
	}

	return int(stat.Size())
}

func read_file(path string, buffer []byte) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Read(buffer)

	if err != nil {
		return err
	}

	return nil
}
