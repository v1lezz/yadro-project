package services

import (
	"errors"
	"golang.org/x/time/rate"
)

var ( //errors
	ErrManyRequests = errors.New("many requests")
)

type Semaphore struct {
	ch chan struct{}
}

func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.ch
}

func NewSemaphore(limit int) Semaphore {
	return Semaphore{
		ch: make(chan struct{}, limit),
	}
}

type LimitService struct {
	rl        map[string]*rate.Limiter
	rateLimit int
	cl        Semaphore
}

func NewLimitService(rateLimit, concurrencyLimit int) *LimitService {
	return &LimitService{
		rl: make(map[string]*rate.Limiter),
		rateLimit:
		cl: NewSemaphore(concurrencyLimit),
	}
}

func (svc *LimitService) Limit(email string) error {
	if _, ok := svc.rl[email]; !ok {
		svc.rl[email] = rate.NewLimiter(rate.Limit(svc.rateLimit), 10)
	}

	select {
	case svc.rl[email].:
		return ErrManyRequests
	default:
	}

	return nil
}
