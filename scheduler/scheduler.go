package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type job struct {
	cmd      func()
	name     string
	schedule schedule
}

type Scheduler struct {
	cv       *sync.Cond
	jobs     []job
	logger   *log.Logger
	mutex    *sync.Mutex
	numReady int
	ready    sync.Map
}

func New(logger *log.Logger) *Scheduler {
	mutex := &sync.Mutex{}
	cv := sync.NewCond(mutex)

	return &Scheduler{
		cv:       cv,
		jobs:     nil,
		logger:   logger,
		mutex:    mutex,
		numReady: 0,
		ready:    sync.Map{},
	}
}

func (s *Scheduler) Add(schedule string, cmd func()) (bool, error) {
	return s.add(schedule, cmd, fmt.Sprintf("job-%d", time.Now().UTC().Unix()))
}

func (s *Scheduler) AddWithName(schedule string, cmd func(), name string) (bool, error) {
	return s.add(schedule, cmd, name)
}

func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Printf("Starting CRON scheduling with [%d] jobs.", len(s.jobs))
	go func() {
		for {
			select {
			case <-ctx.Done():
			default:
				s.mutex.Lock()
				for s.numReady == 0 {
					s.cv.Wait()
				}

				s.ready.Range(func(key, _ any) bool {
					s.ready.Delete(key)
					s.numReady -= 1

					go func() {
						job := s.jobs[key.(int)]

						duration := getDurationTillNextProc(job.schedule)
						s.logger.Printf("Scheduling job [%s] | Time till next proc [%s]", job.name, duration)

						time.Sleep(duration)
						job.cmd()

						s.mutex.Lock()
						defer s.mutex.Unlock()
						s.ready.Store(key, true)
						s.numReady += 1
						s.cv.Signal()
					}()

					return true
				})
				s.mutex.Unlock()
			}
		}
	}()
}

func (s *Scheduler) add(schedule string, cmd func(), name string) (bool, error) {
	ok, err := validateSchedule(schedule)
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	job := job{
		cmd:      cmd,
		schedule: parseSchedule(schedule),
		name:     name,
	}

	s.mutex.Lock()
	s.ready.Store(s.numReady, true)
	s.jobs = append(s.jobs, job)
	s.numReady += 1
	s.mutex.Unlock()

	return true, nil
}
