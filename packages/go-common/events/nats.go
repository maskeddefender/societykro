package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/societykro/go-common/logger"
)

// Bus wraps a NATS connection and provides typed publish/subscribe helpers.
type Bus struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

// Event is the envelope for all events on the bus.
type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// NewBus connects to NATS and enables JetStream.
func NewBus(url string) (*Bus, error) {
	conn, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			logger.Log.Warn().Err(err).Msg("NATS disconnected")
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			logger.Log.Info().Msg("NATS reconnected")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("jetstream init: %w", err)
	}

	logger.Log.Info().Str("url", url).Msg("NATS connected (JetStream)")
	return &Bus{conn: conn, js: js}, nil
}

// EnsureStream creates a stream if it doesn't exist. Call once at startup.
func (b *Bus) EnsureStream(name string, subjects []string) error {
	_, err := b.js.StreamInfo(name)
	if err == nil {
		return nil // already exists
	}

	_, err = b.js.AddStream(&nats.StreamConfig{
		Name:       name,
		Subjects:   subjects,
		Retention:  nats.WorkQueuePolicy,
		MaxAge:     72 * time.Hour,
		Storage:    nats.FileStorage,
		Duplicates: 5 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("create stream %s: %w", name, err)
	}
	logger.Log.Info().Str("stream", name).Strs("subjects", subjects).Msg("NATS stream created")
	return nil
}

// Publish sends a typed event to the given subject.
func (b *Bus) Publish(subject string, eventType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	evt := Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Payload:   data,
	}

	evtBytes, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = b.js.Publish(subject, evtBytes)
	if err != nil {
		return fmt.Errorf("publish to %s: %w", subject, err)
	}

	logger.Log.Debug().Str("subject", subject).Str("type", eventType).Msg("Event published")
	return nil
}

// Subscribe creates a durable pull subscription on the given subject.
// handler receives the raw Event; return nil to ACK, error to NAK (retry).
func (b *Bus) Subscribe(subject, durable string, handler func(Event) error) error {
	sub, err := b.js.PullSubscribe(subject, durable)
	if err != nil {
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}

	go func() {
		for {
			msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				}
				logger.Log.Error().Err(err).Str("subject", subject).Msg("NATS fetch error")
				time.Sleep(1 * time.Second)
				continue
			}
			for _, msg := range msgs {
				var evt Event
				if err := json.Unmarshal(msg.Data, &evt); err != nil {
					logger.Log.Error().Err(err).Msg("Unmarshal event failed")
					msg.Nak()
					continue
				}

				if err := handler(evt); err != nil {
					logger.Log.Error().Err(err).Str("type", evt.Type).Msg("Handler failed, NAK")
					msg.Nak()
				} else {
					msg.Ack()
				}
			}
		}
	}()

	logger.Log.Info().Str("subject", subject).Str("durable", durable).Msg("NATS subscription started")
	return nil
}

// Close gracefully drains and closes the NATS connection.
func (b *Bus) Close() {
	b.conn.Drain()
}
