package metrics

import (
	"errors"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/sirupsen/logrus"
)

var ErrorMetricsAlreadyRunning = errors.New("metrics server is already running")

// Pusher is a small wrapper around prometheus's Pusher to allow you to easily set up metrics to be pushed to a prometheus-pushgateway. It is thread safe
type Pusher struct {
	pusher *push.Pusher
	mu     sync.Mutex
	wg     sync.WaitGroup
	live   bool
	log    *logrus.Logger
}

func NewPusher(logger *logrus.Logger, task string) *Pusher {
	newPusher := &Pusher{
		mu:     sync.Mutex{},
		wg:     sync.WaitGroup{},
		live:   false,
		log:    logger,
		pusher: push.New("http://fleet-pushgateway:9090/", task),
	}

	return newPusher
}

// Add a new collector to the pusher
func (mp *Pusher) AddCollector(collector prometheus.Collector) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.pusher = mp.pusher.Collector(collector)
}

// Start the metrics pusher in a background thread
func (mp *Pusher) Start() error {
	if mp.IsAlive() {
		return ErrorMetricsAlreadyRunning
	}

	// First test that the pushgateway is accessible by running an initial push
	err := mp.Push()
	if err != nil {
		return err
	}

	mp.wg.Add(1)

	go func() {
		defer mp.wg.Done()

		mp.mu.Lock()
		mp.live = true
		mp.mu.Unlock()

		for mp.IsAlive() {
			time.Sleep(time.Second)

			err := mp.Push()
			if err != nil {
				mp.log.Errorf("Failed to push metrics to gateway: %s", err.Error())
				return
			}
		}
	}()

	return nil
}

// Used to check if the metrics pusher is still running
func (mp *Pusher) IsAlive() bool {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	return mp.live
}

// Kills the metrics pusher thread and waits for it to die
func (mp *Pusher) KillAndWait() {
	mp.mu.Lock()
	mp.live = false
	mp.mu.Unlock()

	mp.wg.Wait()
}

// Pushes metrics to the pushgateway
func (mp *Pusher) Push() error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	return mp.pusher.Push()
}
