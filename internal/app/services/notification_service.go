package services

import (
	"context"
	"igo-repo/internal/logging"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type notificationMessage string

const (
	NotifMsgIconCreated notificationMessage = "iconCreated"
)

// subscriber represents a subscriber.
// Messages are sent on the msgs channel and if the client
// cannot keep up with the messages, closeSlow is called.
type subscriber struct {
	msgs      chan string
	closeSlow func()
}

type Notification struct {
	// subscriberMessageBuffer controls the max number
	// of messages that can be queued for a subscriber
	// before it is kicked.
	//
	// Defaults to 16.
	subscriberMessageBuffer int

	// publishLimiter controls the rate limit applied to the publish endpoint.
	//
	// Defaults to one publish every 100ms with a burst of 8.
	publishLimiter *rate.Limiter

	subscribersMu sync.Mutex
	subscribers   map[*subscriber]struct{}

	logger zerolog.Logger
}

func CreateNotificationService(log zerolog.Logger) *Notification {
	ns := &Notification{
		subscriberMessageBuffer: 16,
		subscribers:             make(map[*subscriber]struct{}),
		publishLimiter:          rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
		logger:                  logging.CreateUnitLogger(log, "notification-server"),
	}

	return ns
}

type socketIO interface {
	Close() error
	Write(ctx context.Context, msg string) error
}

func (ns *Notification) Subscribe(ctx context.Context, sIo socketIO) error {
	subs := &subscriber{
		msgs: make(chan string, ns.subscriberMessageBuffer),
		closeSlow: func() {
			sIo.Close()
		},
	}

	ns.addSubscriber(subs)
	defer ns.deleteSubscriber(subs)

	for {
		select {
		case msg := <-subs.msgs:
			err := writeTimeout(ctx, time.Second*5, sIo, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// addSubscriber registers a subscriber.
func (ns *Notification) addSubscriber(s *subscriber) {
	ns.subscribersMu.Lock()
	ns.subscribers[s] = struct{}{}
	ns.subscribersMu.Unlock()
}

// deleteSubscriber deletes the given subscriber.
func (ns *Notification) deleteSubscriber(s *subscriber) {
	ns.subscribersMu.Lock()
	delete(ns.subscribers, s)
	ns.subscribersMu.Unlock()
}

// publish publishes the msg to all subscribers.
// It never blocks and so messages to slow subscribers
// are dropped.
func (cs *Notification) Publish(msg notificationMessage) {
	cs.subscribersMu.Lock()
	defer cs.subscribersMu.Unlock()

	cs.publishLimiter.Wait(context.Background())

	for s := range cs.subscribers {
		select {
		case s.msgs <- string(msg):
		default:
			go s.closeSlow()
		}
	}

}

func writeTimeout(ctx context.Context, timeout time.Duration, sIo socketIO, msg string) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return sIo.Write(ctx, msg)
}
