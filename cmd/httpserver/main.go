package main

import (
	"crypto/sha256"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	headersPkg "www.github.com/jaychillin2607/httpfromtcp/internal/headers"
	"www.github.com/jaychillin2607/httpfromtcp/internal/request"
	"www.github.com/jaychillin2607/httpfromtcp/internal/response"
	"www.github.com/jaychillin2607/httpfromtcp/internal/server"
)

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) *server.HandlerError {
		var statusCode response.StatusCode
		var body []byte
		if req.RequestLine.RequestTarget == "/yourproblem" {
			statusCode = response.BAD_REQUEST
			body = []byte(respond400)
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			statusCode = response.INTERNAL_SERVER_ERROR
			body = []byte(respond500)
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
			target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")

			res, err := http.Get("https://httpbin.org/" + target)
			if err != nil {
				statusCode = response.INTERNAL_SERVER_ERROR
				body = []byte(respond500)
			} else {
				defer res.Body.Close()

				headers := response.GetDefaultHeaders(0)
				headers.Delete("content-length")
				headers.Replace("content-type", "text/plain")
				headers.Set("transfer-encoding", "chunked")
				headers.Set("trailer", "X-Content-SHA256")
				headers.Set("trailer", "X-Content-Length")
				w.WriteStatusLine(response.OK)
				w.WriteHeaders(headers)

				b := make([]byte, 32)
				fullBody := []byte{}
				for {
					n, err := res.Body.Read(b)

					if err != nil {
						break
					}
					fullBody = append(fullBody, b[:n]...)
					w.WriteChunkedBody(b[:n])
				}

				checksum := sha256.Sum256(fullBody)
				trailers := headersPkg.NewHeaders()
				trailers.Set("X-Content-SHA256", string(checksum[:]))
				trailers.Set("X-Content-Length", strconv.Itoa(len(fullBody)))
				w.WriteTrailer(trailers)
				w.WriteChunkedBodyDone()
				return nil
			}
		} else if req.RequestLine.RequestTarget == "/video" {
			videoData, err := os.ReadFile(".assets/vim.mp4")
			if err != nil {
				statusCode = response.INTERNAL_SERVER_ERROR
				body = []byte(respond500)
			} else {
				headers := response.GetDefaultHeaders(len(videoData))
				headers.Replace("content-type", "video/mp4")
				w.WriteStatusLine(response.OK)
				w.WriteHeaders(headers)
				w.WriteBody(videoData)
				w.WriteChunkedBodyDone()
				return nil
			}
		} else {
			statusCode = response.OK
			body = []byte(respond200)
		}

		headers := response.GetDefaultHeaders(len(body))
		w.WriteStatusLine(statusCode)
		w.WriteHeaders(headers)
		w.WriteBody(body)
		return nil
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
