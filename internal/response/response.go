package response

import (
	"fmt"
	"io"
	"strconv"

	"www.github.com/jaychillin2607/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK                    StatusCode = 200
	BAD_REQUEST           StatusCode = 400
	INTERNAL_SERVER_ERROR StatusCode = 500
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeaders := headers.NewHeaders()
	defaultHeaders.Set("content-length", strconv.Itoa(contentLen))
	defaultHeaders.Set("connection", "close")
	defaultHeaders.Set("content-type", "text/html")
	return defaultHeaders
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusMessage := ""
	switch statusCode {
	case OK:
		statusMessage = "OK"
	case BAD_REQUEST:
		statusMessage = "Bad Request"
	case INTERNAL_SERVER_ERROR:
		statusMessage = "Internal Server Error"
	default:
		return fmt.Errorf("unrecognized status code")
	}
	_, err := w.writer.Write(fmt.Appendf(nil, "HTTP/1.1 %d %s\r\n", statusCode, statusMessage))
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	b := []byte{}
	h.ForEach(func(k, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", k, v)
	},
	)
	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	b := strconv.AppendInt(nil, int64(len(p)), 16)
	b = append(b, "\r\n"...)
	b = append(b, p...)
	b = append(b, "\r\n"...)
	n, err := w.writer.Write(b)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.WriteChunkedBody([]byte{})
}

func (w *Writer) WriteTrailer(t headers.Headers) error {
	_, err := w.WriteChunkedBody([]byte{})
	if err != nil {
		return err
	}
	err = w.WriteHeaders(t)
	if err != nil {
		return err
	}

	return nil
}
