package signal_test

import (
	"context"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gojini.dev/signal"
)

type sigH struct {
	s []os.Signal
	l *sync.RWMutex
}

func (s *sigH) handle(sig os.Signal) {
	s.l.Lock()
	s.s = append(s.s, sig)

	s.l.Unlock()
}

func (s *sigH) Len() int {
	s.l.RLock()
	defer s.l.RUnlock()

	return len(s.s)
}

func (s *sigH) Sig(i int) os.Signal {
	return s.s[i]
}

func TestSignals(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)
	ctx := context.Background()

	router := signal.New(ctx)

	assert.NotNil(router)
	assert.False(router.IsIgnored(syscall.SIGINT))
	assert.False(router.IsHandled(syscall.SIGINT))
	router.Handle(syscall.SIGINT, func(os.Signal) {})
	assert.True(router.IsHandled(syscall.SIGINT))
	router.Reset(syscall.SIGINT)
	assert.False(router.IsHandled(syscall.SIGINT))
	router.Ignore(syscall.SIGINT)
	assert.True(router.IsIgnored(syscall.SIGINT))
	assert.False(router.IsHandled(syscall.SIGINT))

	handler := &sigH{[]os.Signal{}, &sync.RWMutex{}}

	router.Handle(syscall.SIGINT, handler.handle)
	assert.True(router.IsHandled(syscall.SIGINT))
	assert.Equal(0, len(handler.s))

	go func() {
		if e := router.Start(); e != nil {
			panic(e)
		}
	}()

	defer router.Stop(nil)

	// Fire a signal
	router.Fire(syscall.SIGINT)

	// Sleep for a second
	time.Sleep(time.Second)
	assert.Equal(1, handler.Len())
	assert.Equal(syscall.SIGINT, handler.Sig(0))
}
