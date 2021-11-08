package signal

import (
	"context"
	"os"
	"os/signal"
	"sync"
)

// Handler is a function that handles the signal.
type Handler func(os.Signal)

// Router routes signals to registered handler.
type Router interface {
	// Handle registers a signal handler.
	Handle(sig os.Signal, h Handler)

	// Reset resets a signal handler.
	Reset(sig os.Signal)

	// Ignore ignores a signal.
	Ignore(sig os.Signal)

	// IsHandled checks if a signal being routed to a handler.
	IsHandled(sig os.Signal) bool

	// IsIgnored checks if a signal is being ignored.
	IsIgnored(sig os.Signal) bool

	// Fire a signal.
	Fire(sig os.Signal)

	// Start the router
	Start() error

	// Stop the router
	Stop(err error)
}

type router struct {
	signalCh   chan os.Signal
	signals    map[os.Signal]Handler
	ignSignals map[os.Signal]struct{}
	ctx        context.Context
	cancelFunc context.CancelFunc
	lock       *sync.RWMutex
}

// New returns a signal router.
func New(ctx context.Context) Router {
	channelSize := 10
	fctx, fcancel := context.WithCancel(ctx)

	return &router{
		signalCh:   make(chan os.Signal, channelSize),
		signals:    make(map[os.Signal]Handler),
		ignSignals: make(map[os.Signal]struct{}),
		ctx:        fctx,
		cancelFunc: fcancel,
		lock:       &sync.RWMutex{},
	}
}

func (s *router) Handle(sig os.Signal, h Handler) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.signals[sig] = h
	signal.Notify(s.signalCh, sig)
	delete(s.ignSignals, sig)
}

func (s *router) Reset(sig os.Signal) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.signals, sig)
	signal.Reset(sig)
}

func (s *router) Ignore(sig os.Signal) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.signals, sig)
	signal.Ignore(sig)

	s.ignSignals[sig] = struct{}{}
}

func (s *router) IsHandled(sig os.Signal) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.signals[sig]

	return ok
}

func (s *router) IsIgnored(sig os.Signal) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.ignSignals[sig]

	return ok
}

// Start starts the signal router and listens for registered signals.
func (s *router) Start() error {
	// This go routine dies with the server
	for {
		select {
		case <-s.ctx.Done():
			// Context got cancelled, exit
			return nil
		case sig := <-s.signalCh:
			func() {
				s.lock.RLock()
				defer s.lock.RUnlock()

				if h, ok := s.signals[sig]; ok {
					h(sig)
				}
			}()
		}
	}
}

// Stop stops the signal router.
func (s *router) Stop(err error) {
	s.cancelFunc()
}

// Fire a signal.
func (s *router) Fire(sig os.Signal) {
	s.signalCh <- sig
}
