package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/metrics"
	"github.com/trunglq04/goride/shared/retry"
	"github.com/trunglq04/goride/shared/tracing"
)

const (
	TripExchange       = "trip"
	DeadLetterExchange = "dlx"
)

type RabbitMQ struct {
	conn    *amqp091.Connection
	Channel *amqp091.Channel
	mu      sync.Mutex
}

// QueueOptions holds optional configuration for queue declaration.
type QueueOptions struct {
	// Time to live of the message in milliseconds (x-message-ttl).
	TTL int64
}

// QueueOption is a functional option for declareAndBindQueue.
type QueueOption func(*QueueOptions)

// WithTTL sets the message TTL (in milliseconds) on the queue.
func WithTTL(ms int64) QueueOption {
	return func(o *QueueOptions) {
		o.TTL = ms
	}
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		err := conn.Close()
		return nil, fmt.Errorf("failed to create  RabbitMQ channel : %v", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}

	// Setup the exchange and queues
	if err := rmq.setupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to setup exchanges and queues: %v", err)
	}

	return rmq, nil
}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}

	if r.Channel != nil {
		r.Channel.Close()
	}
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {
	slog.DebugContext(ctx, "Publishing message",
		"exchange", TripExchange,
		"routing_key", routingKey,
		"owner_id", message.OwnerID,
	)

	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	msg := amqp091.Publishing{
		DeliveryMode: amqp091.Persistent,
		ContentType:  "application/json",
		Body:         jsonMsg,
	}

	err = tracing.TracedPublisher(ctx, TripExchange, routingKey, msg, r.publish)
	if err == nil {
		metrics.RecordPublish(routingKey)
	}
	return err
}

type MessageHandler func(context.Context, amqp091.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	// QoS (Quality of Service): Set prefetch count to 1 for fair dispatch
	// This tells RabbitMQ not to give more than one msg to a service at a time
	// The worker will only get the next message after it has acknowledged the previous one.
	err := r.Channel.Qos(
		1,     // prefetchCount: Limit to 1 unacknowledged msg per consumer
		0,     // prefetchSize: No specific limit on msg size
		false, // global: Apply prefetchCount to each consumer individually
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}

	msgs, err := r.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume a message: %v", err)
	}

	go func() {
		for msg := range msgs {
			if err := tracing.TracedConsumer(msg, func(ctx context.Context, delivery amqp091.Delivery) error {
				slog.DebugContext(ctx, "Received a message",
					"queue", queueName,
					"routing_key", delivery.RoutingKey,
					"message_id", delivery.MessageId,
					"body_size", len(delivery.Body),
				)

				cfg := retry.DefaultConfig()
				err := retry.WithBackoff(ctx, cfg, func() error {
					return handler(ctx, delivery)
				})
				if err != nil {
					slog.ErrorContext(ctx, "Message processing failed after retries, sending to DLQ",
						"queue", queueName,
						"routing_key", delivery.RoutingKey,
						"message_id", delivery.MessageId,
						"max_retries", cfg.MaxRetries,
						"err", err,
					)

					// Add failure context before sending to the DLQ
					header := amqp091.Table{}
					if delivery.Headers != nil {
						header = delivery.Headers
					}

					header["x-death-reason"] = err.Error()
					header["x-original-exchange"] = delivery.Exchange
					header["x-original-routing-key"] = delivery.RoutingKey
					header["x-retry-count"] = cfg.MaxRetries
					delivery.Headers = header

					// Reject without requeue - message will go to the DLQ
					_ = delivery.Nack(false, false)
					metrics.RecordConsume(queueName, "error")
					return err
				}

				// Ack only on success
				if ackErr := msg.Ack(false); ackErr != nil {
					slog.ErrorContext(ctx, "Failed to ack the message",
						"queue", queueName,
						"message_id", delivery.MessageId,
						"err", ackErr,
					)
					return ackErr
				}

				metrics.RecordConsume(queueName, "ok")
				return nil
			}); err != nil {
				slog.Error("Error processing the message", "queue", queueName, "err", err)
			}

		}
	}()

	return nil
}

func (r *RabbitMQ) publish(ctx context.Context, exchange, routingKey string, msg amqp091.Publishing) error {
	return r.Channel.PublishWithContext(ctx,
		TripExchange, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		msg,          // amqp091.Publishing
	)
}

