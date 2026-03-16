package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"www.github.com/jaychillin2607/httpfromtcp/internal/request"
)

const fileName string = "message.txt"

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer f.Close()
		defer close(out)
		str := ""
		readBuffer := make([]byte, 8)
		for {
			charsRead, err := f.Read(readBuffer)
			if err != nil {
				if err == io.EOF {
					err = nil
				} else {
					log.Fatalf("error %v", err)
				}
				break
			}

			if i := bytes.IndexByte(readBuffer[:charsRead], '\n'); i > -1 {
				str += string(readBuffer[:i])
				out <- str
				str = string(readBuffer[i+1 : charsRead])
			} else {
				str += string(readBuffer[:charsRead])
			}
		}
		if len(str) != 0 {
			out <- str
		}
	}()

	return out
}

func main() {
	port := ":42069"
	ln, err := net.Listen("tcp", port)
	fmt.Println("listening on port", port)
	if err != nil {
		log.Fatalf("error %v", err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("connection received", conn)
		go func(c net.Conn) {
			defer conn.Close()

			r, err := request.RequestFromReader(conn)
			if err != nil {
				log.Fatalf("Error while reading request: %v\n", err)
			}

			fmt.Printf(`Request line:
- Method: %s
- Target: %s
- Version: %s
Headers:
`, r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
			r.Headers.ForEach(
				func(k, v string) {
					fmt.Printf("- %s: %s\n", k, v)
				},
			)
		}(conn)
	}
}
