package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/tracing"
)

const (
	TripExchange = "trip"
)

type RabbitMQ struct {
	conn    *amqp091.Connection
	Channel *amqp091.Channel
	mu      sync.Mutex
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

func (r *RabbitMQ) setupExchangesAndQueues() error {
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

	err = r.declareAndBindQueue(
		NotifyDriverNoDriversFoundQueue,
		[]string{contracts.TripEventNoDriversFound},
		TripExchange,
	)
	if err != nil {
		return err
	}

	err = r.declareAndBindQueue(
		NotifyDriverAssignQueue,
		[]string{contracts.TripEventDriverAssigned},
		TripExchange,
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
	)
	if err != nil {
		return err
	}

	err = r.declareAndBindQueue(
		NotifyPaymentSuccessQueue,
		[]string{contracts.PaymentEventSuccess},
		TripExchange,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageTypes []string, exchange string) error {
	q, err := r.Channel.QueueDeclare(
		queueName, // name
		true,      // durability
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		amqp091.Table{
			amqp091.QueueTypeArg: amqp091.QueueTypeQuorum,
		},
	)
	if err != nil {
		log.Fatal(err)
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

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}

	if r.Channel != nil {
		r.Channel.Close()
	}
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {
	log.Printf("Publishing message with routing key: %s", routingKey)

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

	return tracing.TracedPublisher(ctx, TripExchange, routingKey, msg, r.publish)
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

			if err := tracing.TracedConsumer(msg, func(ctx context.Context, d amqp091.Delivery) error {
				log.Printf("Received a message: %s", msg.Body)

				if err := handler(ctx, msg); err != nil {
					log.Printf("Failed to handle the message: %v", err)
					// Nack without requeue — discard poison message to avoid crash loop
					msg.Nack(false, false)

					return err
				}

				// Ack only on success
				if ackErr := msg.Ack(false); ackErr != nil {
					log.Printf("Failed to ack the message: %v. Message body: %s", ackErr, msg.Body)
					return ackErr
				}

				return nil
			}); err != nil {
				log.Printf("Failed to handle the message: %v", err)
			}

		}
	}()

	return nil
}
