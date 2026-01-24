package middleware

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// bufferedWriter captures the response so we can settle payment before sending to client
type bufferedWriter struct {
	gin.ResponseWriter
	body     *bytes.Buffer
	status   int
	header   http.Header
	maxSize  int
	overflow bool
}

func newBufferedWriter(w gin.ResponseWriter, maxSize int) *bufferedWriter {
	return &bufferedWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		status:         200,
		header:         make(http.Header),
		maxSize:        maxSize,
	}
}

func (w *bufferedWriter) Write(data []byte) (int, error) {
	if w.maxSize > 0 && w.body.Len()+len(data) > w.maxSize {
		w.overflow = true
		return 0, fmt.Errorf("response exceeds max buffer size (%d bytes)", w.maxSize)
	}
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
