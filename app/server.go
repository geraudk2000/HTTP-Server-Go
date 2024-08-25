package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handleConnection(conn net.Conn) {

	defer conn.Close()

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request. ", err.Error())
		return
	}

	var res string

	path := request.URL.Path
	//fmt.Printf(request.Method)

	if path == "/" {
		res = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.HasPrefix(path, "/files/") {

		file_name := strings.Split(path, "/")[2]
		args := os.Args

		if len(args) > 2 && args[1] == "--directory" {
			dir := args[2]
			file_path := filepath.Join(dir, file_name)

			if request.Method == "GET" {

				file, err := os.Open(file_path)
				if err != nil {
					res = "HTTP/1.1 404 Not Found\r\n\r\n"
				} else {
					defer file.Close()

					content, err := os.ReadFile(file_path)
					if err != nil {
						fmt.Printf("Failed to read file: %v\n", err)
						return
					}
					fileContent := string(content)
					res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(fileContent), fileContent)
				}
			} else if request.Method == "POST" {
				file, err := os.OpenFile(file_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Printf("Faile to open file: %v\n", err)
				}
				defer file.Close()

				body, err := io.ReadAll(request.Body)
				if err != nil {
					fmt.Println("Cannot read body:", err)
					return
				}
				content := string(body)
				//fmt.Println(string(body))
				_, err = file.Write([]byte(content))
				if err != nil {
					fmt.Printf("Failed to write to file: %v\n", err)
					return
				}
				res = "HTTP/1.1 201 Created\r\n\r\n"

			}
		} else {
			fmt.Println("Usage: go run main.go --directory <directory_path>")
		}

	} else if path == "/user-agent" {
		res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(request.UserAgent()), request.UserAgent())
	} else if strings.HasPrefix(path, "/echo/") {
		echo := strings.Split(path, "/")[2]

		if request.Header.Get("Accept-Encoding") != "" && strings.Contains(request.Header["Accept-Encoding"][0], "gzip") {
			res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
		} else {
			res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
		}

	} else {
		res = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	conn.Write([]byte(res))
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

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
		go handleConnection(conn)
	}

}
