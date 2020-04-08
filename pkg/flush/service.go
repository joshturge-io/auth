package flush

import (
	"context"
	"log"
	"sync"
	"time"
)

type Flusher interface {
	Flush() error
}

type Service struct {
	flush    Flusher
	flushInt time.Duration
	lg       *log.Logger
	wg       sync.WaitGroup
	quit     chan struct{}
	err      error
}

func NewService(lg *log.Logger, flush Flusher, flushInt time.Duration) *Service {
	return &Service{flush, flushInt, lg, sync.WaitGroup{}, make(chan struct{}), nil}
}

func (s *Service) Start() {
	ticker := time.NewTicker(s.flushInt)
	s.wg.Add(1)
	go func() {
		defer func() {
			close(s.quit)
			s.wg.Done()
		}()
		for {
			select {
			case <-ticker.C:
				s.lg.Println("Flushing blacklist...")
				if err := s.flush.Flush(); err != nil {
					s.lg.Printf("ERROR: failed to flush: %s", err.Error())
					s.err = err
				}
			case <-s.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Service) Err() error {
	return s.err
}

func (s *Service) Close(ctx context.Context) error {
	done := make(chan struct{})
	defer close(done)

	go func() {
		s.quit <- struct{}{}
		s.wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	return nil
}