func (r *RabbitMQ) setupDeadLetterExchange() error {
	// Declare the dead letter exchange
	err := r.Channel.ExchangeDeclare(
		DeadLetterExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter exchange: %v", err)
	}

	// Declare the dead letter queue
	q, err := r.Channel.QueueDeclare(
		DeadLetterQueue, // name
		true,            // durability
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments with DLX config
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter queue: %v", err)
	}

	// Bind the queue to the dead letter exchange with a wildcard routing key
	err = r.Channel.QueueBind(
		q.Name,             // queue name
		"#",                // wildcard routing key to catch all messages
		DeadLetterExchange, // exchange
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind dead letter queue to dead letter exchange: %v", err)
	}

	return nil
}

func (r *RabbitMQ) setupExchangesAndQueues() error {
	// First setup the DLQ exchange and queue
	if err := r.setupDeadLetterExchange(); err != nil {
		return err
	}

	err := r.Channel.ExchangeDeclare(
		TripExchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange %s: %v", TripExchange, err)
	}

	err = r.declareAndBindQueue(
		FindAvailableDriversQueue, // queue name
		[]string{ // routing keys
			contracts.TripEventCreated,
			contracts.TripEventDriverNotInterested,
		}, // message types
		TripExchange, // exchange
	)
	if err != nil {
		return err
	}

	err = r.declareAndBindQueue(
		DriverCmdTripRequestQueue,
		[]string{contracts.DriverCmdTripRequest},
		TripExchange,
	)
	if err != nil {
		return err
	}

	err = r.declareAndBindQueue(
		DriverTripResponseQueue,
		[]string{
			contracts.DriverCmdTripAccept,
			contracts.DriverCmdTripDecline,
		},
		TripExchange,
	)
	if err != nil {
		return err
	}

	// Notification queues: 5-minute TTL so stale notifications are
	// dead-lettered rather than delivered to a client that has moved on.
	const notifyTTL = 24 * 60 * 60 * 1000 // 300 000 ms

	err = r.declareAndBindQueue(
		NotifyDriverNoDriversFoundQueue,
		[]string{contracts.TripEventNoDriversFound},
		TripExchange,
		WithTTL(notifyTTL),
	)
	if err != nil {
		return err
	}

	err = r.declareAndBindQueue(
		NotifyDriverAssignQueue,
		[]string{contracts.TripEventDriverAssigned},
		TripExchange,
		WithTTL(notifyTTL),
	)
	if err != nil {
		return err
	}

	err = r.declareAndBindQueue(
		NotifyDriverLocationQueue,
		[]string{contracts.DriverCmdLocation},
		TripExchange,
		WithTTL(notifyTTL),
	)
	if err != nil {
		return err
	}

	err = r.declareAndBindQueue(
		PaymentTripResponseQueue,
		[]string{contracts.PaymentCmdCreateSession},
		TripExchange,
	)
	if err != nil {
		return err
	}

	err = r.declareAndBindQueue(
		NotifyPaymentSessionCreatedQueue,
		[]string{contracts.PaymentEventSessionCreated},
		TripExchange,
		WithTTL(notifyTTL),
	)
	if err != nil {
		return err
	}

	err = r.declareAndBindQueue(
		NotifyPaymentSuccessQueue,
		[]string{contracts.PaymentEventSuccess},
		TripExchange,
		WithTTL(notifyTTL),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageTypes []string, exchange string, opts ...QueueOption) error {
	// Apply functional options
	qOpts := &QueueOptions{}
	for _, opt := range opts {
		opt(qOpts)
	}

	// All queues use the dead-letter exchange
	args := amqp091.Table{}
	args["x-dead-letter-exchange"] = DeadLetterExchange

	// Optionally set a per-queue message TTL
	if qOpts.TTL > 0 {
		args["x-message-ttl"] = qOpts.TTL
	}

	q, err := r.Channel.QueueDeclare(
		queueName, // name
		true,      // durability
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments with DLX config
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %v", queueName, err)
	}

	for _, msg := range messageTypes {
		if err := r.Channel.QueueBind(
			q.Name,   // queue name
			msg,      // routing key
			exchange, // exchange
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue to %s: %v", queueName, err)
		}
	}

	return nil
}
