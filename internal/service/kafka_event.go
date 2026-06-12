package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/CodingFervor/live-commerce-bi/internal/cache"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/pkg/logger"
)

// ═══ Kafka Event Service ═══
// High-throughput event streaming for real-time data pipeline
// Topics: events.track, events.live_metrics, events.order, alerts.trigger

const (
	TopicEventsTrack     = "events.track"
	TopicLiveMetrics     = "events.live_metrics"
	TopicOrderEvent      = "events.order"
	TopicAlertTrigger    = "alerts.trigger"
	TopicDataSync        = "data.sync"
)

// KafkaConfig holds Kafka connection parameters
type KafkaConfig struct {
	Brokers []string
	GroupID string
}

// KafkaProducer wraps a Kafka writer for producing events
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers []string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			BatchSize:    100,
			RequiredAcks: kafka.RequireOne,
			Async:        true,
			Completion: func(messages []kafka.Message, err error) {
				if err != nil {
					logger.Error("Kafka batch write failed: %v", err)
				}
			},
		},
	}
}

// Close closes the producer
func (p *KafkaProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// PublishEvent publishes an event to a Kafka topic
func (p *KafkaProducer) PublishEvent(ctx context.Context, topic string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
		Headers: []kafka.Header{
			{Key: "source", Value: []byte("live-commerce-bi")},
			{Key: "timestamp", Value: []byte(time.Now().Format(time.RFC3339))},
		},
	})
}

// PublishBatch publishes multiple events in a single batch
func (p *KafkaProducer) PublishBatch(ctx context.Context, topic string, events []struct {
	Key   string
	Value interface{}
}) error {
	messages := make([]kafka.Message, 0, len(events))
	for _, e := range events {
		data, err := json.Marshal(e.Value)
		if err != nil {
			continue
		}
		messages = append(messages, kafka.Message{
			Topic: topic,
			Key:   []byte(e.Key),
			Value: data,
		})
	}
	if len(messages) == 0 {
		return nil
	}
	return p.writer.WriteMessages(ctx, messages...)
}

// KafkaConsumer wraps a Kafka reader for consuming events
type KafkaConsumer struct {
	reader  *kafka.Reader
	handler func(ctx context.Context, msg kafka.Message) error
	done    chan struct{}
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(brokers []string, groupID, topic string, handler func(ctx context.Context, msg kafka.Message) error) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:       brokers,
			GroupID:       groupID,
			Topic:         topic,
			MinBytes:      1,
			MaxBytes:      10 * 1024 * 1024, // 10MB
			QueueCapacity: 1000,
			CommitInterval: time.Second,
		}),
		handler: handler,
		done:    make(chan struct{}),
	}
}

// Start begins consuming messages
func (c *KafkaConsumer) Start(ctx context.Context) {
	go func() {
		defer close(c.done)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("[Kafka] Read error: %v", err)
				time.Sleep(time.Second)
				continue
			}

			if err := c.handler(ctx, msg); err != nil {
				log.Printf("[Kafka] Handler error (topic=%s, offset=%d): %v", msg.Topic, msg.Offset, err)
			}
		}
	}()
}

// Close closes the consumer
func (c *KafkaConsumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}

// Done returns a channel that closes when the consumer stops
func (c *KafkaConsumer) Done() <-chan struct{} {
	return c.done
}

// ═══ Event Handlers ═══

// TrackEventHandler processes track events from Kafka
func TrackEventHandler(ctx context.Context, msg kafka.Message) error {
	var event model.TrackEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal track event: %w", err)
	}

	// Push to Redis for real-time dashboard
	eventData, _ := json.Marshal(map[string]interface{}{
		"type":       "track_event",
		"event_name": event.EventName,
		"platform":   event.Platform,
		"live_room":  event.LiveRoomID,
		"timestamp":  event.Timestamp,
	})

	// Update real-time counter in Redis
	rdb := cache.Get()
	if rdb != nil {
		cacheKey := fmt.Sprintf("live:events:%d:count", event.LiveRoomID)
		rdb.IncrBy(ctx, cacheKey, 1)
		rdb.Expire(ctx, cacheKey, 24*time.Hour)

		// Push to WebSocket via Redis Pub/Sub
		PushRealtimeMetric(ctx, "live:metrics:update", eventData)
	}

	return nil
}

