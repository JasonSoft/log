// Package memory implements an in-memory handler useful for testing, as the
// entries can be accessed after writes.
package memory

import (
	"sync"

	"github.com/jasonsoft/log"
)

// Handler implementation.
type Handler struct {
	mu  sync.Mutex
	Out []byte
}

// New handler.
func New() *Handler {
	return &Handler{
		Out: make([]byte, 500),
	}
}

// Hook implements log.Handler.
func (h *Handler) Hook(e *log.Entry) error {
	e.Str("level", e.Level.String())

	return nil
}

// Write implements log.Handler.
func (h *Handler) Write(e *log.Entry) error {

	h.Out = e.Buffer()
	return nil
}

// Flush clear all buffer
func (h *Handler) Flush() error {
	h.Out = []byte{}
	return nil
}
