package broker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/Lagwick/catalog-service/internal/pkg/broker/codec"
)

const (
	consumeBackoffInitial = 500 * time.Millisecond
	consumeBackoffMax     = 30 * time.Second
	consumeMaxRetries     = 5
)

var _ Bus[any] = (*kafkaBus[any])(nil)

type EventIdGetter interface {
	EventId() string
}

type kafkaBus[T any] struct {
	client        *KafkaClient
	codec         codec.Codec[T]
	topic         string
	consumerGroup string
	consumer      sarama.ConsumerGroup
}

func NewBus[T any](client *KafkaClient, c codec.Codec[T], topic, consumerGroup string) (Bus[T], error) {
	if client == nil {
		return nil, fmt.Errorf("broker: client is nil")
	}
	if topic == "" {
		return nil, fmt.Errorf("broker: topic is empty")
	}
	group := Coalesce(consumerGroup, client.DefaultConsumerGroup())
	if group == "" {
		return nil, fmt.Errorf("broker: consumer group is empty")
	}
	return &kafkaBus[T]{
		client:        client,
		codec:         c,
		topic:         topic,
		consumerGroup: group,
	}, nil
}

func MustKafkaBus[T any](client *KafkaClient, c codec.Codec[T], topic, consumerGroup string) Bus[T] {
	bus, err := NewBus(client, c, topic, consumerGroup)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("topic", topic).
			Msg("failed to create kafka bus")
	}
	return bus
}

func getEventId[T any](v *T) string {
	if getter, ok := any(v).(EventIdGetter); ok {
		if id := getter.EventId(); id != "" {
			return id
		}
	}

	return uuid.New().String()
}

func (b *kafkaBus[T]) Send(_ context.Context, msg *T, headers ...Header) error {
	data, err := b.codec.Encode(msg)
	if err != nil {
		return fmt.Errorf("broker: encode message: %w", err)
	}

	key := getEventId(msg)

	saramaMsg := &sarama.ProducerMessage{
		Topic: b.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(data),
	}

	if len(headers) > 0 {
		saramaMsg.Headers = make([]sarama.RecordHeader, 0, len(headers))

		for _, h := range headers {
			saramaMsg.Headers = append(saramaMsg.Headers, sarama.RecordHeader{
				Key:   []byte(h.Key),
				Value: []byte(h.Value),
			})
		}
	}

	if _, _, err := b.client.Producer().SendMessage(saramaMsg); err != nil {
		return fmt.Errorf("broker: send message: %w", err)
	}

	return nil
}

func (b *kafkaBus[T]) QueueName() string {
	return b.topic
}

func (b *kafkaBus[T]) Subscribe(
	ctx context.Context,
	wg *sync.WaitGroup,
	handler MessageHandler[T],
) error {
	consumer, err := b.client.NewConsumerGroup(b.consumerGroup)
	if err != nil {
		return fmt.Errorf("broker: create consumer group %q: %w", b.consumerGroup, err)
	}

	b.consumer = consumer

	h := &consumerGroupHandler[T]{
		codec:   b.codec,
		handler: handler,
		topic:   b.topic,
	}

	if wg != nil {
		wg.Add(1)
	}

	go func() {
		for err := range consumer.Errors() {
			log.Warn().
				Err(err).
				Str("topic", b.topic).
				Msg("consumer group error")
		}
	}()

	go func() {
		if wg != nil {
			defer wg.Done()
		}

		defer func() {
			if err := consumer.Close(); err != nil && !errors.Is(err, sarama.ErrClosedClient) {
				log.Error().
					Err(err).
					Str("topic", b.topic).
					Msg("failed to close consumer group")
			}
		}()

		backoff := consumeBackoffInitial
		fails := 0

		for {
			err := consumer.Consume(ctx, []string{b.topic}, h)

			if ctx.Err() != nil {
				return
			}

			if err == nil {
				backoff = consumeBackoffInitial
				fails = 0
				continue
			}

			fails++

			if fails >= consumeMaxRetries {
				log.Fatal().
					Err(err).
					Str("topic", b.topic).
					Int("fails", fails).
					Msg("consumer group failed too many times")
			}

			log.Warn().
				Err(err).
				Str("topic", b.topic).
				Int("fails", fails).
				Dur("backoff", backoff).
				Msg("consumer group consume failed, retrying")

			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}

			backoff = min(backoff*2, consumeBackoffMax)
		}
	}()

	return nil
}

type consumerGroupHandler[T any] struct {
	codec   codec.Codec[T]
	handler MessageHandler[T]
	topic   string
}

func (h *consumerGroupHandler[T]) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler[T]) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler[T]) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	ctx := session.Context()
	defer session.Commit()

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			if err := h.process(ctx, session, msg); err != nil {
				return err
			}
		}
	}
}

func (h *consumerGroupHandler[T]) process(
	ctx context.Context,
	session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage,
) error {
	decoded, err := h.codec.Decode(msg.Value)
	if err != nil {
		log.Error().
			Err(err).
			Str("topic", h.topic).
			Msg("failed to decode message")

		session.MarkMessage(msg, "")

		return nil
	}

	headers := make([]Header, 0, len(msg.Headers))
	for _, header := range msg.Headers {
		headers = append(headers, Header{
			Key:   string(header.Key),
			Value: string(header.Value),
		})
	}

	err = h.handler(ctx, decoded, headers)
	if err != nil {
		if IsNotCriticalError(err) {
			log.Warn().
				Err(err).
				Str("topic", h.topic).
				Msg("non-critical message processing error")

			session.MarkMessage(msg, "")

			return nil
		}

		log.Error().
			Err(err).
			Str("topic", h.topic).
			Msg("critical message processing error")

		return err
	}

	session.MarkMessage(msg, "")

	return nil
}

func (b *kafkaBus[T]) Close() error {
	if b.consumer != nil {
		return b.consumer.Close()
	}

	return nil
}