// LiveMetricsHandler processes live room metrics from Kafka
func LiveMetricsHandler(ctx context.Context, msg kafka.Message) error {
	var metrics model.LiveRoomMetrics
	if err := json.Unmarshal(msg.Value, &metrics); err != nil {
		return fmt.Errorf("unmarshal live metrics: %w", err)
	}

	// Cache latest metrics
	metricsData, _ := json.Marshal(metrics)
	cache.SetJSON(ctx, fmt.Sprintf("live:room:%d:metrics", metrics.RoomID), string(metricsData), 5*time.Minute)

	// Broadcast to WebSocket subscribers
	PushRealtimeMetric(ctx, "live:metrics:update", metricsData)

	return nil
}

// OrderEventHandler processes order events from Kafka
func OrderEventHandler(ctx context.Context, msg kafka.Message) error {
	var order model.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		return fmt.Errorf("unmarshal order event: %w", err)
	}

	// Update GMV counter in Redis
	rdb := cache.Get()
	if rdb != nil {
		incrByFloat(ctx, "live:gmv:today", order.ActualAmount)
		rdb.IncrBy(ctx, "live:orders:today", 1)

		// Per-room GMV
		incrByFloat(ctx, fmt.Sprintf("live:room:%d:gmv", order.LiveRoomID), order.ActualAmount)

		// Push real-time update
		orderData, _ := json.Marshal(map[string]interface{}{
			"type":        "new_order",
			"order_no":    order.OrderNo,
			"amount":      order.ActualAmount,
			"live_room":   order.LiveRoomID,
			"platform":    order.Platform,
		})
		PushRealtimeMetric(ctx, "live:room:gmv", orderData)
	}

	return nil
}

// KafkaEventService manages the full event pipeline
type KafkaEventService struct {
	producer  *KafkaProducer
	consumers []*KafkaConsumer
	brokers   []string
}

// NewKafkaEventService creates the full Kafka event pipeline
func NewKafkaEventService(brokers []string, groupID string) *KafkaEventService {
	return &KafkaEventService{
		producer: NewKafkaProducer(brokers),
		brokers:  brokers,
	}
}

// StartConsumers starts all topic consumers
func (s *KafkaEventService) StartConsumers(ctx context.Context) {
	topics := []struct {
		topic   string
		handler func(ctx context.Context, msg kafka.Message) error
	}{
		{TopicEventsTrack, TrackEventHandler},
		{TopicLiveMetrics, LiveMetricsHandler},
		{TopicOrderEvent, OrderEventHandler},
	}

	for _, t := range topics {
		consumer := NewKafkaConsumer(s.brokers, "live-bi-"+t.topic, t.topic, t.handler)
		consumer.Start(ctx)
		s.consumers = append(s.consumers, consumer)
	}

	logger.Info("Kafka consumers started for %d topics", len(topics))
}

// GetProducer returns the shared Kafka producer
func (s *KafkaEventService) GetProducer() *KafkaProducer {
	return s.producer
}

// Close shuts down all consumers and the producer
func (s *KafkaEventService) Close() error {
	for _, c := range s.consumers {
		c.Close()
	}
	return s.producer.Close()
}

// helper: Redis IncrByFloat (since our cache package doesn't expose it directly)
func incrByFloat(ctx context.Context, key string, value float64) {
	rdb := cache.Get()
	if rdb == nil {
		return
	}
	if err := rdb.IncrByFloat(ctx, key, value).Err(); err != nil {
		logger.Error("Redis IncrByFloat failed: %v", err)
	}
}
