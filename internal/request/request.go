package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode"

	"www.github.com/jaychillin2607/httpfromtcp/internal/headers"
)

type parserState string

var InvalidBodySizeError = fmt.Errorf("body size is inconsistent to Conten-Length")

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        []byte
	State       parserState
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	str, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}

	intValue, err := strconv.Atoi(str)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func newRequest() *Request {
	return &Request{
		State:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    []byte{},
	}
}

var ErrorMalformedRequestLine = fmt.Errorf("malformed request-line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(requestData []byte) (*RequestLine, int, error) {
	idx := bytes.Index(requestData, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := requestData[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorUnsupportedHttpVersion
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}
	for _, c := range rl.Method {
		if unicode.IsLower(c) || !unicode.IsLetter(c) {
			return nil, 0, ErrorMalformedRequestLine
		}
	}

	return rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		switch r.State {
		case StateError:
			return 0, ErrorRequestInErrorState

		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.State = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.State = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.State = StateError
				return 0, err
			}

			read += n

			if n == 0 {
				break outer
			}

			if done {
				r.State = StateBody
			}

		case StateBody:
			length := getInt(r.Headers, "content-length", 0)
			if length == 0 {
				r.State = StateDone
				break
			}

			remaining := min(length-len(r.Body), len(currentData))
			if remaining < len(currentData) {
				return 0, InvalidBodySizeError
			}

			r.Body = append(r.Body, currentData[:remaining]...)
			read += remaining
			if len(r.Body) == length {
				r.State = StateDone
				break
			}
			break outer

		case StateDone:
			break outer
		default:
			panic("somehow we have programmed poorly")
		}

	}

	return read, nil
}

func (r *Request) done() bool {
	return r.State == StateDone || r.State == StateError
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	// read all
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for {
		n, err := reader.Read(buf[bufLen:])
		if request.done() && n != 0 {
			return nil, InvalidBodySizeError
		}
		if err == io.EOF {
			if request.done() {
				break
			} else {
				return nil, InvalidBodySizeError
			}
		}
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[readN:bufLen])

		bufLen -= readN
	}

	return request, nil
}
