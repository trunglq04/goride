package messaging

import (
	"encoding/json"
	"log/slog"

	"github.com/trunglq04/goride/shared/contracts"
)

type QueueConsumer struct {
	rb        *RabbitMQ
	connMgr   *ConnectionManager
	queueName string
}

func NewQueueConsumer(rb *RabbitMQ, connMgr *ConnectionManager, queueName string) *QueueConsumer {
	return &QueueConsumer{
		rb:        rb,
		connMgr:   connMgr,
		queueName: queueName,
	}
}

func (qc *QueueConsumer) Start() error {
	msgs, err := qc.rb.Channel.Consume(
		qc.queueName, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			var msgBody contracts.AmqpMessage
			if err := json.Unmarshal(msg.Body, &msgBody); err != nil {
				slog.Error("Failed to unmarshal message",
					"queue", qc.queueName,
					"routing_key", msg.RoutingKey,
					"message_id", msg.MessageId,
					"err", err,
				)
				continue
			}

			userID := msgBody.OwnerID
			var payload any
			if msgBody.Data != nil {
				if err := json.Unmarshal(msgBody.Data, &payload); err != nil {
					slog.Error("Failed to unmarshal payload",
						"queue", qc.queueName,
						"routing_key", msg.RoutingKey,
						"user_id", userID,
						"err", err,
					)
					continue
				}
			}

			clientMsg := contracts.WSMessage{
				Type: msg.RoutingKey,
				Data: payload,
			}

			if err := qc.connMgr.SendMessage(userID, clientMsg); err != nil {
				if err == ErrConnectionNotFound {
					// The target WebSocket connection doesn't exist yet — requeue
					// so it can be retried once the client reconnects.
					slog.Warn("Connection not found for user, requeueing message",
						"queue", qc.queueName,
						"routing_key", msg.RoutingKey,
						"user_id", userID,
					)
				} else {
					slog.Error("Failed to send message to user, discarding",
						"queue", qc.queueName,
						"routing_key", msg.RoutingKey,
						"user_id", userID,
						"err", err,
					)
				}
				continue
			}

			slog.Debug("Delivered message to user over WebSocket",
				"queue", qc.queueName,
				"routing_key", msg.RoutingKey,
				"user_id", userID,
			)
		}
	}()

	slog.Info("Started queue consumer", "queue", qc.queueName)
	return nil
}
