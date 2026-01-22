package middleware

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// bufferedWriter captures the response so we can settle payment before sending to client
type bufferedWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
	header http.Header
}

func newBufferedWriter(w gin.ResponseWriter) *bufferedWriter {
	return &bufferedWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		status:         200,
		header:         make(http.Header),
	}
}

func (w *bufferedWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *bufferedWriter) WriteHeader(status int) {
	w.status = status
}

func (w *bufferedWriter) Header() http.Header {
	return w.header
}

func (w *bufferedWriter) Status() int {
	return w.status
}

func (w *bufferedWriter) flush() error {
	// Copy buffered headers to real response
	for k, v := range w.header {
		for _, val := range v {
			w.ResponseWriter.Header().Add(k, val)
		}
	}
	// Write status and body
	w.ResponseWriter.WriteHeader(w.status)
	_, err := w.ResponseWriter.Write(w.body.Bytes())
	return err
}
