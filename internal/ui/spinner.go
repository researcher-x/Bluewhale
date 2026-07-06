// Package ui provides small terminal presentation helpers: a spinner for
// non-verbose progress feedback.
package ui

import (
	"fmt"
	"sync"
	"time"
)

// Spinner renders a rotating indicator on the current terminal line along
// with a status message. It is intended for use when verbose mode is
// disabled, so the user still has feedback that BlueWhale is working.
type Spinner struct {
	mu      sync.Mutex
	frames  []string
	message string
	stopCh  chan struct{}
	doneCh  chan struct{}
	running bool
}

// NewSpinner creates a new Spinner with the given initial status message.
func NewSpinner(message string) *Spinner {
	return &Spinner{
		frames:  []string{"|", "/", "-", "\\"},
		message: message,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
}

// Start begins rendering the spinner in a background goroutine.
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go func() {
		defer close(s.doneCh)
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()

		i := 0
		for {
			select {
			case <-s.stopCh:
				fmt.Print("\r\033[K") // clear line
				return
			case <-ticker.C:
				s.mu.Lock()
				msg := s.message
				s.mu.Unlock()
				fmt.Printf("\r\033[K%s %s", s.frames[i%len(s.frames)], msg)
				i++
			}
		}
	}()
}

// UpdateMessage changes the text shown next to the spinner.
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// Stop halts the spinner and clears its line.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	<-s.doneCh
}
